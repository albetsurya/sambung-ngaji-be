package ai

import (
	"context"

	"pengajian-backend/internal/model"
)

// Provider: abstraksi LLM provider.
type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error)
}

// LLMResult: hasil seragam dari semua provider.
type LLMResult struct {
	Content      string
	Model        string
	ToolCalls    []model.LLMToolCall
	FinishReason string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}
