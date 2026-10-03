package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleSaveFridaySchedule(c *fiber.Ctx, svc *service.FridayService) error {
	in := service.SaveFridayInput{
		GroupID:       BodyString(c, "group_id"),
		Date:       BodyString(c, "date"),
		SermonLeader:    BodyString(c, "sermon_leader"),
		Muadzin:       BodyString(c, "muadzin"),
		Advisor:     BodyString(c, "advisor"),
		ParkingAttendant: BodyString(c, "parking_attendant"),
		FootwearAttendant:  BodyString(c, "footwear_attendant"),
		Notes:       BodyString(c, "notes"),
	}
	if u := UserOf(c); u != nil {
		in.CreatedBy = u.Username
	}
	res, err := svc.Save(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetFridaySchedules(c *fiber.Ctx, svc *service.FridayService) error {
	groupID := BodyString(c, "group_id")
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	res, err := svc.List(c.Context(), groupID, BodyString(c, "from"), BodyString(c, "to"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDeleteFridaySchedule(c *fiber.Ctx, svc *service.FridayService) error {
	if err := svc.Delete(c.Context(), BodyString(c, "group_id"), BodyString(c, "date")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"date": BodyString(c, "date"), "deleted": true})
}

func handleGetFridayReminderStatus(c *fiber.Ctx, svc *service.FridayReminderService) error {
	res, err := svc.GetReminderStatus(c.Context())
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleMarkFridayReminderSent(c *fiber.Ctx, svc *service.FridayReminderService) error {
	if err := svc.MarkSent(c.Context(), BodyString(c, "date")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"marked": true})
}
