package api

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// HandleWAWebhookVerify — GET /wa/webhook
// Dipakai Meta untuk verifikasi endpoint saat pertama kali didaftarkan.
func HandleWAWebhookVerify(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == os.Getenv("WA_VERIFY_TOKEN") {
		log.Info().Msg("WA webhook verified")
		return c.SendString(challenge)
	}
	log.Warn().Str("mode", mode).Msg("WA webhook verify failed")
	return c.SendStatus(403)
}

// HandleWAWebhookReceive — POST /wa/webhook
// Terima notifikasi dari Meta (pesan masuk, status delivery).
// Untuk MVP, cukup log & return 200. Meta retry kalau bukan 200.
func HandleWAWebhookReceive(c *fiber.Ctx) error {
	log.Info().Str("body", string(c.Body())).Msg("WA webhook received")
	return c.SendStatus(200)
}
