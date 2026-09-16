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
		Jam:     BodyString(c, "jam"),
		Acara:   BodyString(c, "acara"),
		GroupID: BodyString(c, "group_id"),
		Materi:  BodyString(c, "materi"),
		Catatan: BodyString(c, "catatan"),
	}

	// hari: bisa array string
	if v, ok := body["hari"].([]interface{}); ok {
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				p.Hari = append(p.Hari, s)
			}
		}
	}

	// kategori_target
	if v, ok := body["kategori_target"].([]interface{}); ok {
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				p.KategoriTarget = append(p.KategoriTarget, s)
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
	items, err := svc.GetTemplates(c.Context())
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}
