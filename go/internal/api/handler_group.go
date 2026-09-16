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

func handleSaveGroup(c *fiber.Ctx, svc *service.GroupService) error {
	body := BodyOf(c)
	in := service.SaveGroupInput{
		GroupID:       BodyString(c, "group_id"),
		GroupCode:     BodyString(c, "group_code"),
		GroupName:     BodyString(c, "group_name"),
		Pembina:       BodyString(c, "pembina"),
		Penandatangan: BodyString(c, "penandatangan"),
		Jadwal:        BodyString(c, "jadwal"),
	}
	if v, ok := body["status_aktif"].(bool); ok {
		in.StatusAktif = &v
	}
	dto, err := svc.Save(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}
