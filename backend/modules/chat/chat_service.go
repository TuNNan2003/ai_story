package chat

import (
	"grandma/backend/models"
	"grandma/backend/modules/rag"
	"grandma/backend/repository"
	"grandma/backend/services"
	"grandma/backend/utils"
	"io"
	"strings"
)

type ChatService struct {
	conversationRepo *repository.ConversationRepository
	documentRepo     *repository.DocumentRepository
	workDocumentRepo *repository.WorkDocumentRepository
	ragService       *rag.RAGService
	config           *ChatConfig
}

type ChatConfig struct {
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	AnthropicAPIKey  string
	AnthropicBaseURL string
}

func NewChatService(conversationRepo *repository.ConversationRepository, documentRepo *repository.DocumentRepository, workDocumentRepo *repository.WorkDocumentRepository, ragService *rag.RAGService, config *ChatConfig) *ChatService {
	return &ChatService{
		conversationRepo: conversationRepo,
		documentRepo:     documentRepo,
		workDocumentRepo: workDocumentRepo,
		ragService:       ragService,
		config:           config,
	}
}

// SendMessage 发送消息并获取流式响应
func (s *ChatService) SendMessage(req *models.ChatRequest, writer io.Writer) (string, string, error) {
	// v1.3: 支持灵感模式（work_id）和普通模式（conversation_id）
	if req.WorkID != "" {
		return s.sendMessageForWork(req, writer)
	}
	return s.sendMessageForConversation(req, writer)
}

// sendMessageForWork 灵感模式：保存到WorkDocument
func (s *ChatService) sendMessageForWork(req *models.ChatRequest, writer io.Writer) (string, string, error) {
	workID := req.WorkID
	var err error

	// 构建API调用的消息数组
	var apiMessages []models.Message

	// 获取用户当前消息内容（用于RAG检索）
	var userQuery string
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			userQuery = lastMsg.Content
		}
	}

	// 灵感模式：添加专门的系统提示
	inspirationSystemPrompt := models.Message{
		Role: "system",
		Content: `你是一位专业的长篇故事创作助手。当前处于"灵感模式"，这是专门用于长篇故事创作的协作模式。

## 创作模式说明
- 你正在协助用户创作一部完整的长篇小说
- 每次创作的内容都是故事的一部分，需要与整体保持高度一致
- 用户可能会提出修改、补充、扩展等要求，你需要灵活应对

## 创作原则
1. **一致性优先**：严格保持人物、世界观、情节的一致性
2. **自然衔接**：新内容必须与已有内容自然衔接，不能突兀
3. **细节呼应**：注意伏笔、线索、细节的呼应和连贯
4. **风格统一**：保持整体文风和叙事风格的一致性

请根据用户的要求和提供的背景信息，创作高质量的故事内容。`,
	}
	apiMessages = append(apiMessages, inspirationSystemPrompt)

	// 使用RAG检索相关上下文（如果启用）
	if s.ragService != nil && userQuery != "" {
		ragContext, ragErr := s.ragService.BuildRAGContext(userQuery, req.UserID, "", workID)
		if ragErr == nil && len(ragContext) > 0 {
			// 将RAG检索到的上下文添加到系统提示之后
			apiMessages = append(apiMessages, ragContext...)
		}
	}

	// 从WorkDocument加载历史消息（只获取最近的5条，因为RAG已经提供了相关背景）
	historyDocs, err := s.workDocumentRepo.GetLatestDocumentsByWorkID(workID, 5)
	if err == nil && len(historyDocs) > 0 {
		// 反转顺序，使其按时间正序排列（最新的在最后）
		for i, j := 0, len(historyDocs)-1; i < j; i, j = i+1, j-1 {
			historyDocs[i], historyDocs[j] = historyDocs[j], historyDocs[i]
		}
		// 将历史文档转换为消息格式
		for _, doc := range historyDocs {
			apiMessages = append(apiMessages, models.Message{
				Role:    doc.Role,
				Content: doc.Content,
			})
		}
	}

	// 添加当前用户消息
	for _, msg := range req.Messages {
		apiMessages = append(apiMessages, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 保存最后一条用户消息（必须存在）
	var userDocID string
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			userDocID = utils.GenerateDocumentID()
			userDoc := &models.WorkDocument{
				ID:      userDocID,
				WorkID:  workID,
				UserID:  req.UserID,
				Title:   "", // 用户消息不需要标题
				Content: lastMsg.Content,
				Role:    "user",
				Model:   req.Model,
			}
			err = s.workDocumentRepo.Create(userDoc)
			if err != nil {
				return "", "", err
			}
			// 索引用户消息（异步）
			if s.ragService != nil {
				s.ragService.IndexDocument(userDocID, req.UserID, "", workID, lastMsg.Content, "user")
			}
		}
	}

	// 调用大模型API获取流式响应
	provider, err := services.GetProvider(
		req.Model,
		s.config.OpenAIAPIKey,
		s.config.OpenAIBaseURL,
		s.config.AnthropicAPIKey,
		s.config.AnthropicBaseURL,
	)
	if err != nil {
		return "", "", err
	}

	// 首先创建助手文档（空内容）
	assistantDocID := utils.GenerateDocumentID()
	assistantDoc := &models.WorkDocument{
		ID:      assistantDocID,
		WorkID:  workID,
		UserID:  req.UserID,
		Title:   "", // AI响应不需要标题
		Content: "",
		Role:    "assistant",
		Model:   req.Model,
	}
	err = s.workDocumentRepo.Create(assistantDoc)
	if err != nil {
		return "", "", err
	}

	// 创建流式响应收集器，在流式返回时逐步更新文档
	responseCollector := &workResponseCollector{
		writer:           writer,
		content:          "",
		workDocumentRepo: s.workDocumentRepo,
		documentID:       assistantDocID,
		updateBuffer:     "",
		bufferSize:       0,
		ragService:       s.ragService,
		userID:           req.UserID,
		conversationID:   "",
		workID:           workID,
		role:             "assistant",
		indexed:          false,
	}
	err = provider.ChatStream(apiMessages, responseCollector)

	// 无论流式响应是否成功，都要保存剩余的缓冲区内容
	if responseCollector.updateBuffer != "" {
		appendErr := s.workDocumentRepo.AppendContent(assistantDocID, responseCollector.updateBuffer)
		if appendErr != nil {
			if err == nil {
				err = appendErr
			}
		}
	}

	// 流式响应结束后，触发最终索引（如果还没有索引过）
	if s.ragService != nil && !responseCollector.indexed && len(responseCollector.content) > 0 {
		doc, docErr := s.workDocumentRepo.GetByIDAndUserID(assistantDocID, req.UserID)
		if docErr == nil && doc != nil {
			s.ragService.IndexDocument(doc.ID, doc.UserID, "", doc.WorkID, doc.Content, doc.Role)
		}
	}

	// 返回workID和文档ID（为了兼容前端，返回workID作为conversationID）
	return workID, assistantDocID, nil
}

// sendMessageForConversation 普通模式：保存到Document
func (s *ChatService) sendMessageForConversation(req *models.ChatRequest, writer io.Writer) (string, string, error) {
	var conversationID string
	var err error

	// 如果没有提供对话ID，创建新对话
	if req.ConversationID == "" {
		conversationID = utils.GenerateConversationID()
		conversation := &models.Conversation{
			ID:          conversationID,
			UserID:      req.UserID,
			Title:       s.generateTitle(req.Messages),
			DocumentIDs: "",
		}
		err = s.conversationRepo.Create(conversation)
		if err != nil {
			return "", "", err
		}
	} else {
		conversationID = req.ConversationID
		// 验证对话属于该用户
		_, err = s.conversationRepo.GetByIDAndUserID(conversationID, req.UserID)
		if err != nil {
			return "", "", err
		}
	}

	// 构建API调用的消息数组
	var apiMessages []models.Message

	// 获取用户当前消息内容（用于RAG检索）
	var userQuery string
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			userQuery = lastMsg.Content
		}
	}

	// 使用RAG检索相关上下文（如果启用）
	if s.ragService != nil && userQuery != "" {
		ragContext, err := s.ragService.BuildRAGContext(userQuery, req.UserID, conversationID, "")
		if err == nil && len(ragContext) > 0 {
			// 将RAG检索到的上下文添加到消息数组开头
			apiMessages = append(apiMessages, ragContext...)
		}
	}

	// 如果提供了对话ID，从数据库加载历史消息
	// 优化：只加载最近的对话上下文，而不是全部历史
	// 保留最近3-5条消息作为直接上下文（保证对话连贯性）
	if req.ConversationID != "" {
		// 获取对话的历史文档（只获取最近的5条，因为RAG已经提供了相关背景）
		historyDocs, err := s.documentRepo.GetLatestDocumentsByConversationID(conversationID, 5)
		if err == nil && len(historyDocs) > 0 {
			// 反转顺序，使其按时间正序排列（最新的在最后）
			for i, j := 0, len(historyDocs)-1; i < j; i, j = i+1, j-1 {
				historyDocs[i], historyDocs[j] = historyDocs[j], historyDocs[i]
			}
			// 将历史文档转换为消息格式
			for _, doc := range historyDocs {
				apiMessages = append(apiMessages, models.Message{
					Role:    doc.Role,
					Content: doc.Content,
				})
			}
		}
	}

	// 添加当前用户消息
	for _, msg := range req.Messages {
		apiMessages = append(apiMessages, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 保存最后一条用户消息（必须存在）
	var userDocID string
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			userDocID = utils.GenerateDocumentID()
			userDoc := &models.Document{
				ID:             userDocID,
				UserID:         req.UserID,
				ConversationID: conversationID,
				Role:           "user",
				Content:        lastMsg.Content,
				Model:          req.Model,
			}
			err = s.documentRepo.Create(userDoc)
			if err != nil {
				return "", "", err
			}
			// 添加用户文档ID到对话的文档ID列表
			err = s.conversationRepo.AppendDocumentID(conversationID, userDocID)
			if err != nil {
				return "", "", err
			}
			// 索引用户消息（异步）
			if s.ragService != nil {
				s.ragService.IndexDocument(userDocID, req.UserID, conversationID, "", lastMsg.Content, "user")
			}
		}
	}

	// 调用大模型API获取流式响应
	provider, err := services.GetProvider(
		req.Model,
		s.config.OpenAIAPIKey,
		s.config.OpenAIBaseURL,
		s.config.AnthropicAPIKey,
		s.config.AnthropicBaseURL,
	)
	if err != nil {
		return "", "", err
	}

	// 首先创建助手文档（空内容）
	assistantDocID := utils.GenerateDocumentID()
	assistantDoc := &models.Document{
		ID:             assistantDocID,
		UserID:         req.UserID,
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        "",
		Model:          req.Model,
	}
	err = s.documentRepo.Create(assistantDoc)
	if err != nil {
		return "", "", err
	}

	// 创建流式响应收集器，在流式返回时逐步更新文档
	responseCollector := &responseCollector{
		writer:         writer,
		content:        "",
		documentRepo:   s.documentRepo,
		documentID:     assistantDocID,
		updateBuffer:   "",
		bufferSize:     0,
		ragService:     s.ragService,
		userID:         req.UserID,
		conversationID: conversationID,
		workID:         "",
		role:           "assistant",
	}
	err = provider.ChatStream(apiMessages, responseCollector)

	// 无论流式响应是否成功，都要保存剩余的缓冲区内容
	// 这样即使客户端断开连接，已接收的内容也会被保存
	if responseCollector.updateBuffer != "" {
		appendErr := s.documentRepo.AppendContent(assistantDocID, responseCollector.updateBuffer)
		if appendErr != nil {
			// 如果保存失败，记录错误但不影响主流程
			// 因为如果流式响应成功，后续会继续保存
			if err == nil {
				err = appendErr
			}
		}
	}

	// 流式响应结束后，触发最终索引（如果还没有索引过）
	if s.ragService != nil && !responseCollector.indexed && len(responseCollector.content) > 0 {
		doc, docErr := s.documentRepo.GetByID(assistantDocID)
		if docErr == nil && doc != nil {
			s.ragService.IndexDocument(doc.ID, doc.UserID, doc.ConversationID, "", doc.Content, doc.Role)
		}
	}

	// 如果流式响应过程中出现错误（可能是客户端断开连接），
	// 仍然保存已接收的内容，并继续添加文档ID到对话列表
	// 这样用户可以切换回对话时看到部分内容
	// 注意：即使流式响应失败，也要继续处理，确保已保存的内容可以被访问
	_ = err // 忽略流式响应错误，继续保存已接收的内容

	// 添加助手文档ID到对话的文档ID列表（即使流式响应失败也要添加）
	errAppend := s.conversationRepo.AppendDocumentID(conversationID, assistantDocID)
	if errAppend != nil {
		// 如果添加文档ID失败，记录错误
		// 但继续执行，确保已保存的内容可以被访问
		_ = errAppend
	}

	// 无论流式响应是否成功，都返回成功
	// 这样即使客户端断开连接，已接收的内容也会被保存并可以被访问
	return conversationID, assistantDocID, nil
}

// generateTitle 从消息内容生成对话标题
func (s *ChatService) generateTitle(messages []models.Message) string {
	if len(messages) == 0 {
		return "新对话"
	}
	// 使用第一条用户消息作为标题
	for _, msg := range messages {
		if msg.Role == "user" {
			title := strings.TrimSpace(msg.Content)
			if len(title) > 50 {
				title = title[:50] + "..."
			}
			if title == "" {
				title = "新对话"
			}
			return title
		}
	}
	return "新对话"
}

// responseCollector 收集流式响应内容并在流式返回时逐步更新文档
type responseCollector struct {
	writer         io.Writer
	content        string
	documentRepo   *repository.DocumentRepository
	documentID     string
	updateBuffer   string
	bufferSize     int
	ragService     *rag.RAGService
	userID         string
	conversationID string
	workID         string
	role           string
	indexed        bool // 是否已经触发索引
}

// workResponseCollector 收集流式响应内容并在流式返回时逐步更新WorkDocument（v1.3：灵感模式）
type workResponseCollector struct {
	writer           io.Writer
	content          string
	workDocumentRepo *repository.WorkDocumentRepository
	documentID       string
	updateBuffer     string
	bufferSize       int
	ragService       *rag.RAGService
	userID           string
	conversationID   string
	workID           string
	role             string
	indexed          bool // 是否已经触发索引
}

const updateBufferThreshold = 100 // 每100个字符更新一次数据库

func (rc *responseCollector) Write(p []byte) (n int, err error) {
	// 尝试写入到客户端，但如果失败也继续保存到数据库
	// 这样即使客户端断开连接，已接收的内容也会被保存
	_, writeErr := rc.writer.Write(p)

	chunk := string(p)
	rc.content += chunk
	rc.updateBuffer += chunk
	rc.bufferSize += len(chunk)

	// 当缓冲区达到阈值时，更新数据库
	// 即使写入客户端失败，也要保存到数据库
	if rc.bufferSize >= updateBufferThreshold {
		err = rc.documentRepo.AppendContent(rc.documentID, rc.updateBuffer)
		if err != nil {
			// 如果保存数据库失败，返回错误
			// 但如果只是写入客户端失败，不影响数据库保存
			return len(p), err
		}
		rc.updateBuffer = ""
		rc.bufferSize = 0
	}

	// 当内容达到一定长度时（如500字符），触发RAG索引（只触发一次）
	if !rc.indexed && len(rc.content) >= 500 && rc.ragService != nil {
		rc.indexed = true
		// 异步索引，不阻塞流式响应
		go func() {
			// 获取当前文档内容进行索引
			doc, err := rc.documentRepo.GetByID(rc.documentID)
			if err == nil && doc != nil {
				rc.ragService.IndexDocument(doc.ID, doc.UserID, doc.ConversationID, "", doc.Content, doc.Role)
			}
		}()
	}

	// 如果写入客户端失败，返回错误，但不影响数据库保存
	// 这样上层可以知道客户端断开，但数据库已经保存了内容
	if writeErr != nil {
		return len(p), writeErr
	}

	return len(p), nil
}

func (rc *workResponseCollector) Write(p []byte) (n int, err error) {
	// 尝试写入到客户端，但如果失败也继续保存到数据库
	// 这样即使客户端断开连接，已接收的内容也会被保存
	_, writeErr := rc.writer.Write(p)

	chunk := string(p)
	rc.content += chunk
	rc.updateBuffer += chunk
	rc.bufferSize += len(chunk)

	// 当缓冲区达到阈值时，更新数据库
	// 即使写入客户端失败，也要保存到数据库
	if rc.bufferSize >= updateBufferThreshold {
		err = rc.workDocumentRepo.AppendContent(rc.documentID, rc.updateBuffer)
		if err != nil {
			// 如果保存数据库失败，返回错误
			// 但如果只是写入客户端失败，不影响数据库保存
			return len(p), err
		}
		rc.updateBuffer = ""
		rc.bufferSize = 0
	}

	// 当内容达到一定长度时（如500字符），触发RAG索引（只触发一次）
	if !rc.indexed && len(rc.content) >= 500 && rc.ragService != nil {
		rc.indexed = true
		// 异步索引，不阻塞流式响应
		go func() {
			// 获取当前文档内容进行索引
			doc, err := rc.workDocumentRepo.GetByIDAndUserID(rc.documentID, rc.userID)
			if err == nil && doc != nil {
				rc.ragService.IndexDocument(doc.ID, doc.UserID, "", doc.WorkID, doc.Content, doc.Role)
			}
		}()
	}

	// 如果写入客户端失败，返回错误，但不影响数据库保存
	// 这样上层可以知道客户端断开，但数据库已经保存了内容
	if writeErr != nil {
		return len(p), writeErr
	}

	return len(p), nil
}
