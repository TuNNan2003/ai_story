package story

import (
	"fmt"
	"grandma/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StoryHandler struct {
	service *StoryService
}

func NewStoryHandler(service *StoryService) *StoryHandler {
	return &StoryHandler{
		service: service,
	}
}

// GetStoryList 获取文档列表
func (h *StoryHandler) GetStoryList(c *gin.Context) {
	fmt.Println("[story_handler GetStoryList] Start")
	var req models.StoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fmt.Printf("[Story_handler GetStoryList] Error: %+v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.GetStoryList(req.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateStory 创建新对话
func (h *StoryHandler) CreateStory(c *gin.Context) {
	var req models.StoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[Story_handler CreateStory] Error: %+v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	story, err := h.service.CreateStory(req.Guid, req.DocumentId, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, story)
}

// DeleteStory 删除文档
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	fmt.Println("[story_handler DeleteStory] Start")
	id := c.Param("id")
	err := h.service.DeleteStory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Story deleted successfully"})
}
