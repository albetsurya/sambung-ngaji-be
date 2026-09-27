package model

import "time"

type ChatRequest struct {
	Message  string        `json:"message"`
	History  []ChatHistory `json:"history"`
	Provider string        `json:"provider"`
}

type ChatHistory struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type ChatResponse struct {
	Reply             string `json:"reply"`
	Provider          string `json:"provider,omitempty"`
	Model             string `json:"model,omitempty"`
	RequestedProvider string `json:"requestedProvider,omitempty"`
}

type LLMMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCalls  []LLMToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type LLMToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function LLMToolCallFunc `json:"function"`
}

type LLMToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMToolDef struct {
	Type     string         `json:"type"`
	Function LLMToolDefFunc `json:"function"`
}

type LLMToolDefFunc struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type AIUsageLog struct {
	UsageID      string
	UserID       *string
	UserNama     string
	Role         string
	Provider     string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Timestamp    time.Time
}

type AIUsageStats struct {
	Today      AIUsageSummary   `json:"today"`
	Month      AIUsageSummary   `json:"month"`
	ByProvider []AIUsageByField `json:"by_provider"`
	ByRole     []AIUsageByField `json:"by_role"`
	TopUsers   []AIUsageByUser  `json:"top_users"`
}

type AIUsageSummary struct {
	ChatCount   int `json:"chat_count"`
	TotalTokens int `json:"total_tokens"`
}

type AIUsageByField struct {
	Field       string `json:"provider,omitempty"`
	Role        string `json:"role,omitempty"`
	ChatCount   int    `json:"chat_count"`
	TotalTokens int    `json:"total_tokens"`
}

type AIUsageByUser struct {
	UserID      string `json:"user_id"`
	UserNama    string `json:"user_nama"`
	Role        string `json:"role"`
	ChatCount   int    `json:"chat_count"`
	TotalTokens int    `json:"total_tokens"`
}
