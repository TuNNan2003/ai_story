package conversation_list

import (
	"grandma/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConversationListHandler struct {
	service *ConversationListService
}

func NewConversationListHandler(service *ConversationListService) *ConversationListHandler {
	return &ConversationListHandler{
		service: service,
	}
}

// GetConversationList 获取对话列表
func (h *ConversationListHandler) GetConversationList(c *gin.Context) {
	var req models.ConversationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	response, err := h.service.GetConversationList(req.UserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateNewConversation 创建新对话
func (h *ConversationListHandler) CreateNewConversation(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.service.CreateNewConversation(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// CreateNewConversationWithTitle 创建新对话并生成标题
func (h *ConversationListHandler) CreateNewConversationWithTitle(c *gin.Context) {
	var req models.CreateConversationWithTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.service.CreateNewConversationWithTitle(req.UserID, req.UserInputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// GenerateTitle 智能命名接口
func (h *ConversationListHandler) GenerateTitle(c *gin.Context) {
	var req models.CreateConversationWithTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title, err := h.service.GenerateTitleForConversation(req.UserInputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"title": title})
}
