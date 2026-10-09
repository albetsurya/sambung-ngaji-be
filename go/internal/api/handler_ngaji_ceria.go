package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func RegisterNgajiCeriaAPI(app *fiber.App, services *Services, authMiddleware fiber.Handler) {
	ngaji := app.Group("/api/v1/ngaji-ceria", authMiddleware)

	ngaji.Get("/progress", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil {
			return Fail(c, "Sesi tidak valid")
		}
		progress, err := services.NgajiCeria.GetProgress(c.Context(), u.UserID)
		if err != nil {
			return Fail(c, "Gagal mengambil progress Ngaji Ceria")
		}
		return Ok(c, progress)
	})

	ngaji.Post("/progress/events", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil {
			return Fail(c, "Sesi tidak valid")
		}
		var input service.NgajiProgressEvent
		if err := c.BodyParser(&input); err != nil {
			return Fail(c, "Payload progress tidak valid")
		}
		progress, err := services.NgajiCeria.AwardProgress(c.Context(), u.UserID, input)
		if err != nil {
			return Fail(c, err.Error())
		}
		return Ok(c, progress)
	})

	ngaji.Post("/missions/:missionId/claim", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil {
			return Fail(c, "Sesi tidak valid")
		}
		reward, err := services.NgajiCeria.ClaimMission(c.Context(), u.UserID, c.Params("missionId"))
		if err != nil {
			return Fail(c, err.Error())
		}
		return Ok(c, reward)
	})

	ngaji.Get("/leaderboard", func(c *fiber.Ctx) error {
		gameID := c.Query("gameId")
		period := c.Query("period", "weekly")
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		entries, err := services.NgajiCeria.GetLeaderboard(c.Context(), gameID, period, limit)
		if err != nil {
			return Fail(c, err.Error())
		}
		return Ok(c, fiber.Map{"gameId": gameID, "period": period, "entries": entries})
	})
	ngaji.Post("/leaderboard", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil {
			return Fail(c, "Sesi tidak valid")
		}
		var input service.NgajiProgressEvent
		if err := c.BodyParser(&input); err != nil {
			return Fail(c, "Skor tidak valid")
		}
		progress, err := services.NgajiCeria.AwardProgress(c.Context(), u.UserID, input)
		if err != nil {
			return Fail(c, err.Error())
		}
		return Ok(c, fiber.Map{"progress": progress})
	})
}
