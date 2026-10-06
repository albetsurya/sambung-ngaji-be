package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/service"
)

func handleGetAttendance(c *fiber.Ctx, svc *service.AttendanceService) error {
	meetingID := BodyString(c, "meeting_id")
	memberID := BodyString(c, "member_id")
	items, err := svc.GetAttendance(c.Context(), meetingID, memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetAttendancePage(c *fiber.Ctx, svc *service.AttendanceService) error {
	meetingID := BodyString(c, "meeting_id")
	page, err := svc.GetAttendancePage(c.Context(), meetingID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, page)
}

func handleSaveAttendance(c *fiber.Ctx, svc *service.AttendanceService) error {
	u := UserOf(c)
	userID := ""
	if u != nil {
		userID = u.UserID
	}
	in := service.SaveAttendanceInput{
		MeetingID: BodyString(c, "meeting_id"),
		MemberID:  BodyString(c, "member_id"),
		Status:    BodyString(c, "status"),
		Notes:     BodyString(c, "notes"),
		UserID:    userID,
	}
	dto, err := svc.SaveAttendance(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleBulkSaveAttendance(c *fiber.Ctx, svc *service.AttendanceService) error {
	u := UserOf(c)
	userID := ""
	if u != nil {
		userID = u.UserID
	}

	body := BodyOf(c)
	meetingID, _ := body["meeting_id"].(string)

	rawItems, _ := body["items"].([]interface{})
	items := make([]repository.BulkItem, 0, len(rawItems))
	for _, raw := range rawItems {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		mid, _ := m["member_id"].(string)
		st, _ := m["status"].(string)
		cat, _ := m["notes"].(string)
		items = append(items, repository.BulkItem{
			MemberID: mid,
			Status:   st,
			Notes:    cat,
		})
	}

	in := service.BulkSaveInput{
		MeetingID: meetingID,
		Items:     items,
		UserID:    userID,
	}
	res, err := svc.BulkSave(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDeleteAttendance(c *fiber.Ctx, svc *service.AttendanceService) error {
	meetingID := BodyString(c, "meeting_id")
	memberID := BodyString(c, "member_id")
	n, err := svc.DeleteAttendance(c.Context(), meetingID, memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": n})
}

func handleDeleteAttendanceByMeeting(c *fiber.Ctx, svc *service.AttendanceService) error {
	meetingID := BodyString(c, "meeting_id")
	n, err := svc.DeleteByMeeting(c.Context(), meetingID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": n})
}

func handleDeleteAttendanceByMember(c *fiber.Ctx, svc *service.AttendanceService) error {
	memberID := BodyString(c, "member_id")
	n, err := svc.DeleteByMember(c.Context(), memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": n})
}
