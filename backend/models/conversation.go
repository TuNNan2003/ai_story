package models

import "time"

// Conversation 对话模型
type Conversation struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	UserID      string     `json:"user_id" gorm:"index"`                       // 用户ID
	Title       string     `json:"title"`                                      // 对话标题
	DocumentIDs string     `json:"document_ids" gorm:"type:text"`              // 文档ID列表，按顺序排列，用逗号分隔
	CreatedAt   time.Time  `json:"created_at"`                                 // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`                                 // 更新时间
	Documents   []Document `json:"documents" gorm:"foreignKey:ConversationID"` // 关联的文档列表
}

// TableName 指定表名
func (Conversation) TableName() string {
	return "conversations"
}
