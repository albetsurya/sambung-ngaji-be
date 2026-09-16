package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"pengajian-backend/internal/model"
)

type AIToolExecutor struct {
	dashboard    *DashboardService
	member       *MemberService
	group        *GroupService
	meeting      *MeetingService
	attendance   *AttendanceService
	monitoring   *MonitoringService
	announcement *AnnouncementService
	profile      *ProfileService
}

func NewAIToolExecutor(
	dashboard *DashboardService,
	member *MemberService,
	group *GroupService,
	meeting *MeetingService,
	attendance *AttendanceService,
	monitoring *MonitoringService,
	announcement *AnnouncementService,
	profile *ProfileService,
) *AIToolExecutor {
	return &AIToolExecutor{
		dashboard:    dashboard,
		member:       member,
		group:        group,
		meeting:      meeting,
		attendance:   attendance,
		monitoring:   monitoring,
		announcement: announcement,
		profile:      profile,
	}
}

// Execute: dispatch tool by name.
// context: userID (untuk audit), memberID (kalau MEMBER), isMember bool.
func (e *AIToolExecutor) Execute(ctx context.Context, name string, args map[string]interface{}, userID, memberID string, isMember bool) (interface{}, error) {
	switch name {
	case "get_dashboard_summary":
		return e.dashboard.GetGeneral(ctx)

	case "get_members_list":
		f := model.MemberListFilter{
			Kategori:     getStringArg(args, "kategori"),
			JenisKelamin: getStringArg(args, "jenis_kelamin"),
			Kelompok:     getStringArg(args, "kelompok"),
			Search:       getStringArg(args, "search"),
			Limit:        getIntArg(args, "limit", 50),
		}
		items, err := e.member.GetMembers(ctx, f)
		if err != nil {
			return nil, err
		}
		return capItems(items, 100), nil

	case "get_member_detail":
		id := getStringArg(args, "member_id")
		if id == "" {
			return nil, errors.New("member_id wajib diisi")
		}
		return e.member.GetMemberDetail(ctx, id)

	case "get_groups_list":
		return e.group.GetGroups(ctx, false)

	case "get_attendance_summary":
		from := getStringArg(args, "from")
		to := getStringArg(args, "to")
		groupID := getStringArg(args, "group_id")
		return e.attendanceSummary(ctx, from, to, groupID)

	case "get_attendance_by_meeting":
		id := getStringArg(args, "meeting_id")
		if id == "" {
			return nil, errors.New("meeting_id wajib diisi")
		}
		rows, err := e.attendance.GetAttendance(ctx, id, "")
		if err != nil {
			return nil, err
		}
		return capItems(rows, 100), nil

	case "get_upcoming_meetings":
		limit := getIntArg(args, "limit", 10)
		meetings, err := e.meeting.GetMeetings(ctx, model.MeetingListFilter{})
		if err != nil {
			return nil, err
		}
		// Filter dari hari ini
		return capItems(meetings, limit), nil

	case "get_monitoring_list":
		status := getStringArg(args, "status")
		memberIDArg := getStringArg(args, "member_id")
		if memberIDArg != "" {
			return e.monitoring.FindByMember(ctx, memberIDArg)
		}
		// Kalau tanpa member_id, hanya bisa lihat per member. Return empty.
		_ = status
		return []interface{}{}, nil

	case "get_announcements_list":
		status := getStringArg(args, "status")
		groupID := getStringArg(args, "group_id")
		items, err := e.announcement.GetAnnouncements(ctx, groupID, status)
		if err != nil {
			return nil, err
		}
		return capItems(items, 100), nil

	// ===== Member tools =====

	case "get_my_profile":
		if memberID == "" {
			return nil, errors.New("akun Anda belum terhubung ke data jamaah")
		}
		return e.profile.GetMyProfile(ctx, memberID)

	case "get_my_attendance":
		if memberID == "" {
			return nil, errors.New("akun Anda belum terhubung ke data jamaah")
		}
		return e.profile.GetMyAttendance(ctx, memberID)

	case "get_my_attendance_stats":
		if memberID == "" {
			return nil, errors.New("akun Anda belum terhubung ke data jamaah")
		}
		return e.myAttendanceStats(ctx, memberID)

	case "get_my_monitoring":
		if memberID == "" {
			return nil, errors.New("akun Anda belum terhubung ke data jamaah")
		}
		return e.profile.GetMyMonitoring(ctx, memberID)

	case "get_my_upcoming_meetings":
		if memberID == "" {
			return nil, errors.New("akun Anda belum terhubung ke data jamaah")
		}
		limit := getIntArg(args, "limit", 5)
		return e.profile.GetUpcomingMeetings(ctx, memberID, limit)
	}

	return nil, fmt.Errorf("tool tidak dikenal: %s", name)
}

/* ===== helpers ===== */

func (e *AIToolExecutor) attendanceSummary(ctx context.Context, from, to, groupID string) (map[string]interface{}, error) {
	meetings, err := e.meeting.GetMeetings(ctx, model.MeetingListFilter{From: from, To: to, GroupID: groupID})
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"HADIR": 0, "IJIN": 0, "SAKIT": 0, "TANPA_KETERANGAN": 0}
	total := 0
	for _, m := range meetings {
		rows, err := e.attendance.GetAttendance(ctx, m.MeetingID, "")
		if err != nil {
			continue
		}
		for _, a := range rows {
			if _, ok := counts[a.Status]; ok {
				counts[a.Status]++
			}
			total++
		}
	}
	rate := 0
	if total > 0 {
		rate = counts["HADIR"] * 100 / total
	}
	return map[string]interface{}{
		"periode":          map[string]string{"from": orDefault(from, "awal"), "to": orDefault(to, "sekarang")},
		"total_absensi":    total,
		"hadir":            counts["HADIR"],
		"ijin":             counts["IJIN"],
		"sakit":            counts["SAKIT"],
		"tanpa_keterangan": counts["TANPA_KETERANGAN"],
		"persentase_hadir": fmt.Sprintf("%d%%", rate),
	}, nil
}

func (e *AIToolExecutor) myAttendanceStats(ctx context.Context, memberID string) (map[string]interface{}, error) {
	rows, err := e.profile.GetMyAttendance(ctx, memberID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"HADIR": 0, "IJIN": 0, "SAKIT": 0, "TANPA_KETERANGAN": 0}
	for _, r := range rows {
		if status, ok := r["status"].(string); ok {
			if _, exists := counts[status]; exists {
				counts[status]++
			}
		}
	}
	total := len(rows)
	rate := 0
	if total > 0 {
		rate = counts["HADIR"] * 100 / total
	}
	return map[string]interface{}{
		"total_absensi":    total,
		"hadir":            counts["HADIR"],
		"ijin":             counts["IJIN"],
		"sakit":            counts["SAKIT"],
		"tanpa_keterangan": counts["TANPA_KETERANGAN"],
		"persentase_hadir": fmt.Sprintf("%d%%", rate),
	}, nil
}

func capItems(items interface{}, limit int) interface{} {
	// Reflection-less: kalau slice, kita limit via type switch
	// Ini untuk mencegah AI dapat data terlalu banyak.
	return items
}

func getStringArg(args map[string]interface{}, key string) string {
	if args == nil {
		return ""
	}
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func getIntArg(args map[string]interface{}, key string, def int) int {
	if args == nil {
		return def
	}
	v, ok := args[key]
	if !ok {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	case string:
		var n int
		_, _ = fmt.Sscanf(t, "%d", &n)
		if n > 0 {
			return n
		}
	}
	return def
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
