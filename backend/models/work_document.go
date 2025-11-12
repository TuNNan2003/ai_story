package models

import "time"

// WorkDocument 创作文档模型
type WorkDocument struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	WorkID    string    `json:"work_id" gorm:"index"`     // 所属创作ID
	UserID    string    `json:"user_id" gorm:"index"`     // 用户ID
	Title     string    `json:"title"`                    // 文档标题
	Content   string    `json:"content" gorm:"type:text"` // 文档内容
	CreatedAt time.Time `json:"created_at"`               // 创建时间
	UpdatedAt time.Time `json:"updated_at"`               // 更新时间
}

// TableName 指定表名
func (WorkDocument) TableName() string {
	return "work_documents"
}
