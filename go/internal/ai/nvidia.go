package ai

import (
	"time"
)

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
