package document

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/utils"
)

type DocumentService struct {
	documentRepo *repository.DocumentRepository
}

func NewDocumentService(documentRepo *repository.DocumentRepository) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
	}
}

// GetDocumentList 获取文档列表
func (s *DocumentService) GetDocumentList(conversationID string) (*models.DocumentListResponse, error) {
	documents, err := s.documentRepo.GetByConversationID(conversationID)
	if err != nil {
		return nil, err
	}

	return &models.DocumentListResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(id string) (*models.Document, error) {
	return s.documentRepo.GetByID(id)
}

// UpdateDocument 更新文档
func (s *DocumentService) UpdateDocument(document *models.Document) error {
	return s.documentRepo.Update(document)
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(id string) error {
	return s.documentRepo.Delete(id)
}

// CreateDocument 创建文档
func (s *DocumentService) CreateDocument(conversationID, role, content, model string) (*models.Document, error) {
	doc := &models.Document{
		ID:             utils.GenerateDocumentID(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Model:          model,
	}
	err := s.documentRepo.Create(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
