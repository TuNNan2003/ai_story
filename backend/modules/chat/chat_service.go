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
			ID:    conversationID,
			Title: s.generateTitle(req.Messages),
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

	// 转换消息格式用于API调用
	apiMessages := make([]models.Message, len(req.Messages))
	for i, msg := range req.Messages {
		apiMessages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 如果是新对话，保存所有历史消息
	// 如果是继续对话，只保存最后一条用户消息（前面的消息应该已经在数据库中）
	if req.ConversationID == "" {
		// 新对话：保存除最后一条之外的所有消息
		for i := 0; i < len(req.Messages)-1; i++ {
			msg := req.Messages[i]
			docID := utils.GenerateDocumentID()
			doc := &models.Document{
				ID:             docID,
				ConversationID: conversationID,
				Role:           msg.Role,
				Content:        msg.Content,
				Model:          req.Model,
			}
			err = s.documentRepo.Create(doc)
			if err != nil {
				return "", "", err
			}
		}
	}

	// 保存最后一条用户消息（必须存在）
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			userDocID := utils.GenerateDocumentID()
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

	// 创建流式响应收集器
	responseCollector := &responseCollector{writer: writer}
	err = provider.ChatStream(apiMessages, responseCollector)
	if err != nil {
		return "", "", err
	}

	// 创建助手消息文档
	assistantDocID := utils.GenerateDocumentID()
	assistantDoc := &models.Document{
		ID:             assistantDocID,
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        responseCollector.content,
		Model:          req.Model,
	}
	err = s.documentRepo.Create(assistantDoc)
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

// responseCollector 收集流式响应内容
type responseCollector struct {
	writer  io.Writer
	content string
}

func (rc *responseCollector) Write(p []byte) (n int, err error) {
	n, err = rc.writer.Write(p)
	if err == nil {
		rc.content += string(p)
	}
	return n, err
}
