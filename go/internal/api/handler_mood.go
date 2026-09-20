package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

// handleSaveMood: member mencatat mood hariannya (member_id dari JWT).
func handleSaveMood(c *fiber.Ctx, svc *service.MoodService) error {
	memberID, err := requireMemberID(c)
	if err != nil {
		return Fail(c, err.Error())
	}
	body := BodyOf(c)
	moodKey, _ := body["mood_key"].(string)
	tanggal, _ := body["tanggal"].(string)

	res, err := svc.Save(c.Context(), service.SaveMoodInput{
		MemberID: memberID,
		MoodKey:  moodKey,
		Tanggal:  tanggal,
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

// handleGetMyMoods: member melihat riwayat mood sendiri.
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

// handleGetMemberMoods: admin/pengawas melihat mood member tertentu.
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
