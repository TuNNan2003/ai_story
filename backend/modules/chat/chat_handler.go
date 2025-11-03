package chat

import (
	"fmt"
	"grandma/backend/models"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *ChatService
}

func NewChatHandler(chatService *ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	fmt.Println("[chat_handler Chat] Start")
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("[chat_handler Chat] Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// 创建writer来包装响应以支持流式输出
	writer := &streamWriter{writer: c.Writer}

	// 发送消息并获取响应
	fmt.Printf("[chat_handler Chat] Req:%+v\n", req)
	conversationID, documentID, err := h.chatService.SendMessage(&req, writer)
	if err != nil {
		fmt.Printf("[chat_handler Chat] Req:%+v, Error: %v\n", req, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回对话ID和文档ID（可选）
	_ = conversationID
	_ = documentID
}

// streamWriter 包装响应写入器以支持流式输出
type streamWriter struct {
	writer io.Writer
}

func (sw *streamWriter) Write(p []byte) (n int, err error) {
	n, err = sw.writer.Write(p)
	if err != nil {
		return n, err
	}

	// 每次写入后刷新，确保流式输出
	if flusher, ok := sw.writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return n, nil
}
