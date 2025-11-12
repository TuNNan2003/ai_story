package models

import "time"

// Work 创作模型
type Work struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"user_id" gorm:"index"`               // 用户ID
	Title     string         `json:"title"`                              // 创作标题
	CreatedAt time.Time      `json:"created_at"`                         // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                         // 更新时间
	Documents []WorkDocument `json:"documents" gorm:"foreignKey:WorkID"` // 关联的文档列表
}

// TableName 指定表名
func (Work) TableName() string {
	return "works"
}
