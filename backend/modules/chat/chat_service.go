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
	if req.ConversationID != "" {
		// 获取对话的历史文档
		historyDocs, err := s.documentRepo.GetByConversationID(conversationID)
		if err == nil && len(historyDocs) > 0 {
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
	if err != nil {
		return "", "", err
	}

	// 刷新剩余的缓冲区内容
	if responseCollector.updateBuffer != "" {
		err = s.documentRepo.AppendContent(assistantDocID, responseCollector.updateBuffer)
		if err != nil {
			return "", "", err
		}
	}

	// 添加助手文档ID到对话的文档ID列表
	err = s.conversationRepo.AppendDocumentID(conversationID, assistantDocID)
	if err != nil {
		return "", "", err
	}

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
	n, err = rc.writer.Write(p)
	if err != nil {
		return n, err
	}

	chunk := string(p)
	rc.content += chunk
	rc.updateBuffer += chunk
	rc.bufferSize += len(chunk)

	// 当缓冲区达到阈值时，更新数据库
	if rc.bufferSize >= updateBufferThreshold {
		err = rc.documentRepo.AppendContent(rc.documentID, rc.updateBuffer)
		if err != nil {
			return n, err
		}
		rc.updateBuffer = ""
		rc.bufferSize = 0
	}

	return n, nil
}
