package rag

import (
	"fmt"
	"grandma/backend/models"
	"grandma/backend/repository"
	"grandma/backend/services"
	"grandma/backend/utils"
	"log"
	"strings"
	"time"
)

// RAGService RAG服务
type RAGService struct {
	embeddingService *services.EmbeddingService
	chunkingService  *services.ChunkingService
	vectorStore      *services.VectorStore
	vectorChunkRepo  *repository.VectorChunkRepository
	documentRepo     *repository.DocumentRepository
	workDocumentRepo *repository.WorkDocumentRepository
	enabled          bool
}

// RAGConfig RAG配置
type RAGConfig struct {
	Enabled          bool
	EmbeddingService *services.EmbeddingService
	VectorChunkRepo  *repository.VectorChunkRepository
	DocumentRepo     *repository.DocumentRepository
	WorkDocumentRepo *repository.WorkDocumentRepository
}

// NewRAGService 创建RAG服务
func NewRAGService(config *RAGConfig) *RAGService {
	if !config.Enabled {
		return &RAGService{enabled: false}
	}

	return &RAGService{
		embeddingService: config.EmbeddingService,
		chunkingService:  services.NewChunkingService(),
		vectorStore:      services.NewVectorStore(),
		vectorChunkRepo:  config.VectorChunkRepo,
		documentRepo:     config.DocumentRepo,
		workDocumentRepo: config.WorkDocumentRepo,
		enabled:          true,
	}
}

// IndexDocument 索引文档（异步处理）
func (r *RAGService) IndexDocument(documentID, userID, conversationID, workID, content, role string) error {
	if !r.enabled {
		return nil
	}

	// 异步处理，不阻塞
	go func() {
		if err := r.indexDocumentSync(documentID, userID, conversationID, workID, content, role); err != nil {
			log.Printf("Failed to index document %s: %v", documentID, err)
		}
	}()

	return nil
}

// indexDocumentSync 同步索引文档
func (r *RAGService) indexDocumentSync(documentID, userID, conversationID, workID, content, role string) error {
	// 删除旧的chunks
	if err := r.vectorChunkRepo.DeleteByDocumentID(documentID); err != nil {
		log.Printf("Failed to delete old chunks for document %s: %v", documentID, err)
	}

	// 如果内容为空或太短，不进行索引
	if len(content) < 50 {
		return nil
	}

	// 对内容进行切片
	chunks := r.chunkingService.ChunkText(content)
	if len(chunks) == 0 {
		return nil
	}

	// 批量获取embedding
	chunkTexts := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkTexts[i] = chunk.Content
	}

	embeddings, err := r.embeddingService.GetEmbeddings(chunkTexts)
	if err != nil {
		return fmt.Errorf("failed to get embeddings: %w", err)
	}

	// 创建vector chunks
	for i, chunk := range chunks {
		if i >= len(embeddings) {
			break
		}

		metadata := map[string]interface{}{
			"role":      role,
			"type":      chunk.Metadata["type"],
			"start_pos": chunk.StartPos,
			"end_pos":   chunk.EndPos,
		}

		vectorChunk := &models.VectorChunk{
			ID:             utils.GenerateID(),
			UserID:         userID,
			ConversationID: conversationID,
			WorkID:         workID,
			DocumentID:     documentID,
			Content:        chunk.Content,
		}

		if err := vectorChunk.SetEmbedding(embeddings[i]); err != nil {
			log.Printf("Failed to set embedding for chunk %d: %v", i, err)
			continue
		}

		if err := vectorChunk.SetMetadataMap(metadata); err != nil {
			log.Printf("Failed to set metadata for chunk %d: %v", i, err)
			continue
		}

		if err := r.vectorChunkRepo.Create(vectorChunk); err != nil {
			log.Printf("Failed to create vector chunk %d: %v", i, err)
			continue
		}
	}

	return nil
}

// RetrieveRelevantChunks 检索相关chunks
func (r *RAGService) RetrieveRelevantChunks(query string, userID, conversationID, workID string, topK int) ([]models.VectorChunk, error) {
	if !r.enabled {
		return nil, nil
	}

	if topK <= 0 {
		topK = 5
	}

	// 获取查询的embedding
	queryEmbedding, err := r.embeddingService.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// 获取所有相关的chunks（排除当前对话/创作）
	allChunks, err := r.vectorChunkRepo.GetAllWithEmbeddings(userID, conversationID, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}

	if len(allChunks) == 0 {
		return nil, nil
	}

	// 提取所有embeddings
	embeddings := make([][]float32, 0, len(allChunks))
	validChunks := make([]models.VectorChunk, 0, len(allChunks))

	for _, chunk := range allChunks {
		embedding, err := chunk.GetEmbedding()
		if err != nil || len(embedding) == 0 {
			continue
		}
		embeddings = append(embeddings, embedding)
		validChunks = append(validChunks, chunk)
	}

	if len(embeddings) == 0 {
		return nil, nil
	}

	// 搜索最相似的chunks
	indices, similarities, err := r.vectorStore.SearchSimilar(queryEmbedding, embeddings, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar: %w", err)
	}

	// 构建结果，应用时间衰减因子
	results := make([]models.VectorChunk, 0, len(indices))
	for i, idx := range indices {
		if idx >= len(validChunks) {
			continue
		}

		chunk := validChunks[idx]

		// 应用时间衰减：较新的内容权重更高
		timeDecay := r.calculateTimeDecay(chunk.CreatedAt, similarities[i])

		// 只返回相似度大于阈值的chunks
		if timeDecay > 0.3 { // 相似度阈值
			results = append(results, chunk)
		}
	}

	return results, nil
}

// calculateTimeDecay 计算时间衰减因子
// 较新的内容权重更高，但相似度仍然是主要因素
func (r *RAGService) calculateTimeDecay(createdAt time.Time, similarity float32) float32 {
	age := time.Since(createdAt).Hours()

	// 时间衰减：24小时内权重1.0，之后逐渐降低，30天后权重0.5
	decayFactor := 1.0
	if age > 24 {
		decayFactor = 1.0 - (age-24)/(30*24)*0.5
		if decayFactor < 0.5 {
			decayFactor = 0.5
		}
	}

	// 结合相似度和时间衰减
	return similarity * float32(decayFactor)
}

// BuildRAGContext 构建RAG增强的上下文消息
func (r *RAGService) BuildRAGContext(userMessage string, userID, conversationID, workID string) ([]models.Message, error) {
	if !r.enabled {
		return nil, nil
	}

	// 检索相关chunks
	chunks, err := r.RetrieveRelevantChunks(userMessage, userID, conversationID, workID, 8)
	if err != nil {
		log.Printf("RAG retrieval failed, falling back to default context: %v", err)
		return nil, nil
	}

	if len(chunks) == 0 {
		return nil, nil
	}

	// 构建上下文消息
	contextMessages := make([]models.Message, 0, len(chunks)+1)

	// 判断是否为灵感模式（长篇故事写作模式）
	isInspirationMode := workID != ""

	if isInspirationMode {
		// 灵感模式：针对长篇故事写作的优化prompt
		contextText := r.buildInspirationModeContext(chunks)
		contextMessages = append(contextMessages, models.Message{
			Role:    "system",
			Content: contextText,
		})
	} else {
		// 普通对话模式：使用简单的背景信息提示
		contextText := "以下是相关的历史背景信息，供参考：\n\n"
		for i, chunk := range chunks {
			metadata, _ := chunk.GetMetadataMap()
			role, _ := metadata["role"].(string)

			if role == "assistant" {
				contextText += fmt.Sprintf("[背景信息 %d]\n%s\n\n", i+1, chunk.Content)
			} else if role == "user" {
				contextText += fmt.Sprintf("[用户之前提到 %d]\n%s\n\n", i+1, chunk.Content)
			} else {
				contextText += fmt.Sprintf("[相关信息 %d]\n%s\n\n", i+1, chunk.Content)
			}
		}
		contextMessages = append(contextMessages, models.Message{
			Role:    "system",
			Content: contextText,
		})
	}

	return contextMessages, nil
}

// buildInspirationModeContext 构建灵感模式（长篇故事写作）的上下文
func (r *RAGService) buildInspirationModeContext(chunks []models.VectorChunk) string {
	var sb strings.Builder

	// 系统提示：强调长篇故事写作的要求
	sb.WriteString("你是一位专业的长篇故事创作助手。在创作时，请严格遵循以下要求：\n\n")
	sb.WriteString("## 核心创作原则\n")
	sb.WriteString("1. **保持一致性**：人物性格、外貌、行为方式必须与已有设定保持一致\n")
	sb.WriteString("2. **世界观统一**：遵循已建立的世界观、规则和设定，不要出现矛盾\n")
	sb.WriteString("3. **情节连贯**：新内容要与已有情节自然衔接，注意伏笔和线索的呼应\n")
	sb.WriteString("4. **风格统一**：保持整体文风和叙事风格的一致性\n\n")

	sb.WriteString("## 重要背景信息\n")
	sb.WriteString("以下是检索到的相关背景信息，请仔细参考并在创作中体现：\n\n")

	// 分类组织chunks：按角色和内容类型
	characterInfo := make([]models.VectorChunk, 0)
	worldInfo := make([]models.VectorChunk, 0)
	plotInfo := make([]models.VectorChunk, 0)
	userRequirements := make([]models.VectorChunk, 0)
	otherInfo := make([]models.VectorChunk, 0)

	for _, chunk := range chunks {
		metadata, _ := chunk.GetMetadataMap()
		role, _ := metadata["role"].(string)
		content := chunk.Content

		// 根据内容特征分类（简单启发式分类）
		contentLower := strings.ToLower(content)
		if strings.Contains(contentLower, "人物") || strings.Contains(contentLower, "角色") ||
			strings.Contains(contentLower, "性格") || strings.Contains(contentLower, "外貌") ||
			strings.Contains(contentLower, "名字") {
			if role == "assistant" {
				characterInfo = append(characterInfo, chunk)
			} else {
				userRequirements = append(userRequirements, chunk)
			}
		} else if strings.Contains(contentLower, "世界") || strings.Contains(contentLower, "设定") ||
			strings.Contains(contentLower, "规则") || strings.Contains(contentLower, "背景") {
			if role == "assistant" {
				worldInfo = append(worldInfo, chunk)
			} else {
				userRequirements = append(userRequirements, chunk)
			}
		} else if strings.Contains(contentLower, "情节") || strings.Contains(contentLower, "故事") ||
			strings.Contains(contentLower, "发生") || strings.Contains(contentLower, "事件") {
			if role == "assistant" {
				plotInfo = append(plotInfo, chunk)
			} else {
				userRequirements = append(userRequirements, chunk)
			}
		} else {
			if role == "user" {
				userRequirements = append(userRequirements, chunk)
			} else {
				otherInfo = append(otherInfo, chunk)
			}
		}
	}

	// 按优先级组织信息
	if len(characterInfo) > 0 {
		sb.WriteString("### 人物设定\n")
		for i, chunk := range characterInfo {
			sb.WriteString(fmt.Sprintf("**人物信息 %d：**\n%s\n\n", i+1, chunk.Content))
		}
	}

	if len(worldInfo) > 0 {
		sb.WriteString("### 世界观设定\n")
		for i, chunk := range worldInfo {
			sb.WriteString(fmt.Sprintf("**世界观信息 %d：**\n%s\n\n", i+1, chunk.Content))
		}
	}

	if len(plotInfo) > 0 {
		sb.WriteString("### 已有情节\n")
		for i, chunk := range plotInfo {
			sb.WriteString(fmt.Sprintf("**情节摘要 %d：**\n%s\n\n", i+1, chunk.Content))
		}
	}

	if len(userRequirements) > 0 {
		sb.WriteString("### 用户要求与设定\n")
		for i, chunk := range userRequirements {
			sb.WriteString(fmt.Sprintf("**用户要求 %d：**\n%s\n\n", i+1, chunk.Content))
		}
	}

	if len(otherInfo) > 0 {
		sb.WriteString("### 其他相关信息\n")
		for i, chunk := range otherInfo {
			sb.WriteString(fmt.Sprintf("**相关信息 %d：**\n%s\n\n", i+1, chunk.Content))
		}
	}

	sb.WriteString("## 创作要求\n")
	sb.WriteString("在创作新内容时：\n")
	sb.WriteString("- 必须严格遵循上述人物设定，不得改变已有角色的性格、外貌、能力等核心特征\n")
	sb.WriteString("- 必须遵循已建立的世界观和规则，不得出现逻辑矛盾\n")
	sb.WriteString("- 新情节必须与已有情节自然衔接，注意前后呼应\n")
	sb.WriteString("- 如果用户要求与已有设定冲突，请优先遵循已有设定，并在创作中巧妙处理冲突\n")
	sb.WriteString("- 保持文风一致，延续已有的叙事风格\n")
	sb.WriteString("- 注意细节的连贯性，如时间线、地点、人物关系等\n\n")

	sb.WriteString("请基于以上背景信息，创作符合要求的新内容。")

	return sb.String()
}
