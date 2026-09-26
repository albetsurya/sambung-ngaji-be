package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHealth(app *fiber.App, db *pgxpool.Pool) {
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error"
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

	app.Get("/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		var result int
		err := db.QueryRow(ctx, "SELECT 1").Scan(&result)
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error"
		}

		_ = db.Stat().TotalConns()

		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"data": fiber.Map{
					"status":   "not_ready",
					"database": dbStatus,
					"time":     time.Now().Format(time.RFC3339),
				},
				"message": "database not ready",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":   "ready",
				"database": dbStatus,
				"time":     time.Now().Format(time.RFC3339),
			},
			"message": "",
		})
	})
}
