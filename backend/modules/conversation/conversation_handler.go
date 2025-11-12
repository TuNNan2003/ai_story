package conversation

import (
	"fmt"
	"grandma/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	service *ConversationService
}

func NewConversationHandler(service *ConversationService) *ConversationHandler {
	return &ConversationHandler{
		service: service,
	}
}

// GetConversationByID 根据ID获取对话
func (h *ConversationHandler) GetConversationByID(c *gin.Context) {
	fmt.Println("[conversation_handler GetConversationByID] Start")
	id := c.Param("id")
	conv, err := h.service.GetConversationByID(id)
	fmt.Printf("[conversation_handler GetConversationByID] conversationID:%v conversationInfo:%v\n", id, conv)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// CreateConversation 创建对话
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	fmt.Println("[conversation_handler CreateConversation] Start")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		req.Title = "新对话"
	}

	conv, err := h.service.CreateConversation(req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// UpdateConversation 更新对话
func (h *ConversationHandler) UpdateConversation(c *gin.Context) {
	fmt.Println("[conversation_handler UpdateConversation] Start")
	id := c.Param("id")
	var conv models.Conversation
	if err := c.ShouldBindJSON(&conv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv.ID = id
	err := h.service.UpdateConversation(&conv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation updated successfully"})
}

// UpdateConversationTitle 更新对话标题
func (h *ConversationHandler) UpdateConversationTitle(c *gin.Context) {
	fmt.Println("[conversation_handler UpdateConversationTitle] Start")
	id := c.Param("id")
	var req struct {
		Title  string `json:"title" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateConversationTitle(id, req.UserID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation title updated successfully"})
}

// DeleteConversation 删除对话
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	fmt.Println("[conversation_handler DeleteConversation] Start")
	id := c.Param("id")
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := h.service.DeleteConversation(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}
