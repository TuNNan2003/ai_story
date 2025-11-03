package repository

import (
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create 创建文档
func (r *DocumentRepository) Create(document *models.Document) error {
	document.CreatedAt = time.Now()
	document.UpdatedAt = time.Now()
	return r.db.Create(document).Error
}

// GetByID 根据ID获取文档
func (r *DocumentRepository) GetByID(id string) (*models.Document, error) {
	var document models.Document
	err := r.db.Where("id = ?", id).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// GetByConversationID 根据对话ID获取文档列表
func (r *DocumentRepository) GetByConversationID(conversationID string) ([]models.Document, error) {
	var documents []models.Document
	err := r.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

// Update 更新文档
func (r *DocumentRepository) Update(document *models.Document) error {
	document.UpdatedAt = time.Now()
	return r.db.Save(document).Error
}

// Delete 删除文档
func (r *DocumentRepository) Delete(id string) error {
	return r.db.Delete(&models.Document{}, "id = ?", id).Error
}

// DeleteByConversationID 删除对话的所有文档
func (r *DocumentRepository) DeleteByConversationID(conversationID string) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&models.Document{}).Error
}

// AppendContent 追加内容到文档（用于流式更新）
func (r *DocumentRepository) AppendContent(id string, content string) error {
	var doc models.Document
	err := r.db.Where("id = ?", id).First(&doc).Error
	if err != nil {
		return err
	}
	doc.Content = doc.Content + content
	doc.UpdatedAt = time.Now()
	return r.db.Save(&doc).Error
}

// UpdateContent 更新文档内容（用于流式更新的初始设置）
func (r *DocumentRepository) UpdateContent(id string, content string) error {
	return r.db.Model(&models.Document{}).
		Where("id = ?", id).
		Update("content", content).
		Update("updated_at", time.Now()).
		Error
}
