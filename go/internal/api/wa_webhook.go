package api

import (
	"crypto/subtle"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// HandleWAWebhookVerify — GET /wa/webhook
// Dipakai Meta untuk verifikasi endpoint saat pertama kali didaftarkan.
// Fail-closed: tolak semua bila WA_VERIFY_TOKEN belum di-set.
func HandleWAWebhookVerify(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	expected := os.Getenv("WA_VERIFY_TOKEN")
	if expected == "" {
		log.Error().Msg("WA webhook verify ditolak: WA_VERIFY_TOKEN belum di-set")
		return c.SendStatus(403)
	}

	if mode == "subscribe" &&
		subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
		log.Info().Msg("WA webhook verified")
		return c.SendString(challenge)
	}
	log.Warn().Str("mode", mode).Msg("WA webhook verify failed")
	return c.SendStatus(403)
}

// HandleWAWebhookReceive — POST /wa/webhook
// Terima notifikasi dari Meta (pesan masuk, status delivery).
// Untuk MVP, cukup log & return 200. Meta retry kalau bukan 200.
// Body TIDAK di-log utuh agar tidak membocorkan data PII ke log.
func HandleWAWebhookReceive(c *fiber.Ctx) error {
	log.Info().Int("body_len", len(c.Body())).Msg("WA webhook received")
	return c.SendStatus(200)
}
