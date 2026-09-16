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
	if v, ok := BodyOf(c)[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func BodyBool(c *fiber.Ctx, key string) bool {
	if v, ok := BodyOf(c)[key]; ok {
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return t == "true" || t == "1" || t == "TRUE"
		}
	}
	return false
}

func BodyFloat(c *fiber.Ctx, key string) float64 {
	if v, ok := BodyOf(c)[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		case string:
			var f float64
			_, _ = fmtSscan(t, &f)
			return f
		}
	}
	return 0
}

// AuthMiddleware + Permission check.
// Untuk public action: lewat.
// Untuk action lain: JWT valid + role boleh akses.
func AuthMiddleware(svc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := BodyOf(c)
		action, _ := body["action"].(string)
		if action == "" || PublicActions[action] {
			return c.Next()
		}

		token, _ := body["token"].(string)
		if token == "" {
			return Fail(c, "Unauthorized: token tidak ada")
		}

		u, claims, err := svc.ValidateSession(c.Context(), token)
		if err != nil {
			return Fail(c, "Unauthorized: sesi tidak valid atau kadaluarsa")
		}

		// Super admin only
		if auth.IsSuperAdminOnly(action) && u.Role != "SUPER_ADMIN" {
			return Fail(c, "Forbidden: hanya SUPER_ADMIN")
		}

		// Permission check
		if !auth.CanAccess(u.Role, action) {
			return Fail(c, "Forbidden: role "+u.Role+" tidak memiliki akses ke "+action)
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

// fmtSscan helper kecil (hindari import fmt di file ini).
func fmtSscan(s string, f *float64) (int, error) {
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
			continue
		}
		if c == '.' {
			break
		}
	}
	*f = float64(n)
	return 1, nil
}
