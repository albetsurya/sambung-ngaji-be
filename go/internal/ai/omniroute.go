package ai

import (
	"time"
)

// OmniRoute: wrapper tipis pakai OpenAICompat.
// Endpoint ngrok butuh header ngrok-skip-browser-warning.
func NewOmniRoute(endpoint, apiKey, model string, timeout time.Duration) Provider {
	return NewOpenAICompat(OpenAICompatConfig{
		Name:    "omniroute",
		BaseURL: endpoint,
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
		ExtraHeaders: map[string]string{
			"ngrok-skip-browser-warning": "true",
		},
	})
}
