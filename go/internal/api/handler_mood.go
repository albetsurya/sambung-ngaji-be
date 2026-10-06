package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleSaveMood(c *fiber.Ctx, svc *service.MoodService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	body := BodyOf(c)
	moodKey, _ := body["mood_key"].(string)
	date, _ := body["date"].(string)

	res, err := svc.Save(c.Context(), service.SaveMoodInput{
		MemberID: memberID,
		MoodKey:  moodKey,
		Date:     date,
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetMyMoods(c *fiber.Ctx, svc *service.MoodService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	res, err := svc.GetByMember(c.Context(), memberID, int(BodyFloat(c, "limit")))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetMemberMoods(c *fiber.Ctx, svc *service.MoodService) error {
	memberID := BodyString(c, "member_id")
	if memberID == "" {
		return Fail(c, "member_id wajib diisi")
	}
	res, err := svc.GetByMember(c.Context(), memberID, int(BodyFloat(c, "limit")))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
