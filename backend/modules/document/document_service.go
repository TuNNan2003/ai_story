package document

import (
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/utils"
)

// DocumentService 文档服务
type DocumentService struct {
	documentRepo *repository.DocumentRepository
}

// NewDocumentService 创建文档服务
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

	// 如果指定了beforeID，不应该使用这个方法，应该使用 /api/documents/ids 获取ID列表
	// 然后前端并发请求每个文档
	// 这里返回空列表，表示不支持此路径
	return &models.DocumentListResponse{
		Documents: []models.Document{},
		Total:     0,
	}, nil
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(id string) (*models.Document, error) {
	return s.documentRepo.GetByID(id)
}

// GetDocumentByIDAndUserID 根据ID和用户ID获取文档（确保数据隔离）
func (s *DocumentService) GetDocumentByIDAndUserID(id, userID string) (*models.Document, error) {
	return s.documentRepo.GetByIDAndUserID(id, userID)
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

// GetDocumentIDsByConversationIDAndUserID 获取对话的文档ID列表（确保数据隔离）
func (s *DocumentService) GetDocumentIDsByConversationIDAndUserID(conversationID, userID string, beforeDocumentID string, limit int) ([]string, error) {
	return s.documentRepo.GetDocumentIDsByConversationIDAndUserID(conversationID, userID, beforeDocumentID, limit)
}
