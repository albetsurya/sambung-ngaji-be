package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/service"
)

func handleLogin(c *fiber.Ctx, svc *auth.Service, auditSvc *service.AuditService) error {
	username := BodyString(c, "username")
	password := BodyString(c, "password")

	token, u, err := svc.Login(c.Context(), username, password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			return Fail(c, "Username atau password salah")
		case auth.ErrUserInactive:
			return Fail(c, "Akun tidak aktif")
		default:
			return Fail(c, "Gagal login: "+err.Error())
		}
	}

	// Audit login manual (public action, UserOf kosong di middleware)
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
