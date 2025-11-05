package repository

import (
	"fmt"
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

type StoryRepository struct {
	db *gorm.DB
}

func NewStoryRepository(db *gorm.DB) *StoryRepository {
	return &StoryRepository{db: db}
}

// Create 创建故事
func (r *StoryRepository) Create(story *models.Story) error {
	story.CreatedAt = time.Now()
	story.UpdatedAt = time.Now()
	return r.db.Create(story).Error
}

// GetByGuid 根据Guid获取故事
func (r *StoryRepository) GetByGuid(Guid string) ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Preload("Document").Where("guid = ?", Guid).Order("created_at DESC").Find(&stories).Error
	if err != nil {
		fmt.Printf("[Story_repo GetByGuid] Error: %+v\n", err)
		return nil, err
	}
	fmt.Printf("[Story_repo GetByGuid] Documents: %+v\n", stories)
	return stories, nil
}

// Delete 删除文档
func (r *StoryRepository) Delete(id string) error {
	return r.db.Delete(&models.Story{}, "id = ?", id).Error
}
