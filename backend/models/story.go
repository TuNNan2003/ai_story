package models

import "time"

type Story struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	DocumentID string    `json:"document_id"`
	Guid       string    `json:"guid"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Document   Document  `json:"document" gorm:"foreignKey:DocumentID"`
}

// TableName 指定表名
func (Story) TableName() string {
	return "stories"
}
