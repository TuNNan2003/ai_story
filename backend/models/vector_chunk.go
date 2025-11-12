package models

import (
	"encoding/json"
	"time"
)

// VectorChunk 向量chunk模型
type VectorChunk struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	UserID         string    `json:"user_id" gorm:"index"`                    // 用户ID（数据隔离）
	ConversationID string    `json:"conversation_id" gorm:"index"`           // 所属对话ID（可选，用于对话模式）
	WorkID         string    `json:"work_id" gorm:"index"`                    // 所属创作ID（可选，用于灵感模式）
	DocumentID     string    `json:"document_id" gorm:"index"`               // 来源文档ID
	Content        string    `json:"content" gorm:"type:text"`               // chunk文本内容
	EmbeddingJSON  string    `json:"embedding_json" gorm:"type:text"`        // 向量嵌入（JSON格式存储）
	Metadata       string    `json:"metadata" gorm:"type:text"`              // JSON格式元数据（角色、位置、类型等）
	CreatedAt      time.Time `json:"created_at"`                             // 创建时间
	UpdatedAt      time.Time `json:"updated_at"`                              // 更新时间
}

// TableName 指定表名
func (VectorChunk) TableName() string {
	return "vector_chunks"
}

// GetEmbedding 获取向量嵌入
func (v *VectorChunk) GetEmbedding() ([]float32, error) {
	if v.EmbeddingJSON == "" {
		return nil, nil
	}
	var embedding []float32
	err := json.Unmarshal([]byte(v.EmbeddingJSON), &embedding)
	return embedding, err
}

// SetEmbedding 设置向量嵌入
func (v *VectorChunk) SetEmbedding(embedding []float32) error {
	data, err := json.Marshal(embedding)
	if err != nil {
		return err
	}
	v.EmbeddingJSON = string(data)
	return nil
}

// GetMetadataMap 获取元数据映射
func (v *VectorChunk) GetMetadataMap() (map[string]interface{}, error) {
	if v.Metadata == "" {
		return make(map[string]interface{}), nil
	}
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(v.Metadata), &metadata)
	return metadata, err
}

// SetMetadataMap 设置元数据映射
func (v *VectorChunk) SetMetadataMap(metadata map[string]interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	v.Metadata = string(data)
	return nil
}

