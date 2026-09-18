package auth

// ROLE_PERMISSIONS — port dari Apps Script Role_Management.js.
// Map: action → daftar role yang boleh akses.
// SUPER_ADMIN punya akses ke SEMUA action (dicek di CanAccess).
var rolePermissions = map[string][]string{
	// — Sesi & keamanan
	"logout":           {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"validateSession":  {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"changeMyPassword": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"changeMyUsername": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	// — Dashboard
	"getDashboard":   {"ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS"},
	"getMyDashboard": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	// — Members
	"getMembers":           {"ADMIN", "PENGAWAS"},
	"getMembersPaged":      {"ADMIN", "PENGAWAS"},
	"getPNKBMembers":       {"TIM_PNKB"},
	"getPNKBMembersPaged":  {"TIM_PNKB"},
	"getMembersForExport":  {"ADMIN", "TIM_PNKB"},
	"getAttendanceMembers": {"ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"getMemberDetail":      {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "PENGAWAS"},
	"createMember":         {"ADMIN"},
	"updateMember":         {"ADMIN"},
	"deactivateMember":     {"ADMIN"},
	"getMemberUserStatus":  {"ADMIN"},
	"getUserDetail":        {"SUPER_ADMIN", "ADMIN"},

	// — Groups
	"getGroups": {"ADMIN", "PENGAWAS"},
	"saveGroup": {"ADMIN"},

	// — Meetings
	"getMeetings":        {"ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"createMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"updateMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"deleteMeeting":      {"ADMIN", "TIM_ABSENSI"},
	"deleteMeetingsBulk": {"ADMIN", "TIM_ABSENSI"},

	// — Bulk Meeting
	"previewBulkMeetings":     {"ADMIN"},
	"bulkCreateMeetings":      {"ADMIN"},
	"getBulkMeetingTemplates": {"ADMIN"},

	// — Attendance
	"getAttendance":             {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"getAttendancePage":         {"SUPER_ADMIN", "ADMIN", "TIM_ABSENSI", "PENGAWAS"},
	"saveAttendance":            {"ADMIN", "TIM_ABSENSI"},
	"bulkSaveAttendance":        {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendance":          {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendanceByMeeting": {"ADMIN", "TIM_ABSENSI"},
	"deleteAttendanceByMember":  {"ADMIN", "TIM_ABSENSI"},

	// — Monitoring
	"getMonitoring":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "PENGAWAS"},
	"createMonitoring": {"ADMIN", "TIM_PNKB", "PENGAWAS"},
	"updateMonitoring": {"ADMIN", "TIM_PNKB", "PENGAWAS"},

	// — Announcements
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

	// — Foto & pengaturan
	"uploadPhoto":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"deletePhoto":    {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getSettings":    {"ADMIN", "TIM_ABSENSI", "TIM_PNKB", "PENGAWAS"},
	"updateSettings": {"SUPER_ADMIN"},

	// — Users (SUPER_ADMIN only — konsisten dengan SuperAdminOnlyActions)
	"getUsers":          {"SUPER_ADMIN"},
	"createUser":        {"SUPER_ADMIN"},
	"updateUser":        {"SUPER_ADMIN"},
	"updateUserRole":    {"SUPER_ADMIN"},
	"resetUserPassword": {"SUPER_ADMIN"},

	// — Audit
	"getAuditLogs": {"SUPER_ADMIN"},

	// — AI
	"aiChat":             {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getCurrentProvider": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"setAIProvider":      {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI"},

	// — AI usage (admin only)
	"getAiUsageStats": {"SUPER_ADMIN", "ADMIN"},

	// — PDF Import (parse jadwal dari PDF)
	"parsePdfMeeting": {"ADMIN", "TIM_ABSENSI"},

	// — Pending
	"getPendingMembers":      {"ADMIN"},
	"getPendingMemberDetail": {"ADMIN"},
	"approvePendingMember":   {"ADMIN"},
	"rejectPendingMember":    {"ADMIN"},

	// — Profile sendiri (semua role, backend cek member_id)
	"getMyProfile":        {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"updateMyProfile":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMyAttendance":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getMyMonitoring":     {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},
	"getUpcomingMeetings": {"SUPER_ADMIN", "ADMIN", "TIM_PNKB", "TIM_ABSENSI", "PENGAWAS", "MEMBER"},

	// — WA Queue
	"createWaQueue":        {"ADMIN", "TIM_ABSENSI"},
	"cancelWaQueue":        {"ADMIN", "TIM_ABSENSI"},
	"getWaQueueStatus":     {"ADMIN", "TIM_ABSENSI"},
	"bulkGetWaQueueStatus": {"ADMIN", "TIM_ABSENSI"},
	"listWaQueue":          {"ADMIN", "TIM_ABSENSI"},
	"retryWaQueue":         {"ADMIN", "TIM_ABSENSI"},
}

// CanAccess: cek role boleh akses action.
// SUPER_ADMIN selalu boleh.
func CanAccess(role, action string) bool {
	if role == "SUPER_ADMIN" {
		return true
	}
	roles, ok := rolePermissions[action]
	if !ok {
		// Action belum terdaftar di permission map.
		// Return false untuk safety (deny by default).
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// SuperAdminOnlyActions — action yang hanya SUPER_ADMIN.
var SuperAdminOnlyActions = map[string]bool{
	"getUsers":          true,
	"getUserDetail":     true,
	"createUser":        true,
	"updateUser":        true,
	"updateUserRole":    true,
	"getAuditLogs":      true,
	"updateSettings":    true,
	"resetUserPassword": true,
}

func IsSuperAdminOnly(action string) bool {
	return SuperAdminOnlyActions[action]
}
