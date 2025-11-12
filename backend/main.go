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
	"grandma/backend/modules/rag"
	"grandma/backend/modules/story"
	"grandma/backend/modules/work"
	"grandma/backend/repository"
	"grandma/backend/services"
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
	workRepo := repository.NewWorkRepository(database.DB)
	workDocumentRepo := repository.NewWorkDocumentRepository(database.DB)
	vectorChunkRepo := repository.NewVectorChunkRepository(database.DB)

	// 创建RAG服务
	var ragSvc *rag.RAGService
	if cfg.EnableRAG && cfg.OpenAIAPIKey != "" {
		embeddingSvc := services.NewEmbeddingService(cfg.EmbeddingAPIKey, cfg.EmbeddingBaseURL, cfg.EmbeddingModel)
		ragSvc = rag.NewRAGService(&rag.RAGConfig{
			Enabled:          true,
			EmbeddingService: embeddingSvc,
			VectorChunkRepo:  vectorChunkRepo,
			DocumentRepo:     documentRepo,
			WorkDocumentRepo: workDocumentRepo,
		})
		log.Println("RAG service initialized")
	} else {
		ragSvc = rag.NewRAGService(&rag.RAGConfig{Enabled: false})
		log.Println("RAG service disabled")
	}

	// 创建Services
	chatSvc := chatService.NewChatService(
		conversationRepo,
		documentRepo,
		workDocumentRepo,
		ragSvc,
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
	workSvc := work.NewWorkService(workRepo, workDocumentRepo)

	// 创建Handlers
	chatHdlr := chatHandler.NewChatHandler(chatSvc)
	conversationListHdlr := conversationListHandler.NewConversationListHandler(conversationListSvc)
	documentHdlr := documentHandler.NewDocumentHandler(documentSvc)
	conversationHdlr := conversationHandler.NewConversationHandler(conversationSvc)
	storiesHdlr := story.NewStoryHandler(storySvc)
	workHdlr := work.NewWorkHandler(workSvc)

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
		api.PUT("/stories/:id", storiesHdlr.UpdateStory)
		api.DELETE("/stories/:id", storiesHdlr.DeleteStory)

		// 创作模块
		api.GET("/works", workHdlr.GetWorkList)
		api.POST("/works", workHdlr.CreateWork)
		api.PUT("/works/:id/title", workHdlr.UpdateWorkTitle)
		api.DELETE("/works/:id", workHdlr.DeleteWork)

		// 创作文档模块
		api.GET("/works/:work_id/documents", workHdlr.GetWorkDocuments)
		api.POST("/works/:work_id/documents", workHdlr.CreateWorkDocument)
		api.GET("/work-documents/:id", workHdlr.GetWorkDocumentByID)
		api.PUT("/work-documents/:id/title", workHdlr.UpdateWorkDocumentTitle)
		api.PUT("/work-documents/:id/content", workHdlr.UpdateWorkDocumentContent)
		api.DELETE("/work-documents/:id", workHdlr.DeleteWorkDocument)

		// 获取可用模型列表
		api.GET("/models", func(c *gin.Context) {
			models := []map[string]string{
				{"id": "openai", "name": "DeepSeek Chat", "provider": "OpenAI"},
				{"id": "anthropic", "name": "Kimi", "provider": "Anthropic"},
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
