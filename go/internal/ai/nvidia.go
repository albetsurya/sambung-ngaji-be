package ai

import (
	"time"
)

// Nvidia: provider OpenAI-compatible via NVIDIA NIM (build.nvidia.com).
// Base URL: https://integrate.api.nvidia.com/v1
// Daftar key gratis (tanpa kartu kredit): https://build.nvidia.com/settings
// Default openai/gpt-oss-20b: tool calling terverifikasi jalan.
func NewNvidia(apiKey string, models []string, timeout time.Duration) Provider {
	models = dedupModels(models)
	if len(models) == 0 {
		models = []string{"openai/gpt-oss-20b"}
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "nvidia",
		BaseURL: "https://integrate.api.nvidia.com/v1",
		APIKey:  apiKey,
		Models:  models,
		Timeout: timeout,
	})
}
