RAG优化实现方案
背景分析
当前系统在 chat_service.go 中只加载最近20条消息作为上下文（第187行），这对于长篇故事写作来说存在以下问题：

无法有效利用早期的重要信息（人物设定、世界观、伏笔等）
上下文长度受限，可能丢失关键背景信息
无法根据当前写作需求智能检索相关记忆
实现方案
1. 技术选型
向量数据库：

选项A：使用 ChromaDB（推荐）- 轻量级、支持Go客户端、易于部署
选项B：使用 SQLite + 向量扩展 - 与现有数据库集成，但需要额外配置
选项C：使用 FAISS（通过CGO） - 性能好但集成复杂
Embedding模型：

使用 OpenAI text-embedding-3-small 或 text-embedding-ada-002
通过HTTP API调用，无需本地模型
Chunking策略：

按段落切片（优先）：以换行符为分隔，每段作为一个chunk
固定字符数切片（备选）：每500-1000字符一个chunk，保留重叠
智能切片（未来优化）：识别章节、对话、描述等结构
2. 数据模型设计
新增数据表 vector_chunks：

type VectorChunk struct {
    ID             string    // chunk唯一ID
    UserID         string    // 用户ID（数据隔离）
    ConversationID string    // 所属对话ID（可选，用于对话模式）
    WorkID         string    // 所属创作ID（可选，用于灵感模式）
    DocumentID     string    // 来源文档ID
    Content        string    // chunk文本内容
    Embedding      []float32 // 向量嵌入（JSON存储或单独表）
    Metadata       string    // JSON格式元数据（角色、位置、类型等）
    CreatedAt      time.Time
}
3. 核心模块设计
3.1 Embedding服务 (services/embedding.go)
封装OpenAI Embedding API调用
提供文本向量化接口
支持批量处理
3.2 向量存储服务 (services/vector_store.go)
封装ChromaDB操作
提供chunk的增删改查
实现相似度搜索接口
3.3 Chunking服务 (services/chunking.go)
实现文本切片逻辑
支持多种切片策略
处理重叠和边界情况
3.4 RAG服务 (modules/rag/rag_service.go)
整合Embedding、Chunking、VectorStore
提供统一的RAG检索接口
管理chunk的生命周期（创建、更新、删除）
4. 集成点修改
4.1 文档保存时自动索引
在 chat_service.go 的 responseCollector.Write() 方法中：

当文档内容达到一定长度时（如500字符），触发chunking和索引
异步处理，不阻塞流式响应
4.2 上下文构建时使用RAG检索
修改 chat_service.go 的 sendMessageForConversation() 和 sendMessageForWork() 方法：

保留最近3-5条消息作为直接上下文（保证对话连贯性）
使用RAG检索相关历史chunks（5-10个最相关的）
将检索到的chunks作为系统提示或上下文消息插入
5. 实现步骤
依赖添加：在 go.mod 中添加ChromaDB Go客户端和相关依赖
数据模型：创建 VectorChunk 模型并添加数据库迁移
服务层：实现Embedding、Chunking、VectorStore服务
RAG模块：创建 modules/rag 模块，整合各服务
集成修改：修改 chat_service.go 集成RAG检索
配置管理：在 config.go 中添加Embedding API配置
测试验证：确保RAG检索不影响现有功能
6. 关键实现细节
检索策略：

使用用户当前消息的embedding进行相似度搜索
结合时间衰减因子（较新的内容权重更高）
过滤掉当前对话中已包含的文档，避免重复
性能优化：

异步索引：文档保存后异步进行chunking和embedding
缓存机制：对频繁检索的query结果进行缓存
批量处理：批量进行embedding计算
错误处理：

Embedding API失败时降级到传统上下文加载
VectorStore不可用时使用最近N条消息策略
文件修改清单
新增文件
backend/models/vector_chunk.go - 向量chunk数据模型
backend/services/embedding.go - Embedding服务
backend/services/chunking.go - 文本切片服务
backend/services/vector_store.go - 向量存储服务
backend/modules/rag/rag_service.go - RAG核心服务
backend/modules/rag/rag_handler.go - RAG处理器（可选，用于管理接口）
backend/repository/vector_chunk_repo.go - 向量chunk数据访问层
修改文件
backend/go.mod - 添加依赖
backend/config/config.go - 添加Embedding API配置
backend/database/db.go - 添加VectorChunk表迁移
backend/modules/chat/chat_service.go - 集成RAG检索
backend/main.go - 初始化RAG服务
注意事项
向后兼容：确保RAG功能失败时能降级到现有逻辑
数据隔离：所有操作必须基于UserID进行过滤
成本控制：Embedding API调用会产生费用，需要合理控制调用频率
性能影响：RAG检索会增加延迟，需要优化和缓存
测试覆盖：确保新功能不影响现有对话和灵感模式