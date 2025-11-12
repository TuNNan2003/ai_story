package repository

import (
	"grandma/backend/models"
	"time"

	"gorm.io/gorm"
)

type VectorChunkRepository struct {
	db *gorm.DB
}

func NewVectorChunkRepository(db *gorm.DB) *VectorChunkRepository {
	return &VectorChunkRepository{db: db}
}

// Create 创建向量chunk
func (r *VectorChunkRepository) Create(chunk *models.VectorChunk) error {
	chunk.CreatedAt = time.Now()
	chunk.UpdatedAt = time.Now()
	return r.db.Create(chunk).Error
}

// GetByID 根据ID获取chunk
func (r *VectorChunkRepository) GetByID(id string) (*models.VectorChunk, error) {
	var chunk models.VectorChunk
	err := r.db.Where("id = ?", id).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// GetByDocumentID 根据文档ID获取所有chunks
func (r *VectorChunkRepository) GetByDocumentID(documentID string) ([]models.VectorChunk, error) {
	var chunks []models.VectorChunk
	err := r.db.Where("document_id = ?", documentID).Find(&chunks).Error
	return chunks, err
}

// GetByConversationID 根据对话ID获取所有chunks
func (r *VectorChunkRepository) GetByConversationID(conversationID string) ([]models.VectorChunk, error) {
	var chunks []models.VectorChunk
	err := r.db.Where("conversation_id = ?", conversationID).Find(&chunks).Error
	return chunks, err
}

// GetByWorkID 根据创作ID获取所有chunks
func (r *VectorChunkRepository) GetByWorkID(workID string) ([]models.VectorChunk, error) {
	var chunks []models.VectorChunk
	err := r.db.Where("work_id = ?", workID).Find(&chunks).Error
	return chunks, err
}

// GetByUserID 根据用户ID获取所有chunks（用于数据隔离）
func (r *VectorChunkRepository) GetByUserID(userID string) ([]models.VectorChunk, error) {
	var chunks []models.VectorChunk
	err := r.db.Where("user_id = ?", userID).Find(&chunks).Error
	return chunks, err
}

// GetAllWithEmbeddings 获取所有带向量嵌入的chunks（用于相似度搜索）
// 注意：这个方法会加载所有chunks到内存，适用于小规模数据
// 对于大规模数据，应该使用分页或更高效的查询方式
// excludeConversationID 和 excludeWorkID 用于排除当前对话/创作，避免重复
func (r *VectorChunkRepository) GetAllWithEmbeddings(userID string, excludeConversationID, excludeWorkID string) ([]models.VectorChunk, error) {
	var chunks []models.VectorChunk
	query := r.db.Where("user_id = ?", userID)

	// 排除当前对话/创作中的文档，避免重复
	if excludeConversationID != "" {
		query = query.Where("(conversation_id != ? OR conversation_id IS NULL OR conversation_id = '')", excludeConversationID)
	}
	if excludeWorkID != "" {
		query = query.Where("(work_id != ? OR work_id IS NULL OR work_id = '')", excludeWorkID)
	}

	err := query.Where("embedding_json != '' AND embedding_json IS NOT NULL").Find(&chunks).Error
	return chunks, err
}

// DeleteByDocumentID 删除文档的所有chunks
func (r *VectorChunkRepository) DeleteByDocumentID(documentID string) error {
	return r.db.Where("document_id = ?", documentID).Delete(&models.VectorChunk{}).Error
}

// DeleteByConversationID 删除对话的所有chunks
func (r *VectorChunkRepository) DeleteByConversationID(conversationID string) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&models.VectorChunk{}).Error
}

// DeleteByWorkID 删除创作的所有chunks
func (r *VectorChunkRepository) DeleteByWorkID(workID string) error {
	return r.db.Where("work_id = ?", workID).Delete(&models.VectorChunk{}).Error
}

// Update 更新chunk
func (r *VectorChunkRepository) Update(chunk *models.VectorChunk) error {
	chunk.UpdatedAt = time.Now()
	return r.db.Save(chunk).Error
}

