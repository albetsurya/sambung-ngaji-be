package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetAnnouncementTemplates(c *fiber.Ctx, svc *service.AnnouncementService) error {
	items, err := svc.GetTemplates(c.Context(), false)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetAllAnnouncementTemplates(c *fiber.Ctx, svc *service.AnnouncementService) error {
	includeInactive := BodyBool(c, "include_inactive")
	items, err := svc.GetTemplates(c.Context(), includeInactive)
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
		NamaTemplate: BodyString(c, "nama_template"),
		Kode:         BodyString(c, "kode"),
		IsiTemplate:  BodyString(c, "isi_template"),
		StatusAktif:  true,
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
	if v, ok := body["nama_template"].(string); ok {
		in.NamaTemplate = &v
	}
	if v, ok := body["kode"].(string); ok {
		in.Kode = &v
	}
	if v, ok := body["isi_template"].(string); ok {
		in.IsiTemplate = &v
	}
	if v, ok := body["status_aktif"].(bool); ok {
		in.StatusAktif = &v
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
		BodyString(c, "nama_template"),
		BodyString(c, "kode"),
		BodyString(c, "source_announcement_id"),
		BodyString(c, "isi_template"),
	)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGenerateAnnouncement(c *fiber.Ctx, svc *service.AnnouncementService) error {
	in := service.GenerateAnnouncementInput{
		TemplateID:    BodyString(c, "template_id"),
		GroupID:       BodyString(c, "group_id"),
		Tanggal:       BodyString(c, "tanggal"),
		Jam:           BodyString(c, "jam"),
		Acara:         BodyString(c, "acara"),
		Materi:        BodyString(c, "materi"),
		Catatan:       BodyString(c, "catatan"),
		Penandatangan: BodyString(c, "penandatangan"),
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
		Tanggal:    BodyString(c, "tanggal"),
		Jam:        BodyString(c, "jam"),
		Acara:      BodyString(c, "acara"),
		Materi:     BodyString(c, "materi"),
		Catatan:    BodyString(c, "catatan"),
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
	if v, ok := body["jam"].(string); ok {
		in.Jam = &v
	}
	if v, ok := body["acara"].(string); ok {
		in.Acara = &v
	}
	if v, ok := body["materi"].(string); ok {
		in.Materi = &v
	}
	if v, ok := body["catatan"].(string); ok {
		in.Catatan = &v
	}
	dto, err := svc.Update(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGetAnnouncements(c *fiber.Ctx, svc *service.AnnouncementService) error {
	items, err := svc.GetAnnouncements(c.Context(), BodyString(c, "group_id"), BodyString(c, "status"))
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
		TemplateID:    BodyString(c, "template_id"),
		GroupID:       BodyString(c, "group_id"),
		WeekStart:     BodyString(c, "week_start"),
		Jam:           BodyString(c, "jam"),
		Acara:         BodyString(c, "acara"),
		Materi:        BodyString(c, "materi"),
		Catatan:       BodyString(c, "catatan"),
		Penandatangan: BodyString(c, "penandatangan"),
	}
	res, err := svc.GenerateWeekly(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
