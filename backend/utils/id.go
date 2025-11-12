package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateConversationID 生成对话ID
func GenerateConversationID() string {
	return generateID("conv")
}

// GenerateDocumentID 生成文档ID
func GenerateDocumentID() string {
	return generateID("doc")
}

// GenerateStoryId 生成故事ID
func GenerateStoryId() string {
	return generateID("story")
}

// GenerateID 生成通用唯一ID（不带前缀）
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d_%s", timestamp, randomHex)
}

// generateID 生成唯一ID
func generateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randomHex)
}
