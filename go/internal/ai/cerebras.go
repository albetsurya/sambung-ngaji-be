package ai

import (
	"time"
)

// Cerebras: provider OpenAI-compatible, ada free tier.
// Endpoint: https://api.cerebras.ai/v1
// Daftar key gratis: https://cloud.cerebras.ai
func NewCerebras(apiKey, model string, timeout time.Duration) Provider {
	if model == "" {
		model = "gpt-oss-120b"
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "cerebras",
		BaseURL: "https://api.cerebras.ai/v1",
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
	})
}
