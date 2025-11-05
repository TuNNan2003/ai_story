package main

import (
	"grandma/backend/config"
	"grandma/backend/database"
	chatHandler "grandma/backend/modules/chat"
	chatService "grandma/backend/modules/chat"
	conversationHandler "grandma/backend/modules/conversation"
	conversationService "grandma/backend/modules/conversation"
	conversationListHandler "grandma/backend/modules/conversation_list"
	conversationListService "grandma/backend/modules/conversation_list"
	documentHandler "grandma/backend/modules/document"
	documentService "grandma/backend/modules/document"
	"grandma/backend/modules/story"
	"grandma/backend/repository"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	err = database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建Repository
	conversationRepo := repository.NewConversationRepository(database.DB)
	documentRepo := repository.NewDocumentRepository(database.DB)
	storyRepo := repository.NewStoryRepository(database.DB)

	// 创建Services
	chatSvc := chatService.NewChatService(
		conversationRepo,
		documentRepo,
		&chatService.ChatConfig{
			OpenAIAPIKey:     cfg.OpenAIAPIKey,
			OpenAIBaseURL:    cfg.OpenAIBaseURL,
			AnthropicAPIKey:  cfg.AnthropicAPIKey,
			AnthropicBaseURL: cfg.AnthropicBaseURL,
		},
	)
	conversationListSvc := conversationListService.NewConversationListService(
		conversationRepo,
		&conversationListService.TitleGenerationConfig{
			OpenAIAPIKey:     cfg.OpenAIAPIKey,
			OpenAIBaseURL:    cfg.OpenAIBaseURL,
			AnthropicAPIKey:  cfg.AnthropicAPIKey,
			AnthropicBaseURL: cfg.AnthropicBaseURL,
			DefaultModel:     "openai", // 默认使用openai生成标题
		},
	)
	documentSvc := documentService.NewDocumentService(documentRepo)
	conversationSvc := conversationService.NewConversationService(conversationRepo, documentRepo)
	storySvc := story.NewStoryService(storyRepo)

	// 创建Handlers
	chatHdlr := chatHandler.NewChatHandler(chatSvc)
	conversationListHdlr := conversationListHandler.NewConversationListHandler(conversationListSvc)
	documentHdlr := documentHandler.NewDocumentHandler(documentSvc)
	conversationHdlr := conversationHandler.NewConversationHandler(conversationSvc)
	storiesHdlr := story.NewStoryHandler(storySvc)

	// 配置路由 - 对话模块
	api := r.Group("/api")
	{
		// 聊天接口
		api.POST("/chat", chatHdlr.Chat)

		// 对话列表模块
		api.GET("/conversations", conversationListHdlr.GetConversationList)
		api.POST("/conversations/new", conversationListHdlr.CreateNewConversation)
		api.POST("/conversations/new-with-title", conversationListHdlr.CreateNewConversationWithTitle)
		api.POST("/conversations/generate-title", conversationListHdlr.GenerateTitle)

		// 对话管理模块
		api.GET("/conversations/:id", conversationHdlr.GetConversationByID)
		api.POST("/conversations", conversationHdlr.CreateConversation)
		api.PUT("/conversations/:id", conversationHdlr.UpdateConversation)
		api.PUT("/conversations/:id/title", conversationHdlr.UpdateConversationTitle)
		api.DELETE("/conversations/:id", conversationHdlr.DeleteConversation)

		// 文档管理模块
		api.GET("/documents", documentHdlr.GetDocumentList)
		api.GET("/documents/ids", documentHdlr.GetDocumentIDs)
		api.GET("/documents/:id", documentHdlr.GetDocumentByID)
		api.PUT("/documents/:id", documentHdlr.UpdateDocument)
		api.DELETE("/documents/:id", documentHdlr.DeleteDocument)

		api.GET("/stories", storiesHdlr.GetStoryList)
		api.POST("/stories", storiesHdlr.CreateStory)
		api.DELETE("/stories/:id", storiesHdlr.DeleteStory)

		// 获取可用模型列表
		api.GET("/models", func(c *gin.Context) {
			models := []map[string]string{
				{"id": "openai", "name": "GPT-3.5 Turbo", "provider": "OpenAI"},
				{"id": "anthropic", "name": "Claude 3.5 Sonnet", "provider": "Anthropic"},
			}
			c.JSON(200, gin.H{"models": models})
		})
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
