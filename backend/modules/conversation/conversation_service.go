package conversation

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/utils"
)

// ConversationService 对话服务
type ConversationService struct {
	conversationRepo *repository.ConversationRepository
	documentRepo     *repository.DocumentRepository
}

// NewConversationService 创建对话服务
func NewConversationService(conversationRepo *repository.ConversationRepository, documentRepo *repository.DocumentRepository) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		documentRepo:     documentRepo,
	}
}

// GetConversationByID 根据ID获取对话
func (s *ConversationService) GetConversationByID(id string) (*models.Conversation, error) {
	return s.conversationRepo.GetByID(id)
}

// CreateConversation 创建对话
func (s *ConversationService) CreateConversation(title string) (*models.Conversation, error) {
	conversation := &models.Conversation{
		ID:    utils.GenerateConversationID(),
		Title: title,
	}
	err := s.conversationRepo.Create(conversation)
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

// UpdateConversation 更新对话
func (s *ConversationService) UpdateConversation(conversation *models.Conversation) error {
	return s.conversationRepo.Update(conversation)
}

// UpdateConversationTitle 更新对话标题
func (s *ConversationService) UpdateConversationTitle(id, userID, title string) error {
	return s.conversationRepo.UpdateTitleByIDAndUserID(id, userID, title)
}

// DeleteConversation 删除对话（包括关联的文档）
func (s *ConversationService) DeleteConversation(id, userID string) error {
	// 先验证对话属于该用户
	conversation, err := s.conversationRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		return err
	}

	// 先删除关联的文档（需要验证用户ID）
	// 注意：这里需要确保文档也属于该用户
	err = s.documentRepo.DeleteByConversationID(conversation.ID)
	if err != nil {
		return err
	}
	// 再删除对话
	return s.conversationRepo.DeleteByIDAndUserID(id, userID)
}
