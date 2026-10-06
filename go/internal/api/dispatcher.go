package api

import (
	"context"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

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

	"getCashLedger":          true,
	"saveCashLedger":         true,
	"deleteCashLedger":       true,
	"duplicateCashLedger":    true,
	"carryForwardCashLedger": true,
	"getMonthlyDues":         true,
	"saveDueMember":          true,
	"deleteDueMember":        true,
	"saveDuePayment":         true,
	"reverseDuePayment":      true,
	"getDueLastNominals":     true,
	"getZakatRecords":        true,
	"getZakatDetail":         true,
	"saveZakatRecord":        true,
	"saveZakatPayers":        true,
	"saveZakatRecipients":    true,
	"saveZakatAllocations":   true,
	"updateZakatStatus":      true,
	"deleteZakatRecord":      true,
	"getZakatMasters":        true,
	"addZakatMaster":         true,
	"postDueToCash":          true,
	"cancelPostDueToCash":    true,
	"runFinanceSync":         true,
	"runFinanceImport":       true,
	"getTilawatiEditorList":  true,
	"publishTilawatiEditor":  true,
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
	Finance        *service.FinanceService
	FinanceSync    *service.FinanceSyncService
	TilawatiEditor *service.TilawatiEditorService
}

func NewServices(pool *pgxpool.Pool, authSvc *auth.Service, providers map[string]ai.Provider, providerOrder []string, storage *service.StorageService) *Services {
	financeSvc := service.NewFinanceService(repository.NewFinanceRepo(pool))
	financeSyncSvc := service.NewFinanceSyncService(
		repository.NewFinanceRepo(pool),
		repository.NewGroupRepo(pool),
		financeSvc,
	)
	financeSvc.AfterWrite = func(groupID string) {
		if !financeSyncSvc.IsConfigured() {
			return
		}
		if err := financeSyncSvc.SyncGroup(context.Background(), groupID); err != nil {
			log.Warn().Err(err).Str("group", groupID).Msg("finance sheet push gagal")
		}
	}

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
			financeSvc,
		),
		repository.NewSettingsRepo(pool),
	)

	fonnteSvc := service.NewFonnteService()

	return &Services{
		Auth: authSvc,
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
		Pending: func() *service.PendingService {
			s := service.NewPendingService(
				repository.NewPendingRepo(pool),
				repository.NewUserRepo(pool),
				repository.NewMemberRepo(pool),
			)
			s.SetGroupRepo(repository.NewGroupRepo(pool))
			return s
		}(),
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
					financeSvc,
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
		Finance:     financeSvc,
		FinanceSync: financeSyncSvc,
		FridayReminder: service.NewFridayReminderService(
			repository.NewFridayRepo(pool),
			fonnteSvc,
			repository.NewSettingsRepo(pool),
		),
		TilawatiEditor: service.NewTilawatiEditorService(
			pool,
			repository.NewRepository(pool),
			storage,
			"/app", // Di Docker aslinya ada di /app, namun service butuh frontendDir untuk public/audio
		),
	}
}

func RegisterAPI(app *fiber.App, svc *Services) {
	RegisterRESTAPI(app, svc)
	RegisterTilawatiEditorAPI(app, svc, AuthMiddleware(svc.Auth))

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

		case "getCashLedger":
			return handleCashLedger(c, svc.Finance)
		case "saveCashLedger":
			return handleCashSave(c, svc.Finance)
		case "deleteCashLedger":
			return handleCashDelete(c, svc.Finance)
		case "duplicateCashLedger":
			return handleCashDuplicate(c, svc.Finance)
		case "carryForwardCashLedger":
			return handleCashCarryForward(c, svc.Finance)
		case "getMonthlyDues":
			return handleDuesData(c, svc.Finance)
		case "saveDueMember":
			return handleDueMemberSave(c, svc.Finance)
		case "deleteDueMember":
			return handleDueMemberDelete(c, svc.Finance)
		case "saveDuePayment":
			return handleDuePaymentSave(c, svc.Finance)
		case "reverseDuePayment":
			return handleDuePaymentReverse(c, svc.Finance)
		case "getDueLastNominals":
			return handleDueLastNominals(c, svc.Finance)
		case "getZakatRecords":
			return handleZakatList(c, svc.Finance)
		case "getZakatDetail":
			return handleZakatDetail(c, svc.Finance)
		case "saveZakatRecord":
			return handleZakatSave(c, svc.Finance)
		case "saveZakatPayers":
			return handleZakatSavePayers(c, svc.Finance)
		case "saveZakatRecipients":
			return handleZakatSaveRecipients(c, svc.Finance)
		case "saveZakatAllocations":
			return handleZakatSaveAllocations(c, svc.Finance)
		case "updateZakatStatus":
			return handleZakatStatus(c, svc.Finance)
		case "deleteZakatRecord":
			return handleZakatDelete(c, svc.Finance)
		case "getZakatMasters":
			return handleZakatMasters(c, svc.Finance)
		case "addZakatMaster":
			return handleZakatAddMaster(c, svc.Finance)
		case "postDueToCash":
			return handleDuePostToCash(c, svc.Finance)
		case "cancelPostDueToCash":
			return handleDueCancelPostToCash(c, svc.Finance)
		case "runFinanceSync":
			return handleFinanceSync(c, svc.FinanceSync)
		case "runFinanceImport":
			return handleFinanceImport(c, svc.FinanceSync)

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

func RegisterRESTAPI(app *fiber.App, svc *Services) {
	v1 := app.Group("/api/v1", BodyParserMiddleware(), AuthMiddleware(svc.Auth), AuditMiddleware(svc.Audit))

	// Auth & Profile
	v1.Post("/auth/login", func(c *fiber.Ctx) error { return handleLogin(c, svc.Auth, svc.Audit) })
	v1.Post("/auth/logout", func(c *fiber.Ctx) error { return handleLogout(c, svc.Auth) })
	v1.Get("/auth/validate-session", func(c *fiber.Ctx) error { return handleValidateSession(c, svc.Auth) })
	v1.Get("/auth/me", func(c *fiber.Ctx) error { return handleGetMyProfile(c, svc.Profile) })
	v1.Put("/auth/me", func(c *fiber.Ctx) error { return handleUpdateMyProfile(c, svc.Profile) })
	v1.Get("/auth/my-attendance", func(c *fiber.Ctx) error { return handleGetMyAttendance(c, svc.Profile) })
	v1.Get("/auth/my-monitoring", func(c *fiber.Ctx) error { return handleGetMyMonitoring(c, svc.Profile) })
	v1.Get("/auth/upcoming-meetings", func(c *fiber.Ctx) error { return handleGetUpcomingMeetings(c, svc.Profile) })
	v1.Post("/auth/change-password", func(c *fiber.Ctx) error { return handleChangeMyPassword(c, svc.User) })
	v1.Post("/auth/change-username", func(c *fiber.Ctx) error { return handleChangeMyUsername(c, svc.User) })

	// Groups
	v1.Get("/groups", func(c *fiber.Ctx) error { return handleGetGroups(c, svc.Group) })
	v1.Post("/groups", func(c *fiber.Ctx) error { return handleSaveGroup(c, svc.Group) })

	// Members
	v1.Get("/members", func(c *fiber.Ctx) error { return handleGetMembers(c, svc.Member) })
	v1.Get("/members/paged", func(c *fiber.Ctx) error { return handleGetMembersPaged(c, svc.Member) })
	v1.Get("/members/pnkb", func(c *fiber.Ctx) error { return handleGetPNKBMembers(c, svc.Member) })
	v1.Get("/members/pnkb-paged", func(c *fiber.Ctx) error { return handleGetPNKBMembersPaged(c, svc.Member) })
	v1.Get("/members/attendance", func(c *fiber.Ctx) error { return handleGetAttendanceMembers(c, svc.Member) })
	v1.Get("/members/export", func(c *fiber.Ctx) error { return handleGetMembersForExport(c, svc.Member) })
	v1.Get("/members/detail", func(c *fiber.Ctx) error { return handleGetMemberDetail(c, svc.Member) })
	v1.Post("/members", func(c *fiber.Ctx) error { return handleCreateMember(c, svc.Member) })
	v1.Put("/members", func(c *fiber.Ctx) error { return handleUpdateMember(c, svc.Member) })
	v1.Post("/members/deactivate", func(c *fiber.Ctx) error { return handleDeactivateMember(c, svc.Member) })
	v1.Delete("/members", func(c *fiber.Ctx) error { return handleDeleteMember(c, svc.Member) })

	// Meetings & Bulk
	v1.Get("/meetings", func(c *fiber.Ctx) error { return handleGetMeetings(c, svc.Meeting) })
	v1.Post("/meetings", func(c *fiber.Ctx) error { return handleCreateMeeting(c, svc.Meeting) })
	v1.Put("/meetings", func(c *fiber.Ctx) error { return handleUpdateMeeting(c, svc.Meeting) })
	v1.Delete("/meetings", func(c *fiber.Ctx) error { return handleDeleteMeeting(c, svc.Meeting) })
	v1.Delete("/meetings/bulk", func(c *fiber.Ctx) error { return handleDeleteMeetingsBulk(c, svc.Meeting) })
	v1.Post("/meetings/bulk-preview", func(c *fiber.Ctx) error { return handlePreviewBulkMeetings(c, svc.BulkMeeting) })
	v1.Post("/meetings/bulk-create", func(c *fiber.Ctx) error { return handleBulkCreateMeetings(c, svc.BulkMeeting) })
	v1.Get("/meetings/bulk-templates", func(c *fiber.Ctx) error { return handleGetBulkMeetingTemplates(c, svc.BulkMeeting) })
	v1.Post("/meetings/parse-pdf", func(c *fiber.Ctx) error { return handleParsePdfMeeting(c, svc.PDFImport) })

	// Attendance
	v1.Get("/attendance", func(c *fiber.Ctx) error { return handleGetAttendance(c, svc.Attendance) })
	v1.Get("/attendance/page", func(c *fiber.Ctx) error { return handleGetAttendancePage(c, svc.Attendance) })
	v1.Post("/attendance", func(c *fiber.Ctx) error { return handleSaveAttendance(c, svc.Attendance) })
	v1.Post("/attendance/bulk", func(c *fiber.Ctx) error { return handleBulkSaveAttendance(c, svc.Attendance) })
	v1.Delete("/attendance", func(c *fiber.Ctx) error { return handleDeleteAttendance(c, svc.Attendance) })
	v1.Delete("/attendance/meeting", func(c *fiber.Ctx) error { return handleDeleteAttendanceByMeeting(c, svc.Attendance) })
	v1.Delete("/attendance/member", func(c *fiber.Ctx) error { return handleDeleteAttendanceByMember(c, svc.Attendance) })

	// Dashboard, Monitoring & Settings
	v1.Get("/dashboard", func(c *fiber.Ctx) error { return handleGetDashboard(c, svc.Dashboard) })
	v1.Get("/dashboard/my", func(c *fiber.Ctx) error { return handleGetMyDashboard(c, svc.Dashboard) })
	v1.Get("/monitoring", func(c *fiber.Ctx) error { return handleGetMonitoring(c, svc.Monitoring) })
	v1.Post("/monitoring", func(c *fiber.Ctx) error { return handleCreateMonitoring(c, svc.Monitoring) })
	v1.Put("/monitoring", func(c *fiber.Ctx) error { return handleUpdateMonitoring(c, svc.Monitoring) })
	v1.Get("/settings", func(c *fiber.Ctx) error { return handleGetSettings(c, svc.Settings) })
	v1.Put("/settings", func(c *fiber.Ctx) error { return handleUpdateSettings(c, svc.Settings) })

	// Announcements
	v1.Get("/announcements", func(c *fiber.Ctx) error { return handleGetAnnouncements(c, svc.Announcement) })
	v1.Post("/announcements", func(c *fiber.Ctx) error { return handleCreateAnnouncement(c, svc.Announcement) })
	v1.Put("/announcements", func(c *fiber.Ctx) error { return handleUpdateAnnouncement(c, svc.Announcement) })
	v1.Get("/announcements/templates", func(c *fiber.Ctx) error { return handleGetAnnouncementTemplates(c, svc.Announcement) })
	v1.Get("/announcements/templates/all", func(c *fiber.Ctx) error { return handleGetAllAnnouncementTemplates(c, svc.Announcement) })
	v1.Get("/announcements/templates/detail", func(c *fiber.Ctx) error { return handleGetAnnouncementTemplateDetail(c, svc.Announcement) })
	v1.Post("/announcements/templates", func(c *fiber.Ctx) error { return handleCreateAnnouncementTemplate(c, svc.Announcement) })
	v1.Put("/announcements/templates", func(c *fiber.Ctx) error { return handleUpdateAnnouncementTemplate(c, svc.Announcement) })
	v1.Delete("/announcements/templates", func(c *fiber.Ctx) error { return handleDeleteAnnouncementTemplate(c, svc.Announcement) })
	v1.Post("/announcements/templates/from-announcement", func(c *fiber.Ctx) error { return handleCreateTemplateFromAnnouncement(c, svc.Announcement) })
	v1.Post("/announcements/generate", func(c *fiber.Ctx) error { return handleGenerateAnnouncement(c, svc.Announcement) })
	v1.Post("/announcements/generate-weekly", func(c *fiber.Ctx) error { return handleGenerateWeeklyAnnouncements(c, svc.Announcement) })
	v1.Get("/announcements/recipient-summary", func(c *fiber.Ctx) error { return handleGetAnnouncementRecipientSummary(c, svc.Announcement) })

	// Public & Pending
	v1.Get("/public/groups", func(c *fiber.Ctx) error { return handleGetPublicGroups(c, svc.Group) })
	v1.Get("/public/check-username", func(c *fiber.Ctx) error { return handleCheckUsernameAvailability(c, svc.Pending) })
	v1.Post("/public/register", func(c *fiber.Ctx) error { return handleSubmitPublicRegistration(c, svc.Pending) })
	v1.Get("/pending", func(c *fiber.Ctx) error { return handleGetPendingMembers(c, svc.Pending) })
	v1.Get("/pending/detail", func(c *fiber.Ctx) error { return handleGetPendingMemberDetail(c, svc.Pending) })
	v1.Post("/pending/approve", func(c *fiber.Ctx) error { return handleApprovePendingMember(c, svc.Pending) })
	v1.Post("/pending/reject", func(c *fiber.Ctx) error { return handleRejectPendingMember(c, svc.Pending) })

	// Users
	v1.Get("/users", func(c *fiber.Ctx) error { return handleGetUsers(c, svc.User) })
	v1.Get("/users/detail", func(c *fiber.Ctx) error { return handleGetUserDetail(c, svc.User) })
	v1.Post("/users", func(c *fiber.Ctx) error { return handleCreateUser(c, svc.User) })
	v1.Put("/users", func(c *fiber.Ctx) error { return handleUpdateUser(c, svc.User) })
	v1.Delete("/users", func(c *fiber.Ctx) error { return handleDeleteUserPermanent(c, svc.User) })
	v1.Put("/users/role", func(c *fiber.Ctx) error { return handleUpdateUserRole(c, svc.User) })
	v1.Get("/users/member-status", func(c *fiber.Ctx) error { return handleGetMemberUserStatus(c, svc.User) })
	v1.Post("/users/reset-password", func(c *fiber.Ctx) error { return handleResetUserPassword(c, svc.User) })

	// Moods
	v1.Post("/moods", func(c *fiber.Ctx) error { return handleSaveMood(c, svc.Mood) })
	v1.Get("/moods/my", func(c *fiber.Ctx) error { return handleGetMyMoods(c, svc.Mood) })
	v1.Get("/moods/member", func(c *fiber.Ctx) error { return handleGetMemberMoods(c, svc.Mood) })

	// Member Requests
	v1.Post("/member-requests/become", func(c *fiber.Ctx) error { return handleRequestBecomeMember(c, svc.MemberReq) })
	v1.Get("/member-requests", func(c *fiber.Ctx) error { return handleListMemberRequests(c, svc.MemberReq) })
	v1.Post("/member-requests/approve", func(c *fiber.Ctx) error { return handleApproveMemberRequest(c, svc.MemberReq) })
	v1.Post("/member-requests/reject", func(c *fiber.Ctx) error { return handleRejectMemberRequest(c, svc.MemberReq) })

	// Audit & AI
	v1.Get("/audit-logs", func(c *fiber.Ctx) error { return handleGetAuditLogs(c, svc.Audit) })
	v1.Post("/ai/chat", func(c *fiber.Ctx) error { return handleAiChat(c, svc.AI) })
	v1.Get("/ai/usage", func(c *fiber.Ctx) error { return handleGetAiUsageStats(c, svc.AI) })
	v1.Get("/ai/provider", func(c *fiber.Ctx) error { return handleGetCurrentProvider(c, svc.AI) })
	v1.Post("/ai/provider", func(c *fiber.Ctx) error { return handleSetAIProvider(c, svc.AI) })

	// Photo
	v1.Post("/photo", func(c *fiber.Ctx) error { return svc.Photo.Upload(c) })
	v1.Delete("/photo", func(c *fiber.Ctx) error { return svc.Photo.Delete(c) })

	// Finance (SabilKas, per-group, auth wajib + scoping group_label)
	fin := func(action string, h fiber.Handler) []fiber.Handler {
		return []fiber.Handler{FinanceAuthMiddleware(svc.Auth, action), h}
	}
	v1.Get("/finance/cash-ledger", fin("getCashLedger", func(c *fiber.Ctx) error { return handleCashLedger(c, svc.Finance) })...)
	v1.Post("/finance/cash-ledger", fin("saveCashLedger", func(c *fiber.Ctx) error { return handleCashSave(c, svc.Finance) })...)
	v1.Delete("/finance/cash-ledger", fin("deleteCashLedger", func(c *fiber.Ctx) error { return handleCashDelete(c, svc.Finance) })...)
	v1.Post("/finance/cash-ledger/duplicate", fin("duplicateCashLedger", func(c *fiber.Ctx) error { return handleCashDuplicate(c, svc.Finance) })...)
	v1.Post("/finance/cash-ledger/carry-forward", fin("carryForwardCashLedger", func(c *fiber.Ctx) error { return handleCashCarryForward(c, svc.Finance) })...)
	v1.Get("/finance/monthly-dues", fin("getMonthlyDues", func(c *fiber.Ctx) error { return handleDuesData(c, svc.Finance) })...)
	v1.Post("/finance/monthly-dues/members", fin("saveDueMember", func(c *fiber.Ctx) error { return handleDueMemberSave(c, svc.Finance) })...)
	v1.Delete("/finance/monthly-dues/members", fin("deleteDueMember", func(c *fiber.Ctx) error { return handleDueMemberDelete(c, svc.Finance) })...)
	v1.Post("/finance/monthly-dues/payments", fin("saveDuePayment", func(c *fiber.Ctx) error { return handleDuePaymentSave(c, svc.Finance) })...)
	v1.Post("/finance/monthly-dues/payments/reverse", fin("reverseDuePayment", func(c *fiber.Ctx) error { return handleDuePaymentReverse(c, svc.Finance) })...)
	v1.Get("/finance/monthly-dues/last-nominals", fin("getDueLastNominals", func(c *fiber.Ctx) error { return handleDueLastNominals(c, svc.Finance) })...)
	v1.Post("/finance/monthly-dues/post-to-kas", fin("postDueToCash", func(c *fiber.Ctx) error { return handleDuePostToCash(c, svc.Finance) })...)
	v1.Post("/finance/monthly-dues/cancel-post-to-kas", fin("cancelPostDueToCash", func(c *fiber.Ctx) error { return handleDueCancelPostToCash(c, svc.Finance) })...)
	v1.Get("/finance/zakat", fin("getZakatRecords", func(c *fiber.Ctx) error { return handleZakatList(c, svc.Finance) })...)
	v1.Get("/finance/zakat/detail", fin("getZakatDetail", func(c *fiber.Ctx) error { return handleZakatDetail(c, svc.Finance) })...)
	v1.Post("/finance/zakat", fin("saveZakatRecord", func(c *fiber.Ctx) error { return handleZakatSave(c, svc.Finance) })...)
	v1.Post("/finance/zakat/payers", fin("saveZakatPayers", func(c *fiber.Ctx) error { return handleZakatSavePayers(c, svc.Finance) })...)
	v1.Post("/finance/zakat/recipients", fin("saveZakatRecipients", func(c *fiber.Ctx) error { return handleZakatSaveRecipients(c, svc.Finance) })...)
	v1.Post("/finance/zakat/allocations", fin("saveZakatAllocations", func(c *fiber.Ctx) error { return handleZakatSaveAllocations(c, svc.Finance) })...)
	v1.Get("/finance/zakat/masters", fin("getZakatMasters", func(c *fiber.Ctx) error { return handleZakatMasters(c, svc.Finance) })...)
	v1.Post("/finance/zakat/masters", fin("addZakatMaster", func(c *fiber.Ctx) error { return handleZakatAddMaster(c, svc.Finance) })...)
	v1.Put("/finance/zakat/status", fin("updateZakatStatus", func(c *fiber.Ctx) error { return handleZakatStatus(c, svc.Finance) })...)
	v1.Delete("/finance/zakat", fin("deleteZakatRecord", func(c *fiber.Ctx) error { return handleZakatDelete(c, svc.Finance) })...)
	v1.Post("/finance/sync", fin("runFinanceSync", func(c *fiber.Ctx) error { return handleFinanceSync(c, svc.FinanceSync) })...)
	v1.Post("/finance/import", fin("runFinanceImport", func(c *fiber.Ctx) error { return handleFinanceImport(c, svc.FinanceSync) })...)

	// Friday
	v1.Get("/friday", func(c *fiber.Ctx) error { return handleGetFridaySchedules(c, svc.Friday) })
	v1.Post("/friday", func(c *fiber.Ctx) error { return handleSaveFridaySchedule(c, svc.Friday) })
	v1.Delete("/friday", func(c *fiber.Ctx) error { return handleDeleteFridaySchedule(c, svc.Friday) })
	v1.Get("/friday/reminder-status", func(c *fiber.Ctx) error { return handleGetFridayReminderStatus(c, svc.FridayReminder) })
	v1.Post("/friday/mark-reminder-sent", func(c *fiber.Ctx) error { return handleMarkFridayReminderSent(c, svc.FridayReminder) })
}
