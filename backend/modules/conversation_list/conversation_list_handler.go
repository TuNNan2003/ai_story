package conversation_list

import (
	"fmt"
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
	fmt.Println("[conversation_list_handler GetConversationList] Start")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.service.GetConversationList(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
