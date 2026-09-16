package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

type PhotoHandler struct {
	storage  *service.StorageService
	members  *service.MemberService
	profiles *service.ProfileService
}

func NewPhotoHandler(
	storage *service.StorageService,
	members *service.MemberService,
	profiles *service.ProfileService,
) *PhotoHandler {
	return &PhotoHandler{storage: storage, members: members, profiles: profiles}
}

func (h *PhotoHandler) Upload(c *fiber.Ctx) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}

	body := BodyOf(c)

	// Tentukan member_id: MEMBER pakai member_id sendiri, admin kirim eksplisit
	var memberID string
	if u.Role == "MEMBER" {
		if u.MemberID == nil || *u.MemberID == "" {
			return Fail(c, "Akun Anda belum terhubung ke data jamaah")
		}
		memberID = *u.MemberID
	} else {
		memberID = BodyString(c, "member_id")
		if memberID == "" {
			return Fail(c, "member_id wajib diisi")
		}
	}

	base64Data := BodyString(c, "base64")
	mimeType := BodyString(c, "mime_type")
	if base64Data == "" || mimeType == "" {
		return Fail(c, "File foto wajib diisi")
	}

	// Hapus foto lama (kalau ada) sebelum upload baru — biar tidak numpuk
	if old, ok := body["_old_foto_url"].(string); ok && old != "" {
		_ = h.storage.DeletePhoto(c.Context(), old)
	}

	url, err := h.storage.UploadPhoto(c.Context(), memberID, base64Data, mimeType)
	if err != nil {
		return Fail(c, err.Error())
	}

	// Update foto_url di members
	if err := h.members.UpdateFotoURL(c.Context(), memberID, url); err != nil {
		return Fail(c, "Gagal simpan URL foto: "+err.Error())
	}

	return Ok(c, fiber.Map{"foto_url": url})
}

func (h *PhotoHandler) Delete(c *fiber.Ctx) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}

	var memberID string
	if u.Role == "MEMBER" {
		if u.MemberID == nil || *u.MemberID == "" {
			return Fail(c, "Akun Anda belum terhubung ke data jamaah")
		}
		memberID = *u.MemberID
	} else {
		memberID = BodyString(c, "member_id")
		if memberID == "" {
			return Fail(c, "member_id wajib diisi")
		}
	}

	// Ambil foto_url existing
	oldURL, err := h.members.GetFotoURL(c.Context(), memberID)
	if err != nil {
		return Fail(c, "Jamaah tidak ditemukan")
	}
	if oldURL == "" {
		return Fail(c, "Jamaah ini tidak memiliki foto")
	}

	// Hapus dari storage
	_ = h.storage.DeletePhoto(c.Context(), oldURL)

	// Kosongkan foto_url
	if err := h.members.UpdateFotoURL(c.Context(), memberID, ""); err != nil {
		return Fail(c, "Gagal hapus URL foto: "+err.Error())
	}

	return Ok(c, fiber.Map{"deleted": true})
}
