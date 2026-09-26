package api

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/service"
)

const sessionCookieName = "pengajian_token"

func isProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

func sessionCookieMaxAge() int {
	if h, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS")); err == nil && h > 0 {
		return h * 3600
	}
	return 12 * 3600
}

func setSessionCookie(c *fiber.Ctx, token string, maxAge int) {
	sameSite := "Lax"
	secure := false
	if isProduction() {
		sameSite = "None"
		secure = true
	}
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func clearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   isProduction(),
		SameSite: "Lax",
	})
}

func handleLogin(c *fiber.Ctx, svc *auth.Service, auditSvc *service.AuditService) error {
	username := BodyString(c, "username")
	password := BodyString(c, "password")

	token, u, err := svc.Login(c.Context(), username, password)
	if err != nil {
		return Fail(c, "Gagal login")
	}

	setSessionCookie(c, token, sessionCookieMaxAge())

	if auditSvc != nil {
		uid := u.UserID
		auditSvc.Log(c.Context(), uid, "login", "session", uid)
	}

	public := svc.ToPublic(c.Context(), u)
	return Ok(c, fiber.Map{
		"token": token,
		"user":  public,
	})
}

func handleLogout(c *fiber.Ctx, svc *auth.Service) error {
	claims := ClaimsOf(c)
	if claims != nil {
		_ = svc.Logout(c.Context(), claims.SessionID)
	}
	clearSessionCookie(c)
	return Ok(c, nil)
}

func handleValidateSession(c *fiber.Ctx, svc *auth.Service) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	public := svc.ToPublic(c.Context(), u)
	return Ok(c, fiber.Map{"user": public})
}
