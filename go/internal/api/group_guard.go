package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

const UnassignedGroup = "__UNASSIGNED_GROUP__"

func ActorOf(c *fiber.Ctx) (groupID string, isSuper bool) {
	u := UserOf(c)
	if u == nil {
		return "", false
	}
	if u.Role == "SUPER_ADMIN" {
		return "", true
	}
	if u.GroupID != nil && *u.GroupID != "" {
		return *u.GroupID, false
	}
	return UnassignedGroup, false
}

func RequireGroupAccess(c *fiber.Ctx, recordGroupID *string) error {
	groupID, isSuper := ActorOf(c)
	if isSuper {
		return nil
	}
	if groupID == "" || groupID == UnassignedGroup {
		return Fail(c, "Akun Anda belum dipetakan ke kelompok")
	}
	rec := ""
	if recordGroupID != nil {
		rec = *recordGroupID
	}
	if rec == "" {
		return Fail(c, "Data ini milik global, hanya SUPER_ADMIN yang boleh mengubah")
	}
	if rec != groupID {
		return Fail(c, "Forbidden: data milik kelompok lain")
	}
	return nil
}

func RequireGroupCreate(c *fiber.Ctx) (string, error) {
	groupID, isSuper := ActorOf(c)
	if isSuper {
		return BodyString(c, "group_id"), nil
	}
	if groupID == "" || groupID == UnassignedGroup {
		return "", Fail(c, "Akun Anda belum dipetakan ke kelompok")
	}
	return groupID, nil
}

/* Guard tulis per-record meeting: pemilik group-nya (atau SUPER_ADMIN) saja. */
func RequireMeetingAccess(c *fiber.Ctx, svc *service.MeetingService, id string) error {
	m, err := svc.GetMeeting(c.Context(), id)
	if err != nil {
		return Fail(c, "Jadwal tidak ditemukan")
	}
	return RequireGroupAccess(c, m.GroupID)
}

func RequireMeetingsAccess(c *fiber.Ctx, svc *service.MeetingService, ids []string) error {
	for _, id := range ids {
		if err := RequireMeetingAccess(c, svc, id); err != nil {
			return err
		}
	}
	return nil
}
