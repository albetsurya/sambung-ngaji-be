package ai

import (
	"time"
)

// Groq: provider OpenAI-compatible, ada free tier.
// Endpoint default: https://api.groq.com/openai/v1
// Daftar key gratis: https://console.groq.com
func NewGroq(apiKey string, models []string, timeout time.Duration) Provider {
	models = dedupModels(models)
	if len(models) == 0 {
		models = []string{"openai/gpt-oss-120b", "openai/gpt-oss-20b"}
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "groq",
		BaseURL: "https://api.groq.com/openai/v1",
		APIKey:  apiKey,
		Models:  models,
		Timeout: timeout,
	})
}
