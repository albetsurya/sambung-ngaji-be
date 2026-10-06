package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/service"
)

func handleGetMeetings(c *fiber.Ctx, svc *service.MeetingService) error {
	groupID := BodyString(c, "group_id")
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	f := model.MeetingListFilter{
		From:    BodyString(c, "from"),
		To:      BodyString(c, "to"),
		GroupID: groupID,
	}
	items, err := svc.GetMeetings(c.Context(), f)
	if err != nil {
		return Fail(c, "Gagal ambil meetings: "+err.Error())
	}
	return Ok(c, items)
}

func parseKategoriTarget(c *fiber.Ctx) []string {
	body := BodyOf(c)
	v, ok := body["target_categories"]
	if !ok {
		return nil
	}
	raw, ok := v.([]interface{})
	if !ok {
		if arr, ok := v.([]string); ok {
			return arr
		}
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, x := range raw {
		if s, ok := x.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func handleCreateMeeting(c *fiber.Ctx, svc *service.MeetingService) error {
	u := UserOf(c)
	createdBy := ""
	if u != nil {
		createdBy = u.UserID
	}
	groupID, err := RequireGroupCreate(c)
	if err != nil {
		return err
	}
	in := service.CreateMeetingInput{
		Date:             BodyString(c, "date"),
		Time:             BodyString(c, "time"),
		StartTime:        BodyString(c, "start_time"),
		GroupID:          groupID,
		Event:            BodyString(c, "event"),
		Topic:            BodyString(c, "topic"),
		Status:           BodyString(c, "status"),
		Notes:            BodyString(c, "notes"),
		TargetCategories: parseKategoriTarget(c),
		GenderTarget:     BodyString(c, "gender_target"),
		CreatedBy:        createdBy,
	}
	dto, err := svc.CreateMeeting(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateMeeting(c *fiber.Ctx, svc *service.MeetingService) error {
	if err := RequireMeetingAccess(c, svc, BodyString(c, "meeting_id")); err != nil {
		return err
	}
	in := service.UpdateMeetingInput{
		MeetingID: BodyString(c, "meeting_id"),
		Date:      BodyString(c, "date"),
		Time:      BodyString(c, "time"),
		StartTime: BodyString(c, "start_time"),
		GroupID:   BodyString(c, "group_id"),
		Event:     BodyString(c, "event"),
		Topic:     BodyString(c, "topic"),
		Status:    BodyString(c, "status"),
		Notes:     BodyString(c, "notes"),
	}
	if _, ok := BodyOf(c)["target_categories"]; ok {
		kat := parseKategoriTarget(c)
		in.TargetCategories = &kat
	}
	if _, ok := BodyOf(c)["gender_target"]; ok {
		gt := BodyString(c, "gender_target")
		in.GenderTarget = &gt
	}
	if _, ok := BodyOf(c)["target_categories"]; ok {
		kat := parseKategoriTarget(c)
		in.TargetCategories = &kat
	}
	dto, err := svc.UpdateMeeting(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleDeleteMeeting(c *fiber.Ctx, svc *service.MeetingService) error {
	id := BodyString(c, "meeting_id")
	if err := RequireMeetingAccess(c, svc, id); err != nil {
		return err
	}
	res, err := svc.DeleteMeeting(c.Context(), id)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDeleteMeetingsBulk(c *fiber.Ctx, svc *service.MeetingService) error {
	body := BodyOf(c)
	raw, ok := body["meeting_ids"].([]interface{})
	if !ok {
		return Fail(c, "meeting_ids wajib diisi (array)")
	}
	ids := make([]string, 0, len(raw))
	for _, x := range raw {
		if s, ok := x.(string); ok && s != "" {
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		return Fail(c, "meeting_ids tidak boleh kosong")
	}

	if err := RequireMeetingsAccess(c, svc, ids); err != nil {
		return err
	}
	in := service.DeleteMeetingsBulkInput{MeetingIDs: ids}
	res, err := svc.DeleteMeetingBulk(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
