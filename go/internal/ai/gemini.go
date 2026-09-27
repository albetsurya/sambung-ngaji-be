package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"pengajian-backend/internal/model"
)

type Gemini struct {
	apiKey string
	models []string
	client *http.Client
}

func NewGemini(apiKey string, models []string, timeout time.Duration) Provider {
	models = dedupModels(models)
	if len(models) == 0 {
		models = []string{"gemini-3.8-flash"}
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Gemini{
		apiKey: apiKey,
		models: models,
		client: &http.Client{Timeout: timeout},
	}
}

func (g *Gemini) Name() string { return "gemini" }

type gemPart struct {
	Text             string       `json:"text,omitempty"`
	FunctionCall     *gemFuncCall `json:"functionCall,omitempty"`
	FunctionResponse *gemFuncResp `json:"functionResponse,omitempty"`
}

type gemFuncCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

type gemFuncResp struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
}

type gemContent struct {
	Role  string    `json:"role,omitempty"`
	Parts []gemPart `json:"parts"`
}

type gemFuncDecl struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type gemTool struct {
	FunctionDeclarations []gemFuncDecl `json:"functionDeclarations"`
}

type gemRequest struct {
	SystemInstruction *gemContent   `json:"systemInstruction,omitempty"`
	Contents          []gemContent  `json:"contents"`
	Tools             []gemTool     `json:"tools,omitempty"`
	GenerationConfig  *gemGenConfig `json:"generationConfig,omitempty"`
}

type gemGenConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type gemResponse struct {
	Candidates []struct {
		Content      gemContent `json:"content"`
		FinishReason string     `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (g *Gemini) Chat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("gemini: API key belum diatur")
	}

	var result *LLMResult
	err := Retry(ctx, DefaultRetryConfig(), func() error {
		var err error
		result, err = g.doChat(ctx, messages, tools)
		return err
	})

	return result, err
}

func (g *Gemini) doChat(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef) (*LLMResult, error) {
	var lastErr error
	for _, m := range g.models {
		result, err := g.doChatOnce(ctx, messages, tools, m)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !IsFallbackable(err) {
			return nil, err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("gemini: tidak ada model yang dikonfigurasi")
	}
	return nil, lastErr
}

func (g *Gemini) doChatOnce(ctx context.Context, messages []model.LLMMessage, tools []model.LLMToolDef, genModel string) (*LLMResult, error) {
	req := gemRequest{
		GenerationConfig: &gemGenConfig{Temperature: 0.3},
	}
	var sysText strings.Builder

	toolCallIDToName := map[string]string{}

	for _, m := range messages {
		switch m.Role {
		case "system":
			if sysText.Len() > 0 {
				sysText.WriteString("\n\n")
			}
			sysText.WriteString(m.Content)

		case "user":
			req.Contents = append(req.Contents, gemContent{
				Role:  "user",
				Parts: []gemPart{{Text: m.Content}},
			})

		case "assistant":
			parts := []gemPart{}
			if m.Content != "" {
				parts = append(parts, gemPart{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var args map[string]interface{}
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				}
				parts = append(parts, gemPart{
					FunctionCall: &gemFuncCall{
						Name: tc.Function.Name,
						Args: args,
					},
				})
				toolCallIDToName[tc.ID] = tc.Function.Name
			}
			if len(parts) == 0 {
				continue
			}
			req.Contents = append(req.Contents, gemContent{
				Role:  "model",
				Parts: parts,
			})

		case "tool":
			name := toolCallIDToName[m.ToolCallID]
			if name == "" {
				name = m.ToolCallID
			}
			var resp interface{}
			if m.Content != "" {
				_ = json.Unmarshal([]byte(m.Content), &resp)
			}
			if resp == nil {
				resp = map[string]interface{}{}
			}
			req.Contents = append(req.Contents, gemContent{
				Role: "user",
				Parts: []gemPart{{
					FunctionResponse: &gemFuncResp{
						Name:     name,
						Response: resp,
					},
				}},
			})
		}
	}

	if sysText.Len() > 0 {
		req.SystemInstruction = &gemContent{
			Parts: []gemPart{{Text: sysText.String()}},
		}
	}

	if len(tools) > 0 {
		decls := make([]gemFuncDecl, 0, len(tools))
		for _, t := range tools {
			decls = append(decls, gemFuncDecl{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			})
		}
		req.Tools = []gemTool{{FunctionDeclarations: decls}}
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/" +
		genModel + ":generateContent?key=" + g.apiKey

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini http: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini HTTP %d: %s", resp.StatusCode, truncateStr(string(raw), 300))
	}

	var parsed gemResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("gemini parse: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini error %d: %s", parsed.Error.Code, parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: tidak ada candidate")
	}
	cand := parsed.Candidates[0]

	result := &LLMResult{
		FinishReason: cand.FinishReason,
		Model:        genModel,
		InputTokens:  parsed.UsageMetadata.PromptTokenCount,
		OutputTokens: parsed.UsageMetadata.CandidatesTokenCount,
		TotalTokens:  parsed.UsageMetadata.TotalTokenCount,
	}

	var textParts []string
	for _, p := range cand.Content.Parts {
		if p.Text != "" {
			textParts = append(textParts, p.Text)
		}
		if p.FunctionCall != nil {
			argsBytes, _ := json.Marshal(p.FunctionCall.Args)
			result.ToolCalls = append(result.ToolCalls, model.LLMToolCall{
				ID:   "gem_" + p.FunctionCall.Name,
				Type: "function",
				Function: model.LLMToolCallFunc{
					Name:      p.FunctionCall.Name,
					Arguments: string(argsBytes),
				},
			})
		}
	}
	result.Content = strings.Join(textParts, "\n")
	return result, nil
}
