package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"pengajian-backend/internal/model"
)

type OpenAICompat struct {
	name         string
	baseURL      string
	apiKey       string
	model        string
	client       *http.Client
	extraHeaders map[string]string
}

type OpenAICompatConfig struct {
	Name         string
	BaseURL      string
	APIKey       string
	Model        string
	Timeout      time.Duration
	ExtraHeaders map[string]string
}

func NewOpenAICompat(cfg OpenAICompatConfig) *OpenAICompat {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.ExtraHeaders == nil {
		cfg.ExtraHeaders = map[string]string{}
	}
	return &OpenAICompat{
		name:         cfg.Name,
		baseURL:      cfg.BaseURL,
		apiKey:       cfg.APIKey,
		model:        cfg.Model,
		client:       &http.Client{Timeout: cfg.Timeout},
		extraHeaders: cfg.ExtraHeaders,
	}
}

type openAIChatRequest struct {
	Model       string             `json:"model"`
	Messages    []model.LLMMessage `json:"messages"`
	Tools       []model.LLMToolDef `json:"tools,omitempty"`
	ToolChoice  string             `json:"tool_choice,omitempty"`
	Temperature float64            `json:"temperature"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int              `json:"index"`
		FinishReason string           `json:"finish_reason"`
		Message      model.LLMMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (o *OpenAICompat) Name() string { return o.name }

func (o *OpenAICompat) Chat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error) {
	var result *LLMResult
	err := Retry(ctx, DefaultRetryConfig(), func() error {
		var err error
		result, err = o.doChat(ctx, messages, tools)
		return err
	})
	return result, err
}

func (o *OpenAICompat) doChat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error) {
	if o.baseURL == "" {
		return nil, fmt.Errorf("%s: base URL belum diatur", o.name)
	}
	if o.apiKey == "" && o.name != "omniroute" {
		return nil, fmt.Errorf("%s: API key belum diatur", o.name)
	}

	body := openAIChatRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: 0.3,
	}
	if len(tools) > 0 {
		body.Tools = tools
		body.ToolChoice = "auto"
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := o.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	for k, v := range o.extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s http: %w", o.name, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s HTTP %d: %s", o.name, resp.StatusCode, truncateStr(string(raw), 200))
	}
	if len(raw) > 0 && raw[0] == '<' {
		return nil, fmt.Errorf("%s mengembalikan HTML", o.name)
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("%s parse: %w", o.name, err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("%s: tidak ada choice", o.name)
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

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// IsQuotaError: cek error 429/rate limit untuk trigger fallback.
func IsQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, kw := range []string{"429", "resource_exhausted", "quota", "rate limit", "rate_limit"} {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

var _ = errors.New
