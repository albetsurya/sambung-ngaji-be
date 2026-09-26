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

	"getMembers":           true,
	"getMembersPaged":      true,
	"getPNKBMembers":       true,
	"getPNKBMembersPaged":  true,
	"getAttendanceMembers": true,
	"getMemberDetail":      true,

	"createMember":        true,
	"updateMember":        true,
	"deactivateMember":    true,
	"deleteMember":        true,
	"getMembersForExport": true,

	"getGroups": true,

	"getMeetings":        true,
	"createMeeting":      true,
	"updateMeeting":      true,
	"deleteMeeting":      true,
	"deleteMeetingsBulk": true,

	"previewBulkMeetings":     true,
	"bulkCreateMeetings":      true,
	"getBulkMeetingTemplates": true,

	"getAttendance":             true,
	"getAttendancePage":         true,
	"saveAttendance":            true,
	"bulkSaveAttendance":        true,
	"deleteAttendance":          true,
	"deleteAttendanceByMeeting": true,
	"deleteAttendanceByMember":  true,

	"getDashboard":   true,
	"getMyDashboard": true,
	"getMonitoring":  true,

	"getSettings":    true,
	"updateSettings": true,

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
	"generateWeeklyAnnouncements":     true,

	"saveGroup": true,

	"createMonitoring": true,
	"updateMonitoring": true,

	"getPendingMembers":         true,
	"getPendingMemberDetail":    true,
	"approvePendingMember":      true,
	"rejectPendingMember":       true,
	"submitPublicRegistration":  true,
	"checkUsernameAvailability": true,

	"getUsers":            true,
	"getUserDetail":       true,
	"createUser":          true,
	"updateUser":          true,
	"updateUserRole":      true,
	"deleteUserPermanent": true,
	"getMemberUserStatus": true,
	"changeMyPassword":    true,
	"changeMyUsername":    true,
	"resetUserPassword":   true,

	"getMyProfile":        true,
	"updateMyProfile":     true,
	"getMyAttendance":     true,
	"getMyMonitoring":     true,
	"getUpcomingMeetings": true,

	"saveMood":       true,
	"getMyMoods":     true,
	"getMemberMoods": true,

	"requestBecomeMember":  true,
	"getMemberRequests":    true,
	"approveMemberRequest": true,
	"rejectMemberRequest":  true,

	"getAuditLogs": true,

	"aiChat":             true,
	"getAiUsageStats":    true,
	"getCurrentProvider": true,
	"setAIProvider":      true,

	"parsePdfMeeting": true,

	"getFridaySchedules":      true,
	"getFridayReminderStatus": true,
	"markFridayReminderSent":  true,
	"saveFridaySchedule":      true,
	"deleteFridaySchedule":    true,
}

type Services struct {
	Auth           *auth.Service
	Member         *service.MemberService
	Group          *service.GroupService
	Meeting        *service.MeetingService
	BulkMeeting    *service.BulkMeetingService
	Attendance     *service.AttendanceService
	Monitoring     *service.MonitoringService
	Dashboard      *service.DashboardService
	Settings       *service.SettingsService
	Announcement   *service.AnnouncementService
	Pending        *service.PendingService
	User           *service.UserService
	Profile        *service.ProfileService
	Mood           *service.MoodService
	MemberReq      *service.MemberRequestService
	Audit          *service.AuditService
	AI             *service.AIService
	Photo          *PhotoHandler
	PDFImport      *service.PDFImportService
	Fonnte         *service.FonnteService
	Reminder       *service.ReminderService
	Friday         *service.FridayService
	FridayReminder *service.FridayReminderService
}

func NewServices(pool *pgxpool.Pool, authSvc *auth.Service, providers map[string]ai.Provider, providerOrder []string, storage *service.StorageService) *Services {
	aiSvc := service.NewAIService(
		providers,
		providerOrder,
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
			service.NewMonitoringService(repository.NewMonitoringRepo(pool), repository.NewMemberRepo(pool)),
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
		repository.NewSettingsRepo(pool),
	)

	fonnteSvc := service.NewFonnteService()

	return &Services{
		Auth:    authSvc,
		Member: func() *service.MemberService {
			s := service.NewMemberService(repository.NewMemberRepo(pool))
			s.SetGroupRepo(repository.NewGroupRepo(pool))
			return s
		}(),
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
		Monitoring: service.NewMonitoringService(
			repository.NewMonitoringRepo(pool),
			repository.NewMemberRepo(pool),
		),
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
		Mood: service.NewMoodService(
			repository.NewMoodRepo(pool),
			repository.NewMemberRepo(pool),
		),
		MemberReq: service.NewMemberRequestService(
			repository.NewMemberRequestRepo(pool),
			repository.NewMemberRepo(pool),
			repository.NewUserAdminRepo(pool),
		),
		AI: aiSvc,
		Photo: NewPhotoHandler(
			storage,
			service.NewMemberService(repository.NewMemberRepo(pool)),
			service.NewProfileService(
				repository.NewMemberRepo(pool),
				repository.NewAttendanceRepo(pool),
				repository.NewMonitoringRepo(pool),
				repository.NewMeetingRepo(pool),
			),
		),
		PDFImport: service.NewPDFImportService(
			service.NewAIService(
				providers,
				providerOrder,
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
					service.NewMonitoringService(
						repository.NewMonitoringRepo(pool),
						repository.NewMemberRepo(pool),
					),
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
				repository.NewSettingsRepo(pool),
			),
		),
		Fonnte: fonnteSvc,
		Reminder: service.NewReminderService(
			repository.NewMeetingRepo(pool),
			fonnteSvc,
		),
		Friday: service.NewFridayService(
			repository.NewFridayRepo(pool),
		),
		FridayReminder: service.NewFridayReminderService(
			repository.NewFridayRepo(pool),
			fonnteSvc,
			repository.NewSettingsRepo(pool),
		),
	}
}

func RegisterAPI(app *fiber.App, svc *Services) {
	app.Post("/api", BodyParserMiddleware(), AuthMiddleware(svc.Auth), AuditMiddleware(svc.Audit), func(c *fiber.Ctx) error {
		action, _ := BodyOf(c)["action"].(string)
		if action == "" {
			return Fail(c, "Parameter action wajib diisi")
		}

		switch action {
		case "login":
			return handleLogin(c, svc.Auth, svc.Audit)
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
		case "createMember":
			return handleCreateMember(c, svc.Member)
		case "updateMember":
			return handleUpdateMember(c, svc.Member)
		case "deactivateMember":
			return handleDeactivateMember(c, svc.Member)
		case "deleteMember":
			return handleDeleteMember(c, svc.Member)
		case "getMembersForExport":
			return handleGetMembersForExport(c, svc.Member)

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
		case "deleteMeetingsBulk":
			return handleDeleteMeetingsBulk(c, svc.Meeting)

		case "previewBulkMeetings":
			return handlePreviewBulkMeetings(c, svc.BulkMeeting)
		case "bulkCreateMeetings":
			return handleBulkCreateMeetings(c, svc.BulkMeeting)
		case "getBulkMeetingTemplates":
			return handleGetBulkMeetingTemplates(c, svc.BulkMeeting)

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

		case "saveGroup":
			return handleSaveGroup(c, svc.Group)

		case "createMonitoring":
			return handleCreateMonitoring(c, svc.Monitoring)
		case "updateMonitoring":
			return handleUpdateMonitoring(c, svc.Monitoring)

		case "getSettings":
			return handleGetSettings(c, svc.Settings)
		case "updateSettings":
			return handleUpdateSettings(c, svc.Settings)

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
		case "generateWeeklyAnnouncements":
			return handleGenerateWeeklyAnnouncements(c, svc.Announcement)

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

		case "getUsers":
			return handleGetUsers(c, svc.User)
		case "getUserDetail":
			return handleGetUserDetail(c, svc.User)
		case "createUser":
			return handleCreateUser(c, svc.User)
		case "updateUser":
			return handleUpdateUser(c, svc.User)
		case "deleteUserPermanent":
			return handleDeleteUserPermanent(c, svc.User)
		case "updateUserRole":
			return handleUpdateUserRole(c, svc.User)
		case "getMemberUserStatus":
			return handleGetMemberUserStatus(c, svc.User)
		case "changeMyPassword":
			return handleChangeMyPassword(c, svc.User)
		case "changeMyUsername":
			return handleChangeMyUsername(c, svc.User)
		case "resetUserPassword":
			return handleResetUserPassword(c, svc.User)

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

		case "saveMood":
			return handleSaveMood(c, svc.Mood)
		case "getMyMoods":
			return handleGetMyMoods(c, svc.Mood)
		case "getMemberMoods":
			return handleGetMemberMoods(c, svc.Mood)

		case "requestBecomeMember":
			return handleRequestBecomeMember(c, svc.MemberReq)
		case "getMemberRequests":
			return handleListMemberRequests(c, svc.MemberReq)
		case "approveMemberRequest":
			return handleApproveMemberRequest(c, svc.MemberReq)
		case "rejectMemberRequest":
			return handleRejectMemberRequest(c, svc.MemberReq)

		case "getAuditLogs":
			return handleGetAuditLogs(c, svc.Audit)

		case "aiChat":
			return handleAiChat(c, svc.AI)
		case "getAiUsageStats":
			return handleGetAiUsageStats(c, svc.AI)
		case "getCurrentProvider":
			return handleGetCurrentProvider(c, svc.AI)
		case "setAIProvider":
			return handleSetAIProvider(c, svc.AI)

		case "uploadPhoto":
			return svc.Photo.Upload(c)
		case "deletePhoto":
			return svc.Photo.Delete(c)

		case "parsePdfMeeting":
			return handleParsePdfMeeting(c, svc.PDFImport)

		case "getFridaySchedules":
			return handleGetFridaySchedules(c, svc.Friday)
		case "saveFridaySchedule":
			return handleSaveFridaySchedule(c, svc.Friday)
		case "deleteFridaySchedule":
			return handleDeleteFridaySchedule(c, svc.Friday)
		case "getFridayReminderStatus":
			return handleGetFridayReminderStatus(c, svc.FridayReminder)
		case "markFridayReminderSent":
			return handleMarkFridayReminderSent(c, svc.FridayReminder)

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
