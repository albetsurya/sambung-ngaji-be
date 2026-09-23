package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

// handleSaveFridaySchedule: admin/tim absensi simpan jadwal petugas jumat (upsert by tanggal).
func handleSaveFridaySchedule(c *fiber.Ctx, svc *service.FridayService) error {
	in := service.SaveFridayInput{
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

// handleGetFridaySchedules: list jadwal petugas jumat dalam rentang tanggal.
func handleGetFridaySchedules(c *fiber.Ctx, svc *service.FridayService) error {
	res, err := svc.List(c.Context(), BodyString(c, "from"), BodyString(c, "to"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

// handleDeleteFridaySchedule: hapus jadwal jumat berdasarkan tanggal.
func handleDeleteFridaySchedule(c *fiber.Ctx, svc *service.FridayService) error {
	if err := svc.Delete(c.Context(), BodyString(c, "tanggal")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"tanggal": BodyString(c, "tanggal"), "deleted": true})
}
