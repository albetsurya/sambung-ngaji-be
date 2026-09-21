package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterHealth memasang endpoint /health dan /ready.
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

	// Readiness probe — stricter check for k8s readiness
	app.Get("/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		// Check DB connectivity with a simple query
		var result int
		err := db.QueryRow(ctx, "SELECT 1").Scan(&result)
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error: " + err.Error()
		}

		// Check pool stats
		stats := db.Stat()
		poolStats := fiber.Map{
			"total_conns":    stats.TotalConns(),
			"idle_conns":     stats.IdleConns(),
			"acquired_conns": stats.AcquiredConns(),
			"max_conns":      stats.MaxConns(),
		}

		// Return 503 if DB not ready
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"data": fiber.Map{
					"status":     "not_ready",
					"database":   dbStatus,
					"pool_stats": poolStats,
					"time":       time.Now().Format(time.RFC3339),
				},
				"message": "database not ready",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":     "ready",
				"database":   dbStatus,
				"pool_stats": poolStats,
				"time":       time.Now().Format(time.RFC3339),
			},
			"message": "",
		})
	})
}
