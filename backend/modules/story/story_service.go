package story

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/utils"
	"time"
)

type StoryService struct {
	storyRepo *repository.StoryRepository
}

func NewStoryService(storyRepo *repository.StoryRepository) *StoryService {
	return &StoryService{
		storyRepo: storyRepo,
	}
}

// GetStoryList 获取文档列表
func (s *StoryService) GetStoryList(guid string) (*models.StoryResponse, error) {
	stories, err := s.storyRepo.GetByGuid(guid)
	if err != nil {
		return nil, err
	}

	return &models.StoryResponse{
		Story: stories,
		Total: len(stories),
	}, nil
}

// DeleteStory 删除文档
func (s *StoryService) DeleteStory(id string) error {
	return s.storyRepo.Delete(id)
}

// CreateStory 创建文档
func (s *StoryService) CreateStory(guid, documentID, title string) (*models.Story, error) {
	story := &models.Story{
		ID:         utils.GenerateStoryId(),
		Title:      title,
		Guid:       guid,
		DocumentID: documentID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := s.storyRepo.Create(story)
	if err != nil {
		return nil, err
	}
	return story, nil
}
