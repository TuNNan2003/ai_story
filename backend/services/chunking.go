package services

import (
	"strings"
)

// ChunkingService 文本切片服务
type ChunkingService struct {
	ChunkSize    int // 每个chunk的最大字符数
	ChunkOverlap int // chunk之间的重叠字符数
	MinChunkSize int // 最小chunk大小
}

// NewChunkingService 创建切片服务
func NewChunkingService() *ChunkingService {
	return &ChunkingService{
		ChunkSize:    1000, // 默认1000字符
		ChunkOverlap: 200,  // 默认200字符重叠
		MinChunkSize: 100,  // 最小100字符
	}
}

// Chunk 文本切片结果
type Chunk struct {
	Content  string            // chunk内容
	StartPos int               // 在原文本中的起始位置
	EndPos   int               // 在原文本中的结束位置
	Metadata map[string]string // 元数据
}

// ChunkByParagraphs 按段落切片（优先策略）
func (c *ChunkingService) ChunkByParagraphs(text string) []Chunk {
	if text == "" {
		return nil
	}

	paragraphs := strings.Split(text, "\n\n")
	chunks := make([]Chunk, 0)
	currentChunk := strings.Builder{}
	currentStart := 0
	pos := 0

	for i, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			pos += 2 // "\n\n"
			continue
		}

		paraWithNewlines := para
		if i > 0 {
			paraWithNewlines = "\n\n" + para
		}

		// 如果当前chunk加上新段落会超过大小限制
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paraWithNewlines) > c.ChunkSize {
			// 保存当前chunk
			chunkContent := currentChunk.String()
			if len(chunkContent) >= c.MinChunkSize {
				chunks = append(chunks, Chunk{
					Content:  chunkContent,
					StartPos: currentStart,
					EndPos:   pos,
					Metadata: map[string]string{"type": "paragraph"},
				})
			}

			// 开始新chunk，保留重叠部分
			overlapText := c.getOverlapText(chunkContent, c.ChunkOverlap)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			currentStart = pos - len(overlapText)
		}

		// 如果单个段落就超过大小限制，需要进一步分割
		if len(para) > c.ChunkSize {
			// 先保存当前chunk
			if currentChunk.Len() > 0 {
				chunkContent := currentChunk.String()
				if len(chunkContent) >= c.MinChunkSize {
					chunks = append(chunks, Chunk{
						Content:  chunkContent,
						StartPos: currentStart,
						EndPos:   pos,
						Metadata: map[string]string{"type": "paragraph"},
					})
				}
				currentChunk.Reset()
			}

			// 对长段落进行固定大小切片
			// para是段落文本，pos是段落在原文本中的位置
			// 需要将para的切片结果的位置信息调整为相对于原文本的位置
			subChunks := c.chunkByFixedSize(para, pos)
			chunks = append(chunks, subChunks...)
			pos += len(paraWithNewlines)
			currentStart = pos
			continue
		}

		// 添加段落到当前chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(paraWithNewlines)
		} else {
			currentChunk.WriteString(para)
			currentStart = pos
		}
		pos += len(paraWithNewlines)
	}

	// 保存最后一个chunk
	if currentChunk.Len() > 0 {
		chunkContent := currentChunk.String()
		if len(chunkContent) >= c.MinChunkSize {
			chunks = append(chunks, Chunk{
				Content:  chunkContent,
				StartPos: currentStart,
				EndPos:   pos,
				Metadata: map[string]string{"type": "paragraph"},
			})
		}
	}

	return chunks
}

// ChunkByFixedSize 按固定大小切片（备选策略）
func (c *ChunkingService) ChunkByFixedSize(text string) []Chunk {
	return c.chunkByFixedSize(text, 0)
}

// chunkByFixedSize 内部实现
// text: 要切片的文本
// startOffset: 文本在原文档中的起始位置（用于计算返回的StartPos和EndPos）
func (c *ChunkingService) chunkByFixedSize(text string, startOffset int) []Chunk {
	if text == "" {
		return nil
	}

	chunks := make([]Chunk, 0)
	textLen := len(text)
	textPos := 0 // 在text中的位置（从0开始）

	for textPos < textLen {
		endTextPos := textPos + c.ChunkSize
		if endTextPos > textLen {
			endTextPos = textLen
		}

		// 提取chunk文本
		chunkText := text[textPos:endTextPos]
		if len(chunkText) >= c.MinChunkSize {
			// 计算在原文档中的位置
			chunkStartPos := startOffset + textPos
			chunkEndPos := startOffset + endTextPos

			chunks = append(chunks, Chunk{
				Content:  chunkText,
				StartPos: chunkStartPos,
				EndPos:   chunkEndPos,
				Metadata: map[string]string{"type": "fixed_size"},
			})
		}

		// 移动到下一个chunk，考虑重叠
		nextTextPos := endTextPos - c.ChunkOverlap
		if nextTextPos <= textPos {
			// 如果重叠导致没有前进，至少前进1个字符，避免无限循环
			nextTextPos = textPos + 1
		}
		if nextTextPos < 0 {
			nextTextPos = 0
		}
		textPos = nextTextPos

		if textPos >= textLen {
			break
		}
	}

	return chunks
}

// getOverlapText 获取重叠文本（从chunk末尾提取）
func (c *ChunkingService) getOverlapText(text string, overlapSize int) string {
	if len(text) <= overlapSize {
		return text
	}
	return text[len(text)-overlapSize:]
}

// ChunkText 智能选择切片策略
func (c *ChunkingService) ChunkText(text string) []Chunk {
	if text == "" {
		return nil
	}

	// 优先使用段落切片
	chunks := c.ChunkByParagraphs(text)

	// 如果段落切片没有产生chunks，使用固定大小切片
	if len(chunks) == 0 {
		chunks = c.ChunkByFixedSize(text)
	}

	return chunks
}
