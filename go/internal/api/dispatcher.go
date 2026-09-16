package api

import (
	"sort"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/auth"
)

// RegisteredActions — daftar action yang sudah diport ke Go.
// Dipakai untuk mengembalikan error yang jelas kalau action belum ada.
var RegisteredActions = map[string]bool{
	"login":           true,
	"logout":          true,
	"validateSession": true,
}

func RegisterAPI(app *fiber.App, svc *auth.Service) {
	app.Post("/api", BodyParserMiddleware(), AuthMiddleware(svc), func(c *fiber.Ctx) error {
		action, _ := BodyOf(c)["action"].(string)
		if action == "" {
			return Fail(c, "Parameter action wajib diisi")
		}

		switch action {
		case "login":
			return handleLogin(c, svc)
		case "logout":
			return handleLogout(c, svc)
		case "validateSession":
			return handleValidateSession(c, svc)
		}

		return Fail(c, "Action belum diimplementasi di Go: "+action)
	})
}

func ListRegisteredActions() []string {
	out := make([]string, 0, len(RegisteredActions))
	for k := range RegisteredActions {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
