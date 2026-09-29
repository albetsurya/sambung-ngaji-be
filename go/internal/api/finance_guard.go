package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/auth"
)

func FinanceAuthMiddleware(svc *auth.Service, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			token = c.Cookies(sessionCookieName)
		}
		if token == "" {
			return Fail(c, "Unauthorized: token tidak ada")
		}
		token = strings.TrimPrefix(token, "Bearer ")

		u, claims, err := svc.ValidateSession(c.Context(), token)
		if err != nil {
			return Fail(c, "Unauthorized: sesi tidak valid atau kadaluarsa")
		}
		if !auth.CanAccess(u.Role, action) {
			return Fail(c, "Forbidden: role "+u.Role+" tidak memiliki akses keuangan")
		}

		c.Locals(LocalsUser, u)
		c.Locals(LocalsClaims, claims)

		if u.Role != "SUPER_ADMIN" {
			body := BodyOf(c)
			if u.GroupID != nil && *u.GroupID != "" {
				body["group_id"] = *u.GroupID
			} else {
				body["group_id"] = UnassignedGroup
			}
		}

		return c.Next()
	}
}
