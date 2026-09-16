package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleCreateMonitoring(c *fiber.Ctx, svc *service.MonitoringService) error {
	u := UserOf(c)
	userID := ""
	if u != nil {
		userID = u.UserID
	}
	in := service.CreateMonitoringInput{
		MemberID:     BodyString(c, "member_id"),
		Tanggal:      BodyString(c, "tanggal"),
		Jenis:        BodyString(c, "jenis"),
		Status:       BodyString(c, "status"),
		Catatan:      BodyString(c, "catatan"),
		TindakLanjut: BodyString(c, "tindak_lanjut"),
		UserID:       userID,
	}
	dto, err := svc.Create(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateMonitoring(c *fiber.Ctx, svc *service.MonitoringService) error {
	body := BodyOf(c)
	in := service.UpdateMonitoringInput{
		MonitoringID: BodyString(c, "monitoring_id"),
	}
	if v, ok := body["catatan"].(string); ok {
		in.Catatan = &v
	}
	if v, ok := body["tindak_lanjut"].(string); ok {
		in.TindakLanjut = &v
	}
	dto, err := svc.UpdateEntry(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}
