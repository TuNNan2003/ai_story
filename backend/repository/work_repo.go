package repository

import (
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

// WorkRepository 创作仓库
type WorkRepository struct {
	db *gorm.DB
}

// NewWorkRepository 创建创作仓库
func NewWorkRepository(db *gorm.DB) *WorkRepository {
	return &WorkRepository{db: db}
}

// Create 创建创作
func (r *WorkRepository) Create(work *models.Work) error {
	work.CreatedAt = time.Now()
	work.UpdatedAt = time.Now()
	return r.db.Create(work).Error
}

// GetByIDAndUserID 根据ID和用户ID获取创作
func (r *WorkRepository) GetByIDAndUserID(id, userID string) (*models.Work, error) {
	var work models.Work
	err := r.db.Preload("Documents").Where("id = ? AND user_id = ?", id, userID).First(&work).Error
	if err != nil {
		return nil, err
	}
	return &work, nil
}

// GetAllByUserID 获取用户的所有创作
func (r *WorkRepository) GetAllByUserID(userID string) ([]models.Work, error) {
	var works []models.Work
	err := r.db.Preload("Documents").Where("user_id = ?", userID).Order("updated_at DESC").Find(&works).Error
	return works, err
}

// Update 更新创作
func (r *WorkRepository) Update(work *models.Work) error {
	work.UpdatedAt = time.Now()
	return r.db.Save(work).Error
}

// UpdateTitleByIDAndUserID 更新创作标题
func (r *WorkRepository) UpdateTitleByIDAndUserID(id, userID, title string) error {
	return r.db.Model(&models.Work{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("title", title).Error
}

// DeleteByIDAndUserID 删除创作
func (r *WorkRepository) DeleteByIDAndUserID(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Work{}).Error
}
