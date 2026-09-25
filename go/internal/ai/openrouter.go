package ai

import (
	"time"
)

// OpenRouter: agregator OpenAI-compatible dengan banyak model gratis suffix ":free".
// Endpoint: https://openrouter.ai/api/v1
// Daftar key gratis: https://openrouter.ai (tanpa kartu kredit untuk model :free)
// Contoh model gratis: meta-llama/llama-3.3-70b-instruct:free, qwen/qwen3-32b:free
func NewOpenRouter(apiKey, model string, timeout time.Duration) Provider {
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "openrouter",
		BaseURL: "https://openrouter.ai/api/v1",
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
		ExtraHeaders: map[string]string{
			"HTTP-Referer": "https://sambung-ngaji.vercel.app",
			"X-Title":      "SambungNgaji",
		},
	})
}
