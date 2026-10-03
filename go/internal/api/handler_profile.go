package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func requireMemberID(c *fiber.Ctx) (string, error) {
	u := UserOf(c)
	if u == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
	if u.MemberID == nil || *u.MemberID == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	return *u.MemberID, nil
}

func handleGetMyProfile(c *fiber.Ctx, svc *service.ProfileService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	res, err := svc.GetMyProfile(c.Context(), memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleUpdateMyProfile(c *fiber.Ctx, svc *service.ProfileService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	body := BodyOf(c)
	patch := map[string]interface{}{}
	for _, k := range []string{
		"nickname", "whatsapp_number", "home_address", "village", "region",
		"occupation", "hobby", "photo_url", "height", "weight",
		"education_level", "school", "major",
		"education_start_year", "education_end_year",
	} {
		if v, ok := body[k]; ok {
			if s, ok := v.(string); ok {
				patch[k] = s
			}
		}
	}
	for _, k := range []string{"is_employed", "is_married"} {
		if v, ok := body[k].(bool); ok {
			patch[k] = v
		}
	}
	res, err := svc.UpdateMyProfile(c.Context(), memberID, patch)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetMyAttendance(c *fiber.Ctx, svc *service.ProfileService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	res, err := svc.GetMyAttendance(c.Context(), memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetMyMonitoring(c *fiber.Ctx, svc *service.ProfileService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	res, err := svc.GetMyMonitoring(c.Context(), memberID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetUpcomingMeetings(c *fiber.Ctx, svc *service.ProfileService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	limit := int(BodyFloat(c, "limit"))
	res, err := svc.GetUpcomingMeetings(c.Context(), memberID, limit)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
