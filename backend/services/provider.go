package services

import (
	"fmt"
	"grandma/backend/models"
	"io"
)

type ChatProvider interface {
	ChatStream(messages []models.Message, writer io.Writer) error
}

func GetProvider(providerName, openaiKey, openaiURL, anthropicKey, anthropicURL string) (ChatProvider, error) {
	switch providerName {
	case "openai", "gpt-3.5-turbo", "gpt-4":
		if openaiKey == "" {
			return nil, fmt.Errorf("OpenAI API key is not configured")
		}
		return NewOpenAIProvider(openaiKey, openaiURL), nil
	case "anthropic", "claude":
		if anthropicKey == "" {
			return nil, fmt.Errorf("Anthropic API key is not configured")
		}
		return NewAnthropicProvider(anthropicKey, anthropicURL), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
