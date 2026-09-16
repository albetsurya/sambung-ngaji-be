package api

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/ai"
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

	// Meetings
	"getMeetings":   true,
	"createMeeting": true,
	"updateMeeting": true,
	"deleteMeeting": true,

	// Bulk Meeting
	"previewBulkMeetings":     true,
	"bulkCreateMeetings":      true,
	"getBulkMeetingTemplates": true,

	// Attendance
	"getAttendance":             true,
	"getAttendancePage":         true,
	"saveAttendance":            true,
	"bulkSaveAttendance":        true,
	"deleteAttendance":          true,
	"deleteAttendanceByMeeting": true,
	"deleteAttendanceByMember":  true,

	// Dashboard & Monitoring
	"getDashboard":   true,
	"getMyDashboard": true,
	"getMonitoring":  true,

	// Settings
	"getSettings":    true,
	"updateSettings": true,

	// Announcements
	"getAnnouncementTemplates":        true,
	"getAllAnnouncementTemplates":     true,
	"getAnnouncementTemplateDetail":   true,
	"createAnnouncementTemplate":      true,
	"updateAnnouncementTemplate":      true,
	"deleteAnnouncementTemplate":      true,
	"createTemplateFromAnnouncement":  true,
	"generateAnnouncement":            true,
	"createAnnouncement":              true,
	"updateAnnouncement":              true,
	"getAnnouncements":                true,
	"getAnnouncementRecipientSummary": true,

	// Pending
	"getPendingMembers":         true,
	"getPendingMemberDetail":    true,
	"approvePendingMember":      true,
	"rejectPendingMember":       true,
	"submitPublicRegistration":  true,
	"checkUsernameAvailability": true,

	// Users
	"getUsers":            true,
	"getUserDetail":       true,
	"createUser":          true,
	"updateUser":          true,
	"updateUserRole":      true,
	"getMemberUserStatus": true,
	"changeMyPassword":    true,
	"resetUserPassword":   true,

	// Profile (member self-service)
	"getMyProfile":        true,
	"updateMyProfile":     true,
	"getMyAttendance":     true,
	"getMyMonitoring":     true,
	"getUpcomingMeetings": true,

	// Audit
	"getAuditLogs": true,

	// AI
	"aiChat":          true,
	"getAiUsageStats": true,
}

type Services struct {
	Auth         *auth.Service
	Member       *service.MemberService
	Group        *service.GroupService
	Meeting      *service.MeetingService
	BulkMeeting  *service.BulkMeetingService
	Attendance   *service.AttendanceService
	Monitoring   *service.MonitoringService
	Dashboard    *service.DashboardService
	Settings     *service.SettingsService
	Announcement *service.AnnouncementService
	Pending      *service.PendingService
	User         *service.UserService
	Profile      *service.ProfileService
	Audit        *service.AuditService
	AI           *service.AIService
}

func NewServices(pool *pgxpool.Pool, authSvc *auth.Service, omniRoute *ai.OmniRoute) *Services {
	return &Services{
		Auth:    authSvc,
		Member:  service.NewMemberService(repository.NewMemberRepo(pool)),
		Group:   service.NewGroupService(repository.NewGroupRepo(pool)),
		Meeting: service.NewMeetingService(repository.NewMeetingRepo(pool)),
		BulkMeeting: service.NewBulkMeetingService(
			repository.NewMeetingRepo(pool),
			repository.NewGroupRepo(pool),
			repository.NewAnnouncementRepo(pool),
		),

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
		Announcement: service.NewAnnouncementService(
			repository.NewAnnouncementRepo(pool),
			repository.NewGroupRepo(pool),
			repository.NewMemberRepo(pool),
		),
		Pending: service.NewPendingService(
			repository.NewPendingRepo(pool),
			repository.NewUserRepo(pool),
			repository.NewMemberRepo(pool),
			service.NewWASender(),
		),
		Audit: service.NewAuditService(
			repository.NewAuditRepo(pool),
			repository.NewUserAdminRepo(pool),
		),
		User: service.NewUserService(
			repository.NewUserAdminRepo(pool),
			repository.NewUserRepo(pool),
			repository.NewMemberRepo(pool),
			service.NewAuditService(repository.NewAuditRepo(pool), repository.NewUserAdminRepo(pool)),
		),
		Profile: service.NewProfileService(
			repository.NewMemberRepo(pool),
			repository.NewAttendanceRepo(pool),
			repository.NewMonitoringRepo(pool),
			repository.NewMeetingRepo(pool),
		),
		AI: service.NewAIService(
			omniRoute,
			repository.NewAIRepo(pool),
			service.NewAIToolExecutor(
				service.NewDashboardService(
					repository.NewMemberRepo(pool),
					repository.NewMeetingRepo(pool),
					repository.NewAttendanceRepo(pool),
					repository.NewMonitoringRepo(pool),
				),
				service.NewMemberService(repository.NewMemberRepo(pool)),
				service.NewGroupService(repository.NewGroupRepo(pool)),
				service.NewMeetingService(repository.NewMeetingRepo(pool)),
				service.NewAttendanceService(
					repository.NewAttendanceRepo(pool),
					repository.NewMeetingRepo(pool),
					service.NewMemberService(repository.NewMemberRepo(pool)),
				),
				service.NewMonitoringService(repository.NewMonitoringRepo(pool)),
				service.NewAnnouncementService(
					repository.NewAnnouncementRepo(pool),
					repository.NewGroupRepo(pool),
					repository.NewMemberRepo(pool),
				),
				service.NewProfileService(
					repository.NewMemberRepo(pool),
					repository.NewAttendanceRepo(pool),
					repository.NewMonitoringRepo(pool),
					repository.NewMeetingRepo(pool),
				),
			),
		),
	}
}

func RegisterAPI(app *fiber.App, svc *Services) {
	app.Post("/api", BodyParserMiddleware(), AuthMiddleware(svc.Auth), func(c *fiber.Ctx) error {
		action, _ := BodyOf(c)["action"].(string)
		if action == "" {
			return Fail(c, "Parameter action wajib diisi")
		}

		switch action {
		// Auth
		case "login":
			return handleLogin(c, svc.Auth)
		case "logout":
			return handleLogout(c, svc.Auth)
		case "validateSession":
			return handleValidateSession(c, svc.Auth)

		// Members
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

		// Groups
		case "getGroups":
			return handleGetGroups(c, svc.Group)

		// Meetings
		case "getMeetings":
			return handleGetMeetings(c, svc.Meeting)
		case "createMeeting":
			return handleCreateMeeting(c, svc.Meeting)
		case "updateMeeting":
			return handleUpdateMeeting(c, svc.Meeting)
		case "deleteMeeting":
			return handleDeleteMeeting(c, svc.Meeting)

		// Bulk Meeting
		case "previewBulkMeetings":
			return handlePreviewBulkMeetings(c, svc.BulkMeeting)
		case "bulkCreateMeetings":
			return handleBulkCreateMeetings(c, svc.BulkMeeting)
		case "getBulkMeetingTemplates":
			return handleGetBulkMeetingTemplates(c, svc.BulkMeeting)

		// Attendance
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

		// Dashboard & Monitoring
		case "getDashboard":
			return handleGetDashboard(c, svc.Dashboard)
		case "getMyDashboard":
			return handleGetMyDashboard(c, svc.Dashboard)
		case "getMonitoring":
			return handleGetMonitoring(c, svc.Monitoring)

		// Settings
		case "getSettings":
			return handleGetSettings(c, svc.Settings)
		case "updateSettings":
			return handleUpdateSettings(c, svc.Settings)

		// Announcements
		case "getAnnouncementTemplates":
			return handleGetAnnouncementTemplates(c, svc.Announcement)
		case "getAllAnnouncementTemplates":
			return handleGetAllAnnouncementTemplates(c, svc.Announcement)
		case "getAnnouncementTemplateDetail":
			return handleGetAnnouncementTemplateDetail(c, svc.Announcement)
		case "createAnnouncementTemplate":
			return handleCreateAnnouncementTemplate(c, svc.Announcement)
		case "updateAnnouncementTemplate":
			return handleUpdateAnnouncementTemplate(c, svc.Announcement)
		case "deleteAnnouncementTemplate":
			return handleDeleteAnnouncementTemplate(c, svc.Announcement)
		case "createTemplateFromAnnouncement":
			return handleCreateTemplateFromAnnouncement(c, svc.Announcement)
		case "generateAnnouncement":
			return handleGenerateAnnouncement(c, svc.Announcement)
		case "createAnnouncement":
			return handleCreateAnnouncement(c, svc.Announcement)
		case "updateAnnouncement":
			return handleUpdateAnnouncement(c, svc.Announcement)
		case "getAnnouncements":
			return handleGetAnnouncements(c, svc.Announcement)
		case "getAnnouncementRecipientSummary":
			return handleGetAnnouncementRecipientSummary(c, svc.Announcement)

		// Pending
		case "checkUsernameAvailability":
			return handleCheckUsernameAvailability(c, svc.Pending)
		case "submitPublicRegistration":
			return handleSubmitPublicRegistration(c, svc.Pending)
		case "getPendingMembers":
			return handleGetPendingMembers(c, svc.Pending)
		case "getPendingMemberDetail":
			return handleGetPendingMemberDetail(c, svc.Pending)
		case "approvePendingMember":
			return handleApprovePendingMember(c, svc.Pending)
		case "rejectPendingMember":
			return handleRejectPendingMember(c, svc.Pending)

		// Users
		case "getUsers":
			return handleGetUsers(c, svc.User)
		case "getUserDetail":
			return handleGetUserDetail(c, svc.User)
		case "createUser":
			return handleCreateUser(c, svc.User)
		case "updateUser":
			return handleUpdateUser(c, svc.User)
		case "updateUserRole":
			return handleUpdateUserRole(c, svc.User)
		case "getMemberUserStatus":
			return handleGetMemberUserStatus(c, svc.User)
		case "changeMyPassword":
			return handleChangeMyPassword(c, svc.User)
		case "resetUserPassword":
			return handleResetUserPassword(c, svc.User)

		// Profile
		case "getMyProfile":
			return handleGetMyProfile(c, svc.Profile)
		case "updateMyProfile":
			return handleUpdateMyProfile(c, svc.Profile)
		case "getMyAttendance":
			return handleGetMyAttendance(c, svc.Profile)
		case "getMyMonitoring":
			return handleGetMyMonitoring(c, svc.Profile)
		case "getUpcomingMeetings":
			return handleGetUpcomingMeetings(c, svc.Profile)

			// Audit
		case "getAuditLogs":
			return handleGetAuditLogs(c, svc.Audit)

		// AI
		case "aiChat":
			return handleAiChat(c, svc.AI)
		case "getAiUsageStats":
			return handleGetAiUsageStats(c, svc.AI)

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
