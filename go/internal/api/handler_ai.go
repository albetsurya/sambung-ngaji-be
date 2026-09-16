package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/service"
)

func handleAiChat(c *fiber.Ctx, svc *service.AIService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}

	body := BodyOf(c)
	req := model.ChatRequest{
		Message:  BodyString(c, "message"),
		Provider: BodyString(c, "provider"),
	}

	// history: bisa string JSON atau array
	if v, ok := body["history"].(string); ok && v != "" {
		var hist []model.ChatHistory
		if err := json.Unmarshal([]byte(v), &hist); err == nil {
			req.History = hist
		}
	} else if v, ok := body["history"].([]interface{}); ok {
		for _, raw := range v {
			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			text, _ := m["text"].(string)
			req.History = append(req.History, model.ChatHistory{Role: role, Text: text})
		}
	}

	res, err := svc.Chat(c.Context(), u, req)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetAiUsageStats(c *fiber.Ctx, svc *service.AIService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	res, err := svc.GetUsageStats(c.Context(), u)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
