package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleSaveFridaySchedule(c *fiber.Ctx, svc *service.FridayService) error {
	in := service.SaveFridayInput{
		GroupID:       BodyString(c, "group_id"),
		Tanggal:       BodyString(c, "tanggal"),
		KhatibImam:    BodyString(c, "khatib_imam"),
		Muadzin:       BodyString(c, "muadzin"),
		Penasihat:     BodyString(c, "penasihat"),
		PetugasParkir: BodyString(c, "petugas_parkir"),
		PenataSandal:  BodyString(c, "penata_sandal"),
		Catatan:       BodyString(c, "catatan"),
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
	if err := svc.Delete(c.Context(), BodyString(c, "group_id"), BodyString(c, "tanggal")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"tanggal": BodyString(c, "tanggal"), "deleted": true})
}

func handleGetFridayReminderStatus(c *fiber.Ctx, svc *service.FridayReminderService) error {
	res, err := svc.GetReminderStatus(c.Context())
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleMarkFridayReminderSent(c *fiber.Ctx, svc *service.FridayReminderService) error {
	if err := svc.MarkSent(c.Context(), BodyString(c, "tanggal")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"marked": true})
}
