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

func GenerateStoryId() string {
	return generateID("story")
}

// generateID 生成唯一ID
func generateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randomHex)
}
