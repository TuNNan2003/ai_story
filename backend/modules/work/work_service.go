package work

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/utils"
)

// WorkService 创作服务
type WorkService struct {
	workRepo         *repository.WorkRepository
	workDocumentRepo *repository.WorkDocumentRepository
}

// NewWorkService 创建创作服务
func NewWorkService(workRepo *repository.WorkRepository, workDocumentRepo *repository.WorkDocumentRepository) *WorkService {
	return &WorkService{
		workRepo:         workRepo,
		workDocumentRepo: workDocumentRepo,
	}
}

// GetWorkList 获取创作列表
func (s *WorkService) GetWorkList(userID string) (*models.WorkResponse, error) {
	works, err := s.workRepo.GetAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &models.WorkResponse{
		Works: works,
		Total: len(works),
	}, nil
}

// CreateWork 创建创作
func (s *WorkService) CreateWork(userID, title string) (*models.Work, error) {
	work := &models.Work{
		ID:     utils.GenerateID(),
		UserID: userID,
		Title:  title,
	}

	if err := s.workRepo.Create(work); err != nil {
		return nil, err
	}

	return work, nil
}

// UpdateWorkTitle 更新创作标题
func (s *WorkService) UpdateWorkTitle(id, userID, title string) error {
	return s.workRepo.UpdateTitleByIDAndUserID(id, userID, title)
}

// DeleteWork 删除创作
func (s *WorkService) DeleteWork(id, userID string) error {
	return s.workRepo.DeleteByIDAndUserID(id, userID)
}

// GetWorkDocuments 获取创作的所有文档
func (s *WorkService) GetWorkDocuments(workID, userID string) (*models.WorkDocumentResponse, error) {
	docs, err := s.workDocumentRepo.GetByWorkIDAndUserID(workID, userID)
	if err != nil {
		return nil, err
	}

	return &models.WorkDocumentResponse{
		Documents: docs,
		Total:     len(docs),
	}, nil
}

// CreateWorkDocument 创建创作文档
func (s *WorkService) CreateWorkDocument(userID, workID, title, content string) (*models.WorkDocument, error) {
	doc := &models.WorkDocument{
		ID:      utils.GenerateID(),
		WorkID:  workID,
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	if err := s.workDocumentRepo.Create(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// UpdateWorkDocumentTitle 更新文档标题
func (s *WorkService) UpdateWorkDocumentTitle(id, userID, title string) error {
	return s.workDocumentRepo.UpdateTitleByIDAndUserID(id, userID, title)
}

// UpdateWorkDocumentContent 更新文档内容
func (s *WorkService) UpdateWorkDocumentContent(id, userID, content string) error {
	return s.workDocumentRepo.UpdateContentByIDAndUserID(id, userID, content)
}

// DeleteWorkDocument 删除文档
func (s *WorkService) DeleteWorkDocument(id, userID string) error {
	return s.workDocumentRepo.DeleteByIDAndUserID(id, userID)
}

// GetWorkDocumentByID 根据ID获取文档
func (s *WorkService) GetWorkDocumentByID(id, userID string) (*models.WorkDocument, error) {
	return s.workDocumentRepo.GetByIDAndUserID(id, userID)
}

