package ai

import (
	"time"
)

// Groq: provider OpenAI-compatible.
// Endpoint default: https://api.groq.com/openai/v1
func NewGroq(apiKey, model string, timeout time.Duration) Provider {
	if model == "" {
		model = "openai/gpt-oss-120b"
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "groq",
		BaseURL: "https://api.groq.com/openai/v1",
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
	})
}
