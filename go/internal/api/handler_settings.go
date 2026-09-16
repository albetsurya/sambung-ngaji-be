package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetSettings(c *fiber.Ctx, svc *service.SettingsService) error {
	settings, err := svc.GetSettings(c.Context())
	if err != nil {
		return Fail(c, "Gagal ambil settings: "+err.Error())
	}
	return Ok(c, settings)
}

func handleUpdateSettings(c *fiber.Ctx, svc *service.SettingsService) error {
	key := BodyString(c, "key")
	if key == "" {
		return Fail(c, "key wajib diisi")
	}
	body := BodyOf(c)
	value, ok := body["value"]
	if !ok {
		return Fail(c, "value wajib diisi")
	}
	if err := svc.UpdateSettings(c.Context(), key, value); err != nil {
		return Fail(c, "Gagal simpan settings: "+err.Error())
	}
	return Ok(c, fiber.Map{"key": key, "value": value})
}
