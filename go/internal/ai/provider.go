package ai

import (
	"context"

	"pengajian-backend/internal/model"
)

type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error)
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
