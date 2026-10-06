package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func parseBulkParams(c *fiber.Ctx) service.BulkParams {
	body := BodyOf(c)

	p := service.BulkParams{
		Tahun:   int(BodyFloat(c, "tahun")),
		Bulan:   int(BodyFloat(c, "bulan")),
		Time:    BodyString(c, "time"),
		Event:   BodyString(c, "event"),
		GroupID: BodyString(c, "group_id"),
		Topic:   BodyString(c, "topic"),
		Notes:   BodyString(c, "notes"),
	}

	if v, ok := body["day"].([]interface{}); ok {
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				p.Day = append(p.Day, s)
			}
		}
	}

	if v, ok := body["target_categories"].([]interface{}); ok {
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				p.TargetCategories = append(p.TargetCategories, s)
			}
		}
	}

	return p
}

func handlePreviewBulkMeetings(c *fiber.Ctx, svc *service.BulkMeetingService) error {
	p := parseBulkParams(c)
	res, err := svc.Preview(c.Context(), p)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleBulkCreateMeetings(c *fiber.Ctx, svc *service.BulkMeetingService) error {
	u := UserOf(c)
	userID := ""
	if u != nil {
		userID = u.UserID
	}
	p := parseBulkParams(c)
	res, err := svc.BulkCreate(c.Context(), p, userID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetBulkMeetingTemplates(c *fiber.Ctx, svc *service.BulkMeetingService) error {
	groupID := BodyString(c, "group_id")
	items, err := svc.GetTemplates(c.Context(), groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}
