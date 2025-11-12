package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingService Embedding服务
type EmbeddingService struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewEmbeddingService 创建Embedding服务
func NewEmbeddingService(apiKey, baseURL, model string) *EmbeddingService {
	return &EmbeddingService{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}

// EmbeddingRequest Embedding API请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse Embedding API响应
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens   int `json:"total_tokens"`
	} `json:"usage"`
}

// GetEmbedding 获取单个文本的向量嵌入
func (e *EmbeddingService) GetEmbedding(text string) ([]float32, error) {
	embeddings, err := e.GetEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// GetEmbeddings 批量获取文本的向量嵌入
func (e *EmbeddingService) GetEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	url := fmt.Sprintf("%s/embeddings", e.BaseURL)

	payload := EmbeddingRequest{
		Model: e.Model,
		Input: texts,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 按索引排序，确保顺序正确
	embeddings := make([][]float32, len(result.Data))
	for _, item := range result.Data {
		if item.Index >= 0 && item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}

	return embeddings, nil
}

