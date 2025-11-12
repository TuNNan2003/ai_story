package models

import "time"

// Document 文档模型
type Document struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	UserID         string    `json:"user_id" gorm:"index"`     // 用户ID
	ConversationID string    `json:"conversation_id"`          // 所属对话ID
	Role           string    `json:"role"`                     // 角色：user 或 assistant
	Content        string    `json:"content" gorm:"type:text"` // 文档内容
	Model          string    `json:"model"`                    // 使用的模型
	CreatedAt      time.Time `json:"created_at"`               // 创建时间
	UpdatedAt      time.Time `json:"updated_at"`               // 更新时间
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}
