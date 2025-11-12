package repository

import (
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

type WorkDocumentRepository struct {
	db *gorm.DB
}

func NewWorkDocumentRepository(db *gorm.DB) *WorkDocumentRepository {
	return &WorkDocumentRepository{db: db}
}

// Create 创建创作文档
func (r *WorkDocumentRepository) Create(doc *models.WorkDocument) error {
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	return r.db.Create(doc).Error
}

// GetByIDAndUserID 根据ID和用户ID获取文档
func (r *WorkDocumentRepository) GetByIDAndUserID(id, userID string) (*models.WorkDocument, error) {
	var doc models.WorkDocument
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetByWorkIDAndUserID 根据创作ID和用户ID获取所有文档
func (r *WorkDocumentRepository) GetByWorkIDAndUserID(workID, userID string) ([]models.WorkDocument, error) {
	var docs []models.WorkDocument
	err := r.db.Where("work_id = ? AND user_id = ?", workID, userID).Order("created_at ASC").Find(&docs).Error
	return docs, err
}

// Update 更新文档
func (r *WorkDocumentRepository) Update(doc *models.WorkDocument) error {
	doc.UpdatedAt = time.Now()
	return r.db.Save(doc).Error
}

// UpdateTitleByIDAndUserID 更新文档标题
func (r *WorkDocumentRepository) UpdateTitleByIDAndUserID(id, userID, title string) error {
	return r.db.Model(&models.WorkDocument{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("title", title).Error
}

// UpdateContentByIDAndUserID 更新文档内容
func (r *WorkDocumentRepository) UpdateContentByIDAndUserID(id, userID, content string) error {
	return r.db.Model(&models.WorkDocument{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"content":    content,
			"updated_at": time.Now(),
		}).Error
}

// DeleteByIDAndUserID 删除文档
func (r *WorkDocumentRepository) DeleteByIDAndUserID(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.WorkDocument{}).Error
}
