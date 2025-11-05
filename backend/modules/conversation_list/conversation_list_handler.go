package conversation_list

import (
	"grandma/backend/models"
	"net/http"
	"strconv"

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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.service.GetConversationList(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateNewConversation 创建新对话
func (h *ConversationListHandler) CreateNewConversation(c *gin.Context) {
	conversation, err := h.service.CreateNewConversation()
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

	conversation, err := h.service.CreateNewConversationWithTitle(req.UserInputs)
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
