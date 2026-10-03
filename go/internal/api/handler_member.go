package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/service"
)

func parseMemberFilter(c *fiber.Ctx) model.MemberListFilter {
	return model.MemberListFilter{
		Search:          BodyString(c, "search"),
		GroupLabel:        BodyString(c, "group_label"),
		Gender:    BodyString(c, "gender"),
		Village:            BodyString(c, "village"),
		Kategori:        BodyString(c, "kategori"),
		IncludeInactive: BodyBool(c, "includeInactive"),
		Limit:           int(BodyFloat(c, "limit")),
		Offset:          int(BodyFloat(c, "offset")),
	}
}

func handleGetMembers(c *fiber.Ctx, svc *service.MemberService) error {
	f := parseMemberFilter(c)
	groupID := BodyString(c, "group_id")
	if groupID == "" {
		groupID = f.GroupLabel
	}
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	items, err := svc.GetMembers(c.Context(), f, groupID)
	if err != nil {
		return Fail(c, "Gagal ambil members: "+err.Error())
	}
	return Ok(c, items)
}

func handleGetMembersPaged(c *fiber.Ctx, svc *service.MemberService) error {
	f := parseMemberFilter(c)
	groupID := BodyString(c, "group_id")
	if groupID == "" {
		groupID = f.GroupLabel
	}
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	items, total, err := svc.GetMembersPaged(c.Context(), f, groupID)
	if err != nil {
		return Fail(c, "Gagal ambil members: "+err.Error())
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 30
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	return Ok(c, fiber.Map{
		"items":    items,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": offset+len(items) < total,
	})
}

func handleGetPNKBMembers(c *fiber.Ctx, svc *service.MemberService) error {
	f := parseMemberFilter(c)
	f.Kategori = "PRA_NIKAH"
	groupID := BodyString(c, "group_id")
	items, err := svc.GetMembers(c.Context(), f, groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetPNKBMembersPaged(c *fiber.Ctx, svc *service.MemberService) error {
	f := parseMemberFilter(c)
	f.Kategori = "PRA_NIKAH"
	groupID := BodyString(c, "group_id")
	items, total, err := svc.GetMembersPaged(c.Context(), f, groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 30
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	return Ok(c, fiber.Map{
		"items":    items,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": offset+len(items) < total,
	})
}

func handleGetAttendanceMembers(c *fiber.Ctx, svc *service.MemberService) error {
	f := parseMemberFilter(c)
	groupID := BodyString(c, "group_id")
	items, err := svc.GetAttendanceMembers(c.Context(), f, groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetMemberDetail(c *fiber.Ctx, svc *service.MemberService) error {
	id := BodyString(c, "member_id")
	if id == "" {
		return Fail(c, "member_id wajib diisi")
	}
	dto, err := svc.GetMemberDetail(c.Context(), id)
	if err != nil {
		return Fail(c, "Jamaah tidak ditemukan")
	}
	return Ok(c, dto)
}
