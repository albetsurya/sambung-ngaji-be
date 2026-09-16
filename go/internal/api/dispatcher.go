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

	"getMembers":           true,
	"getMembersPaged":      true,
	"getPNKBMembers":       true,
	"getPNKBMembersPaged":  true,
	"getAttendanceMembers": true,
	"getMemberDetail":      true,

	"getGroups": true,

	"getMeetings":   true,
	"createMeeting": true,
	"updateMeeting": true,
	"deleteMeeting": true,

	// Attendance
	"getAttendance":             true,
	"getAttendancePage":         true,
	"saveAttendance":            true,
	"bulkSaveAttendance":        true,
	"deleteAttendance":          true,
	"deleteAttendanceByMeeting": true,
	"deleteAttendanceByMember":  true,
}

type Services struct {
	Auth       *auth.Service
	Member     *service.MemberService
	Group      *service.GroupService
	Meeting    *service.MeetingService
	Attendance *service.AttendanceService
	Monitoring *service.MonitoringService
	Dashboard  *service.DashboardService
	Settings   *service.SettingsService
}

func NewServices(pool *pgxpool.Pool, authSvc *auth.Service) *Services {
	return &Services{
		Auth:    authSvc,
		Member:  service.NewMemberService(repository.NewMemberRepo(pool)),
		Group:   service.NewGroupService(repository.NewGroupRepo(pool)),
		Meeting: service.NewMeetingService(repository.NewMeetingRepo(pool)),
		Attendance: service.NewAttendanceService(
			repository.NewAttendanceRepo(pool),
			repository.NewMeetingRepo(pool),
			service.NewMemberService(repository.NewMemberRepo(pool)),
		),
		Monitoring: service.NewMonitoringService(repository.NewMonitoringRepo(pool)),
		Dashboard: service.NewDashboardService(
			repository.NewMemberRepo(pool),
			repository.NewMeetingRepo(pool),
			repository.NewAttendanceRepo(pool),
			repository.NewMonitoringRepo(pool),
		),
		Settings: service.NewSettingsService(repository.NewSettingsRepo(pool)),
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

		case "getMeetings":
			return handleGetMeetings(c, svc.Meeting)
		case "createMeeting":
			return handleCreateMeeting(c, svc.Meeting)
		case "updateMeeting":
			return handleUpdateMeeting(c, svc.Meeting)
		case "deleteMeeting":
			return handleDeleteMeeting(c, svc.Meeting)

		case "getAttendance":
			return handleGetAttendance(c, svc.Attendance)
		case "getAttendancePage":
			return handleGetAttendancePage(c, svc.Attendance)
		case "saveAttendance":
			return handleSaveAttendance(c, svc.Attendance)
		case "bulkSaveAttendance":
			return handleBulkSaveAttendance(c, svc.Attendance)
		case "deleteAttendance":
			return handleDeleteAttendance(c, svc.Attendance)
		case "deleteAttendanceByMeeting":
			return handleDeleteAttendanceByMeeting(c, svc.Attendance)
		case "deleteAttendanceByMember":
			return handleDeleteAttendanceByMember(c, svc.Attendance)

		case "getDashboard":
			return handleGetDashboard(c, svc.Dashboard)
		case "getMyDashboard":
			return handleGetMyDashboard(c, svc.Dashboard)
		case "getMonitoring":
			return handleGetMonitoring(c, svc.Monitoring)

		case "getSettings":
			return handleGetSettings(c, svc.Settings)
		case "updateSettings":
			return handleUpdateSettings(c, svc.Settings)
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
