package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"grandma/backend/models"
	"io"
	"log"
	"net/http"
)

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

func (p *OpenAIProvider) ChatStream(messages []models.Message, writer io.Writer) error {
	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)

	// 将消息数组转换为API格式
	apiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		apiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	payload := map[string]interface{}{
		"model":    "deepseek-chat",
		"messages": apiMessages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	log.Printf("[OpenAiProvider ChatStream] payload: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai api error: %s", string(body))
	}

	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			data := string(buffer[:n])
			lines := parseSSE(data)
			for _, line := range lines {
				if line == "" || line == "[DONE]" {
					continue
				}

				var streamResp struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
					continue
				}

				if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
					_, _ = writer.Write([]byte(streamResp.Choices[0].Delta.Content))
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func parseSSE(data string) []string {
	var lines []string
	current := ""

	for _, char := range data {
		if char == '\n' {
			if current != "" {
				if len(current) > 6 && current[:6] == "data: " {
					lines = append(lines, current[6:])
				}
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" && len(current) > 6 && current[:6] == "data: " {
		lines = append(lines, current[6:])
	}

	return lines
}

// Chat 非流式聊天，用于生成标题等场景
func (p *OpenAIProvider) Chat(messages []models.Message) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)

	// 将消息数组转换为API格式
	apiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		apiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	payload := map[string]interface{}{
		"model":    "deepseek-chat",
		"messages": apiMessages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai api error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}
