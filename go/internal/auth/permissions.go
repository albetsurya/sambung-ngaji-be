package auth

var rolePermissions = map[string][]string{
	"logout":           {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"validateSession":  {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"changeMyPassword": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"changeMyUsername": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	"getDashboard":   {"ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS"},
	"getMyDashboard": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	"getMembers":           {"ADMIN", "PENGAWAS"},
	"getMembersPaged":      {"ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"getPNKBMembers":       {"TIM_PNKB"},
	"getPNKBMembersPaged":  {"TIM_PNKB"},
	"getMembersForExport":  {"ADMIN", "TIM_PNKB"},
	"getAttendanceMembers": {"ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"getMemberDetail":      {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "PENGAWAS"},
	"createMember":         {"ADMIN"},
	"updateMember":         {"ADMIN"},
	"deactivateMember":     {"ADMIN"},
	"deleteMember":         {"SUPER_ADMIN", "ADMIN"},
	"getMemberUserStatus":  {"ADMIN"},
	"getUserDetail":        {"SUPER_ADMIN", "ADMIN"},

	"getGroups": {"ADMIN", "PENGAWAS"},
	"saveGroup": {"ADMIN"},

	"getMeetings":        {"ADMIN", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"createMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"updateMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"deleteMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"deleteMeetingsBulk": {"ADMIN", "TIM_ABSENSI"},

	"previewBulkMeetings":     {"ADMIN"},
	"bulkCreateMeetings":      {"ADMIN"},
	"getBulkMeetingTemplates": {"ADMIN"},

	"getAttendance":             {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"getAttendancePage":         {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"saveAttendance":            {"ADMIN", "TIM_ABSENSI"},
	"bulkSaveAttendance":        {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendance":          {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendanceByMeeting": {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendanceByMember":  {"ADMIN", "TIM_ABSENSI"},

	"getMonitoring":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "PENGAWAS"},
	"createMonitoring": {"ADMIN", "TIM_PNKB", "PENGAWAS"},
	"updateMonitoring": {"ADMIN", "TIM_PNKB", "PENGAWAS"},

	"getAnnouncementTemplates":        {"ADMIN", "PENGAWAS"},
	"getAllAnnouncementTemplates":     {"ADMIN"},
	"getAnnouncementTemplateDetail":   {"ADMIN"},
	"createAnnouncementTemplate":      {"ADMIN"},
	"updateAnnouncementTemplate":      {"ADMIN"},
	"deleteAnnouncementTemplate":      {"ADMIN"},
	"createTemplateFromAnnouncement":  {"ADMIN"},
	"getAnnouncements":                {"ADMIN", "PENGAWAS"},
	"createAnnouncement":              {"ADMIN"},
	"updateAnnouncement":              {"ADMIN"},
	"generateAnnouncement":            {"ADMIN"},
	"generateWeeklyAnnouncements":     {"ADMIN"},
	"getAnnouncementRecipientSummary": {"ADMIN"},

	"uploadPhoto":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"deletePhoto":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getSettings":    {"ADMIN", "TIM_ABSENSI", "TIM_PNKB", "PENGAWAS"},
	"updateSettings": {"SUPER_ADMIN"},

	"getUsers":            {"SUPER_ADMIN", "ADMIN"},
	"createUser":          {"SUPER_ADMIN", "ADMIN"},
	"updateUser":          {"SUPER_ADMIN", "ADMIN"},
	"updateUserRole":      {"SUPER_ADMIN", "ADMIN"},
	"deleteUserPermanent": {"SUPER_ADMIN"},
	"resetUserPassword":   {"SUPER_ADMIN", "ADMIN"},

	"getAuditLogs": {"SUPER_ADMIN"},

	"aiChat":             {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getCurrentProvider": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"setAIProvider":      {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI"},

	"getAiUsageStats": {"SUPER_ADMIN"},

	"parsePdfMeeting": {"ADMIN", "TIM_ABSENSI"},

	"getPendingMembers":      {"ADMIN"},
	"getPendingMemberDetail": {"ADMIN"},
	"approvePendingMember":   {"ADMIN"},
	"rejectPendingMember":    {"ADMIN"},

	"getMyProfile":        {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"updateMyProfile":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMyAttendance":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMyMonitoring":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getUpcomingMeetings": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	"saveMood":       {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMyMoods":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMemberMoods": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "PENGAWAS"},

	"requestBecomeMember":  {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMemberRequests":    {"SUPER_ADMIN", "ADMIN"},
	"approveMemberRequest": {"SUPER_ADMIN", "ADMIN"},
	"rejectMemberRequest":  {"SUPER_ADMIN", "ADMIN"},

	"getFinanceKas":                 {"SUPER_ADMIN", "TIM_KU"},
	"saveFinanceKas":                {"SUPER_ADMIN", "TIM_KU"},
	"deleteFinanceKas":              {"SUPER_ADMIN", "TIM_KU"},
	"duplicateFinanceKas":           {"SUPER_ADMIN", "TIM_KU"},
	"carryForwardFinanceKas":        {"SUPER_ADMIN", "TIM_KU"},
	"getFinanceShodaqoh":            {"SUPER_ADMIN", "TIM_KU"},
	"saveFinanceShodaqohMember":     {"SUPER_ADMIN", "TIM_KU"},
	"deleteFinanceShodaqohMember":   {"SUPER_ADMIN", "TIM_KU"},
	"saveFinanceShodaqohPayment":    {"SUPER_ADMIN", "TIM_KU"},
	"reverseFinanceShodaqohPayment": {"SUPER_ADMIN", "TIM_KU"},
	"getFinanceShodaqohNominals":    {"SUPER_ADMIN", "TIM_KU"},
	"getFinanceZakat":               {"SUPER_ADMIN", "TIM_KU"},
	"saveFinanceZakat":              {"SUPER_ADMIN", "TIM_KU"},
	"updateFinanceZakatStatus":      {"SUPER_ADMIN", "TIM_KU"},
	"deleteFinanceZakat":            {"SUPER_ADMIN", "TIM_KU"},

	"getFridaySchedules":      {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"saveFridaySchedule":      {"ADMIN", "TIM_ABSENSI"},
	"deleteFridaySchedule":    {"ADMIN", "TIM_ABSENSI"},
	"getFridayReminderStatus": {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI"},
	"markFridayReminderSent":  {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI"},

	"createWaQueue":        {"ADMIN", "TIM_ABSENSI"},
	"cancelWaQueue":        {"ADMIN", "TIM_ABSENSI"},
	"getWaQueueStatus":     {"ADMIN", "TIM_ABSENSI"},
	"bulkGetWaQueueStatus": {"ADMIN", "TIM_ABSENSI"},
	"listWaQueue":          {"ADMIN", "TIM_ABSENSI"},
	"retryWaQueue":         {"ADMIN", "TIM_ABSENSI"},
}

func CanAccess(role, action string) bool {
	if role == "SUPER_ADMIN" {
		return true
	}
	roles, ok := rolePermissions[action]
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

var SuperAdminOnlyActions = map[string]bool{
	"deleteUserPermanent": true,
	"getAuditLogs":        true,
	"updateSettings":      true,
}

func IsSuperAdminOnly(action string) bool {
	return SuperAdminOnlyActions[action]
}
