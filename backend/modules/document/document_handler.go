package document

import (
	"fmt"
	"grandma/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	service *DocumentService
}

func NewDocumentHandler(service *DocumentService) *DocumentHandler {
	return &DocumentHandler{
		service: service,
	}
}

// GetDocumentList 获取文档列表（支持翻页）
func (h *DocumentHandler) GetDocumentList(c *gin.Context) {
	fmt.Println("[document_handler GetDocumentList] Start")
	var req models.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fmt.Printf("[document_handler GetDocumentList] Error: %+v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10 // 默认10个
	}

	response, err := h.service.GetDocumentList(req.ConversationID, req.BeforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDocumentByID 根据ID获取文档
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	fmt.Println("[document_handler GetDocumentByID] Start")
	id := c.Param("id")
	doc, err := h.service.GetDocumentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// UpdateDocument 更新文档
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	fmt.Println("[document_handler UpdateDocument] Start")
	var doc models.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateDocument(&doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document updated successfully"})
}

// DeleteDocument 删除文档
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	fmt.Println("[document_handler DeleteDocument] Start")
	id := c.Param("id")
	err := h.service.DeleteDocument(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// GetDocumentIDs 获取文档ID列表（用于翻页）
func (h *DocumentHandler) GetDocumentIDs(c *gin.Context) {
	var req models.DocumentIDsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	documentIDs, err := h.service.GetDocumentIDsByConversationID(req.ConversationID, req.BeforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.DocumentIDsResponse{
		DocumentIDs: documentIDs,
	})
}

// GetDocumentsByIDs 根据文档ID列表批量获取文档
func (h *DocumentHandler) GetDocumentsByIDs(c *gin.Context) {
	var req models.GetDocumentsByIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	documents, err := h.service.GetDocumentsByIDs(req.DocumentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.DocumentListResponse{
		Documents: documents,
		Total:     len(documents),
	})
}
