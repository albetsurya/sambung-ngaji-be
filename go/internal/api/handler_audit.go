package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetAuditLogs(c *fiber.Ctx, svc *service.AuditService) error {
	limit := int(BodyFloat(c, "limit"))
	if limit <= 0 {
		limit = 200
	}
	items, err := svc.GetAuditLogs(c.Context(),
		BodyString(c, "user_id"),
		BodyString(c, "target_type"),
		limit,
	)
	if err != nil {
		return Fail(c, "Gagal ambil audit logs: "+err.Error())
	}
	return Ok(c, items)
}
