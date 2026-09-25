package ai

import (
	"time"
)

// Groq: provider OpenAI-compatible, ada free tier.
// Endpoint default: https://api.groq.com/openai/v1
// Daftar key gratis: https://console.groq.com
func NewGroq(apiKey, model string, timeout time.Duration) Provider {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "groq",
		BaseURL: "https://api.groq.com/openai/v1",
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
	})
}
