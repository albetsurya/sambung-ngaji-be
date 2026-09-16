package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/model"
)

const (
	LocalsBody   = "body"
	LocalsUser   = "user"
	LocalsClaims = "claims"
)

// PublicActions — action yang tidak butuh token.
var PublicActions = map[string]bool{
	"login":                     true,
	"submitPublicRegistration":  true,
	"checkUsernameAvailability": true,
	"health":                    true,
}

// BodyParserMiddleware baca raw body JSON sekali, simpan di Locals.
// Handler berikutnya akses via BodyOf(c).
func BodyParserMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Body()
		if len(raw) == 0 {
			return c.Next()
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return c.Next()
		}
		c.Locals(LocalsBody, m)
		return c.Next()
	}
}

func BodyOf(c *fiber.Ctx) map[string]interface{} {
	if v := c.Locals(LocalsBody); v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return map[string]interface{}{}
}

func BodyString(c *fiber.Ctx, key string) string {
	m := BodyOf(c)
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// AuthMiddleware memverifikasi JWT dan set user di Locals.
func AuthMiddleware(svc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		action, _ := BodyOf(c)["action"].(string)
		if action == "" || PublicActions[action] {
			return c.Next()
		}

		token, _ := BodyOf(c)["token"].(string)
		if token == "" {
			return Fail(c, "Unauthorized: token tidak ada")
		}

		u, claims, err := svc.ValidateSession(c.Context(), token)
		if err != nil {
			return Fail(c, "Unauthorized: sesi tidak valid atau kadaluarsa")
		}

		c.Locals(LocalsUser, u)
		c.Locals(LocalsClaims, claims)
		return c.Next()
	}
}

func UserOf(c *fiber.Ctx) *model.User {
	if v := c.Locals(LocalsUser); v != nil {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}

func ClaimsOf(c *fiber.Ctx) *model.SessionClaims {
	if v := c.Locals(LocalsClaims); v != nil {
		if cl, ok := v.(*model.SessionClaims); ok {
			return cl
		}
	}
	return nil
}
