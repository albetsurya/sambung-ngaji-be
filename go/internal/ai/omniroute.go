package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"pengajian-backend/internal/model"
)

type OmniRouteConfig struct {
	Endpoint string
	APIKey   string
	Model    string
	Timeout  time.Duration
}

type OmniRoute struct {
	cfg    OmniRouteConfig
	client *http.Client
}

func NewOmniRoute(cfg OmniRouteConfig) *OmniRoute {
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &OmniRoute{
		cfg: cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

func (o *OmniRoute) Name() string { return "omniroute" }

// ===== Wire format =====

type orChatRequest struct {
	Model       string             `json:"model"`
	Messages    []model.LLMMessage `json:"messages"`
	Tools       []model.LLMToolDef `json:"tools,omitempty"`
	ToolChoice  string             `json:"tool_choice,omitempty"`
	Temperature float64            `json:"temperature"`
}

type orChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int               `json:"index"`
		FinishReason string            `json:"finish_reason"`
		Message      model.LLMMessage  `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type LLMResult struct {
	Content      string
	Model        string
	ToolCalls    []model.LLMToolCall
	FinishReason string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// Chat: panggil /chat/completions. tools boleh nil.
func (o *OmniRoute) Chat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error) {
	if o.cfg.Endpoint == "" {
		return nil, errors.New("OMNIROUTE_ENDPOINT belum diatur")
	}

	reqBody := orChatRequest{
		Model:       o.cfg.Model,
		Messages:    messages,
		Temperature: 0.3,
	}
	if len(tools) > 0 {
		reqBody.Tools = tools
		reqBody.ToolChoice = "auto"
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := o.cfg.Endpoint + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ngrok-skip-browser-warning", "true")
	if o.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("omniroute HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	// Deteksi HTML (ngrok warning)
	if len(raw) > 0 && raw[0] == '<' {
		return nil, errors.New("omniroute mengembalikan HTML (ngrok warning?)")
	}

	var parsed orChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("omniroute tidak mengembalikan choice")
	}
	choice := parsed.Choices[0]
	return &LLMResult{
		Content:      choice.Message.Content,
		Model:        parsed.Model,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
		TotalTokens:  parsed.Usage.TotalTokens,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
