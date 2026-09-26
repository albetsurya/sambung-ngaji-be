package api

import (
	"context"
	"crypto/subtle"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func HandleCronReminder(c *fiber.Ctx) error {
	secret := os.Getenv("CRON_SECRET")
	if secret == "" {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "CRON_SECRET belum di-set",
		})
	}

	if subtle.ConstantTimeCompare([]byte(c.Get("X-Cron-Secret")), []byte(secret)) != 1 {
		log.Warn().Str("ip", c.IP()).Msg("cron request unauthorized")
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "unauthorized",
		})
	}

	if cronServices != nil && cronServices.Settings != nil {
		_ = cronServices.Settings.UpdateSettings(c.Context(), "last_cron_hit",
			time.Now().Format("2006-01-02T15:04:05.000Z07:00"))
	}

	go func() {
		ctx := context.Background()
		if err := cronServices.Reminder.RunOnce(ctx); err != nil {
			log.Error().Err(err).Msg("cron reminder error")
		}
	}()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "cron triggered",
	})
}

var cronServices *Services

func SetCronServices(s *Services) {
	cronServices = s
}
