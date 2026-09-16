package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterHealth memasang endpoint /health.
// Response mengikuti kontrak existing: { success, data, message }.
func RegisterHealth(app *fiber.App, db *pgxpool.Pool) {
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error: " + err.Error()
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":   "ok",
				"database": dbStatus,
				"time":     time.Now().Format(time.RFC3339),
			},
			"message": "",
		})
	})
}
