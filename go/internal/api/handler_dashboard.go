package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetDashboard(c *fiber.Ctx, svc *service.DashboardService) error {
	d, err := svc.GetGeneral(c.Context())
	if err != nil {
		return Fail(c, "Gagal ambil dashboard: "+err.Error())
	}
	return Ok(c, d)
}

func handleGetMyDashboard(c *fiber.Ctx, svc *service.DashboardService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	if u.MemberID == nil || *u.MemberID == "" {
		return Fail(c, "Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	d, err := svc.GetMyDashboard(c.Context(), *u.MemberID)
	if err != nil {
		return Fail(c, "Gagal ambil dashboard: "+err.Error())
	}
	return Ok(c, d)
}

func handleGetMonitoring(c *fiber.Ctx, svc *service.MonitoringService) error {
	memberID := BodyString(c, "member_id")
	if memberID == "" {
		return Fail(c, "member_id wajib diisi")
	}
	items, err := svc.FindByMember(c.Context(), memberID)
	if err != nil {
		return Fail(c, "Gagal ambil monitoring: "+err.Error())
	}
	return Ok(c, items)
}
