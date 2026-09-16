package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetGroups(c *fiber.Ctx, svc *service.GroupService) error {
	includeInactive := BodyBool(c, "includeInactive")
	items, err := svc.GetGroups(c.Context(), includeInactive)
	if err != nil {
		return Fail(c, "Gagal ambil groups: "+err.Error())
	}
	return Ok(c, items)
}
