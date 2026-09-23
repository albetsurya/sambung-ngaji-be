package api

import (
	"context"
	"crypto/subtle"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// HandleCronReminder — POST /cron/reminder
// Dipanggil cron-job.org setiap 1 jam untuk trigger reminder.
// Header X-Cron-Secret harus match dengan CRON_SECRET env.
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

	// Trigger async — return 200 cepat biar cron-job.org tidak timeout
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

// SetCronServices — dipanggil dari main.go untuk inject service
var cronServices *Services

func SetCronServices(s *Services) {
	cronServices = s
}
