package handlers

import (
	"grandma/backend/config"
	"grandma/backend/models"
	"grandma/backend/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	config *config.Config
}

func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		config: cfg,
	}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req struct {
		Model   string `json:"model" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// 创建provider
	provider, err := services.GetProvider(
		req.Model,
		h.config.OpenAIAPIKey,
		h.config.OpenAIBaseURL,
		h.config.AnthropicAPIKey,
		h.config.AnthropicBaseURL,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 流式返回响应
	c.Stream(func(w io.Writer) bool {
		if err := provider.ChatStream([]models.Message{
			{
				Role:    "user",
				Content: req.Message,
			},
		}, w); err != nil {
			return false
		}
		return false
	})
}
