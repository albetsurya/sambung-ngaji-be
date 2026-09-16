package api

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/service"
)

var RegisteredActions = map[string]bool{
	"login":           true,
	"logout":          true,
	"validateSession": true,

	// Members
	"getMembers":           true,
	"getMembersPaged":      true,
	"getPNKBMembers":       true,
	"getPNKBMembersPaged":  true,
	"getAttendanceMembers": true,
	"getMemberDetail":      true,

	// Groups
	"getGroups": true,
}

type Services struct {
	Auth   *auth.Service
	Member *service.MemberService
	Group  *service.GroupService
}

func NewServices(pool *pgxpool.Pool, authSvc *auth.Service) *Services {
	return &Services{
		Auth:   authSvc,
		Member: service.NewMemberService(repository.NewMemberRepo(pool)),
		Group:  service.NewGroupService(repository.NewGroupRepo(pool)),
	}
}

func RegisterAPI(app *fiber.App, svc *Services) {
	app.Post("/api", BodyParserMiddleware(), AuthMiddleware(svc.Auth), func(c *fiber.Ctx) error {
		action, _ := BodyOf(c)["action"].(string)
		if action == "" {
			return Fail(c, "Parameter action wajib diisi")
		}

		switch action {
		case "login":
			return handleLogin(c, svc.Auth)
		case "logout":
			return handleLogout(c, svc.Auth)
		case "validateSession":
			return handleValidateSession(c, svc.Auth)

		case "getMembers":
			return handleGetMembers(c, svc.Member)
		case "getMembersPaged":
			return handleGetMembersPaged(c, svc.Member)
		case "getPNKBMembers":
			return handleGetPNKBMembers(c, svc.Member)
		case "getPNKBMembersPaged":
			return handleGetPNKBMembersPaged(c, svc.Member)
		case "getAttendanceMembers":
			return handleGetAttendanceMembers(c, svc.Member)
		case "getMemberDetail":
			return handleGetMemberDetail(c, svc.Member)

		case "getGroups":
			return handleGetGroups(c, svc.Group)
		}

		if !RegisteredActions[action] {
			return Fail(c, "Action belum diimplementasi di Go: "+action)
		}
		return Fail(c, "Action tidak ditemukan: "+action)
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
