package repository

import (
	"fmt"
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

// StoryRepository 故事仓库
type StoryRepository struct {
	db *gorm.DB
}

// NewStoryRepository 创建故事仓库
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

// GetByGuidAndUserID 根据Guid和用户ID获取故事（确保数据隔离）
func (r *StoryRepository) GetByGuidAndUserID(Guid, userID string) ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Preload("Document").Where("guid = ? AND user_id = ?", Guid, userID).Order("created_at DESC").Find(&stories).Error
	if err != nil {
		fmt.Printf("[Story_repo GetByGuidAndUserID] Error: %+v\n", err)
		return nil, err
	}
	return stories, nil
}

// GetAll 获取所有故事（用于默认guid）
func (r *StoryRepository) GetAll() ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Preload("Document").Order("created_at DESC").Find(&stories).Error
	if err != nil {
		fmt.Printf("[Story_repo GetAll] Error: %+v\n", err)
		return nil, err
	}
	return stories, nil
}

// GetAllByUserID 根据用户ID获取所有故事（确保数据隔离）
func (r *StoryRepository) GetAllByUserID(userID string) ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Preload("Document").Where("user_id = ?", userID).Order("created_at DESC").Find(&stories).Error
	if err != nil {
		fmt.Printf("[Story_repo GetAllByUserID] Error: %+v\n", err)
		return nil, err
	}
	return stories, nil
}

// Delete 删除文档
func (r *StoryRepository) Delete(id string) error {
	return r.db.Delete(&models.Story{}, "id = ?", id).Error
}

// DeleteByIDAndUserID 根据ID和用户ID删除故事（确保数据隔离）
func (r *StoryRepository) DeleteByIDAndUserID(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Story{}).Error
}

// GetByContentHash 根据内容特征值查找故事
func (r *StoryRepository) GetByContentHash(guid, contentHash string) (*models.Story, error) {
	var story models.Story
	err := r.db.Where("guid = ? AND content_hash = ?", guid, contentHash).First(&story).Error
	if err != nil {
		// 如果是记录不存在的错误，返回nil而不是错误
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &story, nil
}

// GetByContentHashAndUserID 根据内容特征值和用户ID查找故事（确保数据隔离）
func (r *StoryRepository) GetByContentHashAndUserID(guid, userID, contentHash string) (*models.Story, error) {
	var story models.Story
	err := r.db.Where("guid = ? AND user_id = ? AND content_hash = ?", guid, userID, contentHash).First(&story).Error
	if err != nil {
		// 如果是记录不存在的错误，返回nil而不是错误
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &story, nil
}

// GetByID 根据ID获取故事
func (r *StoryRepository) GetByID(id string) (*models.Story, error) {
	var story models.Story
	err := r.db.Where("id = ?", id).First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// GetByIDAndUserID 根据ID和用户ID获取故事（确保数据隔离）
func (r *StoryRepository) GetByIDAndUserID(id, userID string) (*models.Story, error) {
	var story models.Story
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// Update 更新故事
func (r *StoryRepository) Update(story *models.Story) error {
	story.UpdatedAt = time.Now()
	return r.db.Save(story).Error
}
