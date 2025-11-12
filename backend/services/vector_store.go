package services

import (
	"fmt"
	"math"
)

// VectorStore 向量存储服务（基于内存的简单实现）
// 注意：实际向量数据存储在数据库中，这里只提供相似度计算功能
type VectorStore struct {
}

// NewVectorStore 创建向量存储服务
func NewVectorStore() *VectorStore {
	return &VectorStore{}
}

// CosineSimilarity 计算两个向量的余弦相似度
func (v *VectorStore) CosineSimilarity(vec1, vec2 []float32) (float32, error) {
	if len(vec1) != len(vec2) {
		return 0, fmt.Errorf("vectors must have the same length")
	}

	var dotProduct float32
	var norm1, norm2 float32

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0, nil
	}

	similarity := dotProduct / (float32(math.Sqrt(float64(norm1))) * float32(math.Sqrt(float64(norm2))))
	return similarity, nil
}

// SearchSimilar 在向量列表中搜索最相似的向量
// 返回相似度从高到低排序的索引和相似度分数
func (v *VectorStore) SearchSimilar(queryVec []float32, vectors [][]float32, topK int) ([]int, []float32, error) {
	if len(vectors) == 0 {
		return nil, nil, nil
	}

	type scorePair struct {
		index    int
		similarity float32
	}

	scores := make([]scorePair, 0, len(vectors))

	for i, vec := range vectors {
		similarity, err := v.CosineSimilarity(queryVec, vec)
		if err != nil {
			continue
		}
		scores = append(scores, scorePair{
			index:      i,
			similarity: similarity,
		})
	}

	// 按相似度降序排序
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].similarity < scores[j].similarity {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 返回topK个结果
	if topK > len(scores) {
		topK = len(scores)
	}

	indices := make([]int, topK)
	similarities := make([]float32, topK)
	for i := 0; i < topK; i++ {
		indices[i] = scores[i].index
		similarities[i] = scores[i].similarity
	}

	return indices, similarities, nil
}

