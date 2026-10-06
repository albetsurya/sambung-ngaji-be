package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleCheckUsernameAvailability(c *fiber.Ctx, svc *service.PendingService) error {
	username := BodyString(c, "username")
	res, err := svc.CheckUsername(c.Context(), username)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleSubmitPublicRegistration(c *fiber.Ctx, svc *service.PendingService) error {
	clientIP := c.IP()
	if v := c.Get("X-Forwarded-For"); v != "" {
		clientIP = v
	}
	in := service.SubmitRegistrationInput{
		GroupID:            BodyString(c, "group_id"),
		FullName:           BodyString(c, "full_name"),
		Nickname:           BodyString(c, "nickname"),
		Gender:             BodyString(c, "gender"),
		BirthPlace:         BodyString(c, "birth_place"),
		BirthDate:          BodyString(c, "birth_date"),
		WhatsappNumber:     BodyString(c, "whatsapp_number"),
		HomeAddress:        BodyString(c, "home_address"),
		Village:            BodyString(c, "village"),
		Region:             BodyString(c, "region"),
		Occupation:         BodyString(c, "occupation"),
		Hobby:              BodyString(c, "hobby"),
		IsMarried:          BodyBool(c, "is_married"),
		EducationLevel:     BodyString(c, "education_level"),
		School:             BodyString(c, "school"),
		Major:              BodyString(c, "major"),
		EducationStartYear: BodyString(c, "education_start_year"),
		EducationEndYear:   BodyString(c, "education_end_year"),
		PhotoURL:           BodyString(c, "photo_url"),
		Username:           BodyString(c, "username"),
		Password:           BodyString(c, "password"),
		ClientIP:           clientIP,
	}
	res, err := svc.SubmitRegistration(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetPendingMembers(c *fiber.Ctx, svc *service.PendingService) error {
	groupID := BodyString(c, "group_id")
	/* Akun ber-group_label dikunci ke kelompoknya (abaikan param klien). */
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	status := BodyString(c, "status")
	items, err := svc.GetPendingMembers(c.Context(), groupID, status)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetPendingMemberDetail(c *fiber.Ctx, svc *service.PendingService) error {
	id := BodyString(c, "submission_id")
	dto, err := svc.GetPendingMemberDetail(c.Context(), id)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleApprovePendingMember(c *fiber.Ctx, svc *service.PendingService) error {
	u := UserOf(c)
	reviewerID := ""
	if u != nil {
		reviewerID = u.UserID
	}
	res, err := svc.ApproveWithGroup(c.Context(),
		BodyString(c, "submission_id"),
		BodyString(c, "group_id"),
		BodyString(c, "group_label"),
		reviewerID,
	)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleRejectPendingMember(c *fiber.Ctx, svc *service.PendingService) error {
	u := UserOf(c)
	reviewerID := ""
	if u != nil {
		reviewerID = u.UserID
	}
	if err := svc.Reject(c.Context(),
		BodyString(c, "submission_id"),
		reviewerID,
		BodyString(c, "reason"),
	); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{
		"submission_id": BodyString(c, "submission_id"),
		"status":        "REJECTED",
	})
}
