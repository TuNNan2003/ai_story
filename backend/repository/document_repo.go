package repository

import (
	"fmt"
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
		fmt.Printf("[document_repo GetByConversationID] Error: %+v\n", err)
		return nil, err
	}
	fmt.Printf("[document_repo GetByConversationID] Documents: %+v\n", documents)
	return documents, nil
}

// GetDocumentIDsByConversationID 根据对话ID获取文档ID列表（用于翻页）
// beforeDocumentID: 如果提供，返回比该文档更早的文档ID
// limit: 返回的最大数量
func (r *DocumentRepository) GetDocumentIDsByConversationID(conversationID string, beforeDocumentID string, limit int) ([]string, error) {
	var documentIDs []string
	query := r.db.Model(&models.Document{}).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC")

	if beforeDocumentID != "" {
		// 找到 beforeDocumentID 对应的文档的 created_at
		var beforeDoc models.Document
		if err := r.db.Where("id = ?", beforeDocumentID).First(&beforeDoc).Error; err == nil {
			query = query.Where("created_at < ?", beforeDoc.CreatedAt)
		}
	}

	if limit <= 0 {
		limit = 10
	}

	err := query.Limit(limit).Pluck("id", &documentIDs).Error
	if err != nil {
		return nil, err
	}

	return documentIDs, nil
}

// GetLatestDocumentsByConversationID 获取对话的最新文档（按created_at倒序）
func (r *DocumentRepository) GetLatestDocumentsByConversationID(conversationID string, limit int) ([]models.Document, error) {
	var documents []models.Document
	query := r.db.Where("conversation_id = ?", conversationID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&documents).Error
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
