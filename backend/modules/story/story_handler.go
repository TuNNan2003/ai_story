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
	// guid参数可选，如果不提供则默认为"default"
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果绑定失败，尝试使用默认值
		req.Guid = "default"
	}

	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	response, err := h.service.GetStoryList(req.UserID, req.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateStory 创建新故事
func (h *StoryHandler) CreateStory(c *gin.Context) {
	var req models.StoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[Story_handler CreateStory] Error: %+v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必需字段
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	story, err := h.service.CreateStory(req.UserID, req.Guid, req.DocumentId, req.Title, req.Content, req.ContentHash)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "duplicate_story" {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate_story"})
			return
		}
		if err.Error() == "hash_mismatch" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hash_mismatch"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, story)
}

// DeleteStory 删除故事
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	fmt.Println("[story_handler DeleteStory] Start")
	id := c.Param("id")
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := h.service.DeleteStory(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Story deleted successfully"})
}

// UpdateStory 更新故事
func (h *StoryHandler) UpdateStory(c *gin.Context) {
	fmt.Println("[story_handler UpdateStory] Start")
	id := c.Param("id")

	var req models.StoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[Story_handler UpdateStory] Error: %+v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必需字段
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	story, err := h.service.UpdateStory(id, req.UserID, req.Title, req.Content, req.ContentHash)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "duplicate_story" {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate_story"})
			return
		}
		if err.Error() == "hash_mismatch" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hash_mismatch"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, story)
}
