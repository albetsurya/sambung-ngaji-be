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
		MemberID: BodyString(c, "member_id"),
		Date:     BodyString(c, "date"),
		Type:     BodyString(c, "type"),
		Status:   BodyString(c, "status"),
		Notes:    BodyString(c, "notes"),
		FollowUp: BodyString(c, "follow_up"),
		UserID:   userID,
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
	if v, ok := body["notes"].(string); ok {
		in.Notes = &v
	}
	if v, ok := body["follow_up"].(string); ok {
		in.FollowUp = &v
	}
	dto, err := svc.UpdateEntry(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}
