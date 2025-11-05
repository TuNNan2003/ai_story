package conversation_list

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/services"
	"grandma/backend/utils"
	"strings"
)

type ConversationListService struct {
	conversationRepo *repository.ConversationRepository
	config           *TitleGenerationConfig
}

type TitleGenerationConfig struct {
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	AnthropicAPIKey  string
	AnthropicBaseURL string
	DefaultModel     string // 默认使用哪个模型生成标题
}

func NewConversationListService(conversationRepo *repository.ConversationRepository, config *TitleGenerationConfig) *ConversationListService {
	return &ConversationListService{
		conversationRepo: conversationRepo,
		config:           config,
	}
}

// GetConversationList 获取对话列表
func (s *ConversationListService) GetConversationList(page, pageSize int) (*models.ConversationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	conversations, total, err := s.conversationRepo.List(page, pageSize)
	if err != nil {
		return nil, err
	}

	return &models.ConversationListResponse{
		Conversations: conversations,
		Total:         int(total),
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// CreateNewConversation 创建新对话
func (s *ConversationListService) CreateNewConversation() (*models.Conversation, error) {
	conversationID := utils.GenerateConversationID()
	conversation := &models.Conversation{
		ID:          conversationID,
		Title:       "新对话",
		DocumentIDs: "",
	}
	err := s.conversationRepo.Create(conversation)
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

// CreateNewConversationWithTitle 创建新对话并生成标题
func (s *ConversationListService) CreateNewConversationWithTitle(userInputs []string) (*models.Conversation, error) {
	conversationID := utils.GenerateConversationID()

	// 生成标题
	title := "新对话"
	if len(userInputs) > 0 {
		generatedTitle, err := s.generateTitle(userInputs)
		if err == nil && generatedTitle != "" {
			title = generatedTitle
		}
	}

	conversation := &models.Conversation{
		ID:          conversationID,
		Title:       title,
		DocumentIDs: "",
	}
	err := s.conversationRepo.Create(conversation)
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

// generateTitle 根据用户输入生成对话标题
func (s *ConversationListService) generateTitle(userInputs []string) (string, error) {
	if len(userInputs) == 0 {
		return "新对话", nil
	}

	// 构建提示词
	userMessagesText := strings.Join(userInputs, "\n")
	prompt := "请根据以下用户输入，生成一个简洁的对话标题（不超过20个字，不要包含标点符号）：\n\n" + userMessagesText

	// 构建消息
	messages := []models.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 获取provider
	model := s.config.DefaultModel
	if model == "" {
		model = "openai" // 默认使用openai
	}

	provider, err := services.GetProvider(
		model,
		s.config.OpenAIAPIKey,
		s.config.OpenAIBaseURL,
		s.config.AnthropicAPIKey,
		s.config.AnthropicBaseURL,
	)
	if err != nil {
		return "", err
	}

	// 调用LLM生成标题
	title, err := provider.Chat(messages)
	if err != nil {
		return "", err
	}

	// 清理标题（去除前后空格、换行等）
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\n", "")
	title = strings.ReplaceAll(title, "\r", "")

	// 限制长度
	if len(title) > 50 {
		title = title[:50]
	}

	if title == "" {
		title = "新对话"
	}

	return title, nil
}

// GenerateTitleForConversation 为对话生成标题（公开方法，用于智能命名接口）
func (s *ConversationListService) GenerateTitleForConversation(userInputs []string) (string, error) {
	return s.generateTitle(userInputs)
}
