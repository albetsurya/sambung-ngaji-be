package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetAnnouncementTemplates(c *fiber.Ctx, svc *service.AnnouncementService) error {
	groupID := BodyString(c, "group_id")
	items, err := svc.GetTemplates(c.Context(), groupID, false)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetAllAnnouncementTemplates(c *fiber.Ctx, svc *service.AnnouncementService) error {
	groupID := BodyString(c, "group_id")
	includeInactive := BodyBool(c, "include_inactive")
	items, err := svc.GetTemplates(c.Context(), groupID, includeInactive)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetAnnouncementTemplateDetail(c *fiber.Ctx, svc *service.AnnouncementService) error {
	id := BodyString(c, "template_id")
	dto, err := svc.GetTemplateDetail(c.Context(), id)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleCreateAnnouncementTemplate(c *fiber.Ctx, svc *service.AnnouncementService) error {
	in := service.CreateTemplateInput{
		GroupID:      BodyString(c, "group_id"),
		TemplateName: BodyString(c, "template_name"),
		Kode:         BodyString(c, "kode"),
		TemplateBody: BodyString(c, "template_body"),
		IsActive:     true,
	}
	dto, err := svc.CreateTemplate(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateAnnouncementTemplate(c *fiber.Ctx, svc *service.AnnouncementService) error {
	body := BodyOf(c)
	in := service.UpdateTemplateInput{
		TemplateID: BodyString(c, "template_id"),
	}
	if v, ok := body["template_name"].(string); ok {
		in.TemplateName = &v
	}
	if v, ok := body["kode"].(string); ok {
		in.Kode = &v
	}
	if v, ok := body["template_body"].(string); ok {
		in.TemplateBody = &v
	}
	if v, ok := body["is_active"].(bool); ok {
		in.IsActive = &v
	}
	dto, err := svc.UpdateTemplate(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleDeleteAnnouncementTemplate(c *fiber.Ctx, svc *service.AnnouncementService) error {
	id := BodyString(c, "template_id")
	if err := svc.DeleteTemplate(c.Context(), id); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true, "template_id": id})
}

func handleCreateTemplateFromAnnouncement(c *fiber.Ctx, svc *service.AnnouncementService) error {
	dto, err := svc.CreateTemplateFromAnnouncement(
		c.Context(),
		BodyString(c, "template_name"),
		BodyString(c, "kode"),
		BodyString(c, "source_announcement_id"),
		BodyString(c, "template_body"),
	)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGenerateAnnouncement(c *fiber.Ctx, svc *service.AnnouncementService) error {
	in := service.GenerateAnnouncementInput{
		TemplateID: BodyString(c, "template_id"),
		GroupID:    BodyString(c, "group_id"),
		Date:       BodyString(c, "date"),
		Time:       BodyString(c, "time"),
		Event:      BodyString(c, "event"),
		Topic:      BodyString(c, "topic"),
		Notes:      BodyString(c, "notes"),
		Signatory:  BodyString(c, "signatory"),
	}
	res, err := svc.Generate(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleCreateAnnouncement(c *fiber.Ctx, svc *service.AnnouncementService) error {
	u := UserOf(c)
	userID := ""
	if u != nil {
		userID = u.UserID
	}
	in := service.CreateAnnouncementInput{
		TemplateID: BodyString(c, "template_id"),
		MeetingID:  BodyString(c, "meeting_id"),
		GroupID:    BodyString(c, "group_id"),
		Date:       BodyString(c, "date"),
		Time:       BodyString(c, "time"),
		Event:      BodyString(c, "event"),
		Topic:      BodyString(c, "topic"),
		Notes:      BodyString(c, "notes"),
		UserID:     userID,
	}
	dto, err := svc.Create(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateAnnouncement(c *fiber.Ctx, svc *service.AnnouncementService) error {
	body := BodyOf(c)
	in := service.UpdateAnnouncementInput{
		AnnouncementID: BodyString(c, "announcement_id"),
	}
	if v, ok := body["generated_text"].(string); ok {
		in.GeneratedText = &v
	}
	if v, ok := body["status"].(string); ok {
		in.Status = &v
	}
	if v, ok := body["time"].(string); ok {
		in.Time = &v
	}
	if v, ok := body["event"].(string); ok {
		in.Event = &v
	}
	if v, ok := body["topic"].(string); ok {
		in.Topic = &v
	}
	if v, ok := body["notes"].(string); ok {
		in.Notes = &v
	}
	dto, err := svc.Update(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGetAnnouncements(c *fiber.Ctx, svc *service.AnnouncementService) error {
	groupID := BodyString(c, "group_id")
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	items, err := svc.GetAnnouncements(c.Context(), groupID, BodyString(c, "status"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetAnnouncementRecipientSummary(c *fiber.Ctx, svc *service.AnnouncementService) error {
	res, err := svc.GetRecipientSummary(c.Context(), BodyString(c, "group_id"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGenerateWeeklyAnnouncements(c *fiber.Ctx, svc *service.AnnouncementService) error {
	in := service.WeeklyGenerateInput{
		TemplateID: BodyString(c, "template_id"),
		GroupID:    BodyString(c, "group_id"),
		WeekStart:  BodyString(c, "week_start"),
		Time:       BodyString(c, "time"),
		Event:      BodyString(c, "event"),
		Topic:      BodyString(c, "topic"),
		Notes:      BodyString(c, "notes"),
		Signatory:  BodyString(c, "signatory"),
	}
	res, err := svc.GenerateWeekly(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
