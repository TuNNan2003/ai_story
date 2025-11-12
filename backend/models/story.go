package models

import "time"

// Story 故事模型
type Story struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"` // 用户ID
	DocumentID  string    `json:"document_id"`
	Guid        string    `json:"guid"`
	Title       string    `json:"title"`
	Content     string    `json:"content" gorm:"type:text"`
	ContentHash string    `json:"content_hash" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Document    Document  `json:"document" gorm:"foreignKey:DocumentID"`
}

// TableName 指定表名
func (Story) TableName() string {
	return "stories"
}
