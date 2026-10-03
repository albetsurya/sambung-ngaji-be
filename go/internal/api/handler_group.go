package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/service"
)

func handleGetGroups(c *fiber.Ctx, svc *service.GroupService) error {
	includeInactive := BodyBool(c, "includeInactive")
	items, err := svc.GetGroups(c.Context(), includeInactive)
	if err != nil {
		return Fail(c, "Gagal ambil groups: "+err.Error())
	}
	/* Akun ber-group_label hanya melihat kelompoknya sendiri.
	   SUPER_ADMIN / akun global tetap melihat semua (perilaku lama). */
	if groupID, isSuper := ActorOf(c); !isSuper && groupID != "" && groupID != UnassignedGroup {
		filtered := make([]model.GroupDTO, 0, 1)
		for _, g := range items {
			if g.GroupID == groupID {
				filtered = append(filtered, g)
			}
		}
		items = filtered
	}
	return Ok(c, items)
}

func handleSaveGroup(c *fiber.Ctx, svc *service.GroupService) error {
	body := BodyOf(c)
	in := service.SaveGroupInput{
		GroupID:       BodyString(c, "group_id"),
		GroupCode:     BodyString(c, "group_code"),
		GroupName:     BodyString(c, "group_name"),
		Mentor:       BodyString(c, "mentor"),
		Signatory: BodyString(c, "signatory"),
		Schedule:        BodyString(c, "schedule"),
	}
	if v, ok := body["is_active"].(bool); ok {
		in.IsActive = &v
	}
	dto, err := svc.Save(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGetPublicGroups(c *fiber.Ctx, svc *service.GroupService) error {
	items, err := svc.GetGroups(c.Context(), false)
	if err != nil {
		return Fail(c, "Gagal ambil kelompok: "+err.Error())
	}
	return Ok(c, items)
}
