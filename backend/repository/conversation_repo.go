package repository

import (
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create 创建对话
func (r *ConversationRepository) Create(conversation *models.Conversation) error {
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()
	return r.db.Create(conversation).Error
}

// GetByID 根据ID获取对话
func (r *ConversationRepository) GetByID(id string) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.Preload("Documents").Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// List 获取对话列表
func (r *ConversationRepository) List(page, pageSize int) ([]models.Conversation, int64, error) {
	var conversations []models.Conversation
	var total int64

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	err := r.db.Model(&models.Conversation{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

// Update 更新对话
func (r *ConversationRepository) Update(conversation *models.Conversation) error {
	conversation.UpdatedAt = time.Now()
	return r.db.Save(conversation).Error
}

// Delete 删除对话
func (r *ConversationRepository) Delete(id string) error {
	return r.db.Delete(&models.Conversation{}, "id = ?", id).Error
}

// UpdateTitle 更新对话标题
func (r *ConversationRepository) UpdateTitle(id, title string) error {
	return r.db.Model(&models.Conversation{}).Where("id = ?", id).Update("title", title).Error
}

// AppendDocumentID 添加文档ID到对话的文档ID列表
func (r *ConversationRepository) AppendDocumentID(id, documentID string) error {
	conversation, err := r.GetByID(id)
	if err != nil {
		return err
	}

	var newDocumentIDs string
	if conversation.DocumentIDs == "" {
		newDocumentIDs = documentID
	} else {
		newDocumentIDs = conversation.DocumentIDs + "," + documentID
	}

	return r.db.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("document_ids", newDocumentIDs).
		Update("updated_at", time.Now()).
		Error
}
