package ai

import (
	"time"
)

// Nvidia: provider OpenAI-compatible via NVIDIA NIM (build.nvidia.com).
// Base URL: https://integrate.api.nvidia.com/v1
// Daftar key gratis (tanpa kartu kredit): https://build.nvidia.com/settings
// Default openai/gpt-oss-20b: tool calling terverifikasi jalan.
func NewNvidia(apiKey, model string, timeout time.Duration) Provider {
	if model == "" {
		model = "openai/gpt-oss-20b"
	}
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "nvidia",
		BaseURL: "https://integrate.api.nvidia.com/v1",
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
	})
}
