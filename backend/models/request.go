package models

// ChatRequest 聊天请求
type ChatRequest struct {
	ConversationID string    `json:"conversation_id"` // 可选，如果为空则创建新对话
	Model          string    `json:"model" binding:"required"`
	UserID         string    `json:"user_id" binding:"required"` // 用户ID
	Messages       []Message `json:"messages" binding:"required"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ConversationListRequest 对话列表请求
type ConversationListRequest struct {
	UserID   string `json:"user_id" form:"user_id" binding:"required"` // 用户ID
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

// ConversationListResponse 对话列表响应
type ConversationListResponse struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
}

// DocumentListRequest 文档列表请求
type DocumentListRequest struct {
	UserID         string `json:"user_id" form:"user_id" binding:"required"` // 用户ID
	ConversationID string `json:"conversation_id" form:"conversation_id" binding:"required"`
	BeforeID       string `json:"before_id" form:"before_id"` // 用于翻页，获取比该ID更早的文档
	Limit          int    `json:"limit" form:"limit"`         // 返回数量，默认10
}

// DocumentIDsRequest 获取文档ID列表的请求
type DocumentIDsRequest struct {
	UserID         string `json:"user_id" form:"user_id" binding:"required"` // 用户ID
	ConversationID string `json:"conversation_id" form:"conversation_id" binding:"required"`
	BeforeID       string `json:"before_id" form:"before_id"` // 用于翻页，获取比该ID更早的文档ID
	Limit          int    `json:"limit" form:"limit"`         // 返回数量，默认10
}

// DocumentIDsResponse 文档ID列表响应
type DocumentIDsResponse struct {
	DocumentIDs []string `json:"document_ids"`
}

// DocumentListResponse 文档列表响应
type DocumentListResponse struct {
	Documents []Document `json:"documents"`
	Total     int        `json:"total"`
}

// CreateConversationWithTitleRequest 创建对话并生成标题的请求
type CreateConversationWithTitleRequest struct {
	UserID     string   `json:"user_id" binding:"required"` // 用户ID
	UserInputs []string `json:"user_inputs" binding:"required"`
}

// StoryRequest 故事列表请求
type StoryRequest struct {
	UserID      string `json:"user_id" form:"user_id" binding:"required"` // 用户ID
	Guid        string `json:"guid" form:"guid"`                          // guid可选，默认为"default"
	DocumentId  string `json:"document_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentHash string `json:"content_hash"`
}

// StoryResponse 故事列表响应
type StoryResponse struct {
	Story []Story `json:"stories"`
	Total int     `json:"total"`
}
