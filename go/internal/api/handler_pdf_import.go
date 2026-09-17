package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleParsePdfMeeting(c *fiber.Ctx, svc *service.PDFImportService) error {
	text := BodyString(c, "text")
	if text == "" {
		return Fail(c, "text PDF wajib diisi")
	}

	result, err := svc.ParsePDF(c.Context(), text)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, result)
}
