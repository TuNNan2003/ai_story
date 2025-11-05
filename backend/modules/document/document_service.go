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

// GetDocumentList 获取文档列表（支持翻页）
func (s *DocumentService) GetDocumentList(conversationID string, beforeID string, limit int) (*models.DocumentListResponse, error) {
	// 如果没有指定limit，默认为10
	if limit <= 0 {
		limit = 10
	}

	// 如果没有指定beforeID，获取最新的文档（按created_at倒序，取前limit个，然后反转）
	if beforeID == "" {
		documents, err := s.documentRepo.GetLatestDocumentsByConversationID(conversationID, limit)
		if err != nil {
			return nil, err
		}
		// 反转顺序，使其按时间正序排列
		for i, j := 0, len(documents)-1; i < j; i, j = i+1, j-1 {
			documents[i], documents[j] = documents[j], documents[i]
		}
		return &models.DocumentListResponse{
			Documents: documents,
			Total:     len(documents),
		}, nil
	}

	// 如果指定了beforeID，获取比该ID更早的文档
	documentIDs, err := s.documentRepo.GetDocumentIDsByConversationID(conversationID, beforeID, limit)
	if err != nil {
		return nil, err
	}

	documents, err := s.GetDocumentsByIDs(documentIDs)
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

// GetDocumentIDsByConversationID 获取对话的文档ID列表（用于翻页）
func (s *DocumentService) GetDocumentIDsByConversationID(conversationID string, beforeDocumentID string, limit int) ([]string, error) {
	return s.documentRepo.GetDocumentIDsByConversationID(conversationID, beforeDocumentID, limit)
}

// GetDocumentsByIDs 根据文档ID列表批量获取文档
func (s *DocumentService) GetDocumentsByIDs(documentIDs []string) ([]models.Document, error) {
	if len(documentIDs) == 0 {
		return []models.Document{}, nil
	}

	var documents []models.Document
	for _, id := range documentIDs {
		doc, err := s.documentRepo.GetByID(id)
		if err != nil {
			continue // 跳过找不到的文档
		}
		documents = append(documents, *doc)
	}

	return documents, nil
}
