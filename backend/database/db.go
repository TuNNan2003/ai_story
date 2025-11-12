package database

import (
	"grandma/backend/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&models.Conversation{},
		&models.Document{},
		&models.Story{},
		&models.Work{},
		&models.WorkDocument{},
		&models.VectorChunk{},
	)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}
