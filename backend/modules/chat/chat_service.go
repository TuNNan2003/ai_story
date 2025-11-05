package chat

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/services"
	"grandma/backend/utils"
	"io"
	"strings"
)

type ChatService struct {
	conversationRepo *repository.ConversationRepository
	documentRepo     *repository.DocumentRepository
	config           *ChatConfig
}

type ChatConfig struct {
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	AnthropicAPIKey  string
	AnthropicBaseURL string
}

func NewChatService(conversationRepo *repository.ConversationRepository, documentRepo *repository.DocumentRepository, config *ChatConfig) *ChatService {
	return &ChatService{
		conversationRepo: conversationRepo,
		documentRepo:     documentRepo,
		config:           config,
	}
}

// SendMessage 发送消息并获取流式响应
func (s *ChatService) SendMessage(req *models.ChatRequest, writer io.Writer) (string, string, error) {
	var conversationID string
	var conversation *models.Conversation
	var err error

	// 如果没有提供对话ID，创建新对话
	if req.ConversationID == "" {
		conversationID = utils.GenerateConversationID()
		conversation = &models.Conversation{
			ID:          conversationID,
			Title:       s.generateTitle(req.Messages),
			DocumentIDs: "",
		}
		err = s.conversationRepo.Create(conversation)
		if err != nil {
			return "", "", err
		}
	} else {
		conversationID = req.ConversationID
		conversation, err = s.conversationRepo.GetByID(conversationID)
		if err != nil {
			return "", "", err
		}
	}

	// 构建API调用的消息数组
	var apiMessages []models.Message

	// 如果提供了对话ID，从数据库加载历史消息
	// 优化：只加载最近的对话上下文，而不是全部历史
	if req.ConversationID != "" {
		// 获取对话的历史文档（只获取最近的20条，避免上下文过长）
		historyDocs, err := s.documentRepo.GetLatestDocumentsByConversationID(conversationID, 20)
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
		writer:       writer,
		content:      "",
		documentRepo: s.documentRepo,
		documentID:   assistantDocID,
		updateBuffer: "",
		bufferSize:   0,
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
	writer       io.Writer
	content      string
	documentRepo *repository.DocumentRepository
	documentID   string
	updateBuffer string
	bufferSize   int
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

	// 如果写入客户端失败，返回错误，但不影响数据库保存
	// 这样上层可以知道客户端断开，但数据库已经保存了内容
	if writeErr != nil {
		return len(p), writeErr
	}

	return len(p), nil
}
