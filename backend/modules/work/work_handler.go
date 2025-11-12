package work

import (
	"grandma/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WorkHandler 创作处理器
type WorkHandler struct {
	service *WorkService
}

// NewWorkHandler 创建创作处理器
func NewWorkHandler(service *WorkService) *WorkHandler {
	return &WorkHandler{
		service: service,
	}
}

// GetWorkList 获取创作列表
func (h *WorkHandler) GetWorkList(c *gin.Context) {
	var req models.WorkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.GetWorkList(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateWork 创建创作
func (h *WorkHandler) CreateWork(c *gin.Context) {
	var req models.WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	work, err := h.service.CreateWork(req.UserID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, work)
}

// UpdateWorkTitle 更新创作标题
func (h *WorkHandler) UpdateWorkTitle(c *gin.Context) {
	id := c.Param("id")
	var req models.WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateWorkTitle(id, req.UserID, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteWork 删除创作
func (h *WorkHandler) DeleteWork(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.service.DeleteWork(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetWorkDocuments 获取创作的所有文档
func (h *WorkHandler) GetWorkDocuments(c *gin.Context) {
	workID := c.Param("work_id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	response, err := h.service.GetWorkDocuments(workID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateWorkDocument 创建创作文档
func (h *WorkHandler) CreateWorkDocument(c *gin.Context) {
	var req models.WorkDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.service.CreateWorkDocument(req.UserID, req.WorkID, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// UpdateWorkDocumentTitle 更新文档标题
func (h *WorkHandler) UpdateWorkDocumentTitle(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateWorkDocumentTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateWorkDocumentTitle(id, req.UserID, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// UpdateWorkDocumentContent 更新文档内容
func (h *WorkHandler) UpdateWorkDocumentContent(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateWorkDocumentContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateWorkDocumentContent(id, req.UserID, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteWorkDocument 删除文档
func (h *WorkHandler) DeleteWorkDocument(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.service.DeleteWorkDocument(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetWorkDocumentByID 根据ID获取文档
func (h *WorkHandler) GetWorkDocumentByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	doc, err := h.service.GetWorkDocumentByID(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

