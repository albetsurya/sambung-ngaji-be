package service

import (
	"context"
	"sort"
	"strconv"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type DashboardService struct {
	memberRepo     *repository.MemberRepo
	meetingRepo    *repository.MeetingRepo
	attendanceRepo *repository.AttendanceRepo
	monitoringRepo *repository.MonitoringRepo
}

func NewDashboardService(
	memberRepo *repository.MemberRepo,
	meetingRepo *repository.MeetingRepo,
	attendanceRepo *repository.AttendanceRepo,
	monitoringRepo *repository.MonitoringRepo,
) *DashboardService {
	return &DashboardService{
		memberRepo:     memberRepo,
		meetingRepo:    meetingRepo,
		attendanceRepo: attendanceRepo,
		monitoringRepo: monitoringRepo,
	}
}

type AttentionItem struct {
	MemberID string   `json:"member_id"`
	FullName string   `json:"full_name"`
	PhotoURL string   `json:"photo_url"`
	Reasons  []string `json:"reasons"`
}

type GeneralDashboard struct {
	TotalJamaah          int                    `json:"total_jamaah"`
	TotalJamaahAktif     int                    `json:"total_jamaah_aktif"`
	PerKategori          map[string]int         `json:"per_kategori"`
	RataRataKehadiran    int                    `json:"rata_rata_kehadiran"`
	PengajianTerdekat    map[string]interface{} `json:"pengajian_terdekat"`
	JamaahPerluPerhatian []AttentionItem        `json:"jamaah_perlu_perhatian"`
	DataBelumLengkap     int                    `json:"data_belum_lengkap"`
}

func (s *DashboardService) GetGeneral(ctx context.Context, groupID string) (*GeneralDashboard, error) {
	members, err := s.memberRepo.FindAllByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	meetings, err := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{GroupID: groupID})
	if err != nil {
		return nil, err
	}

	meetingIDs := make([]string, 0, len(meetings))
	for _, m := range meetings {
		meetingIDs = append(meetingIDs, m.MeetingID)
	}
	attByMeeting, err := s.attendanceRepo.FindByMeetingIDs(ctx, meetingIDs)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]string, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.MemberID)
	}
	attByMember, err := s.attendanceRepo.FindByMemberIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}

	attGrouped := make(map[string][]model.Attendance, len(members))
	for _, a := range attByMember {
		attGrouped[a.MemberID] = append(attGrouped[a.MemberID], a)
	}

	perKategori := map[string]int{
		util.KatBalita: 0, util.KatCaberawit: 0, util.KatPraRemaja: 0,
		util.KatRemaja: 0, util.KatPraNikah: 0, util.KatDewasa: 0, util.KatIstimewa: 0,
	}
	for _, m := range members {
		k := util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried)
		perKategori[k]++
	}

	today := time.Now().Format("2006-01-02")
	var upcoming *model.Meeting
	var futureMeetings []model.Meeting
	for _, m := range meetings {
		if m.Date.Format("2006-01-02") >= today {
			futureMeetings = append(futureMeetings, m)
		}
	}
	if len(futureMeetings) > 0 {
		sort.Slice(futureMeetings, func(i, j int) bool {
			return futureMeetings[i].Date.Before(futureMeetings[j].Date)
		})
		upcoming = &futureMeetings[0]
	}

	liburMeetingIDs := util.LiburMeetingIDs(meetings)

	allCount, hadirCount := 0, 0
	for _, a := range attByMeeting {
		if liburMeetingIDs[a.MeetingID] {
			continue
		}
		allCount++
		if a.Status == "HADIR" {
			hadirCount++
		}
	}
	rate := 0
	if allCount > 0 {
		rate = (hadirCount * 100) / allCount
	}

	meetingDates := make(map[string]time.Time, len(meetings))
	for _, m := range meetings {
		meetingDates[m.MeetingID] = m.Date
	}

	attention := s.buildAttentionList(members, attGrouped, meetingDates, liburMeetingIDs)

	incomplete := 0
	for _, m := range members {
		if m.BirthDate == nil || m.GroupLabel == "" || m.Village == "" || m.HomeAddress == "" {
			incomplete++
		}
	}

	var terdekat map[string]interface{}
	if upcoming != nil {
		terdekat = map[string]interface{}{
			"meeting_id":        upcoming.MeetingID,
			"date":              upcoming.Date.Format("2006-01-02"),
			"day":               upcoming.Day,
			"event":             upcoming.Event,
			"target_categories": upcoming.TargetCategories,
		}
	}

	return &GeneralDashboard{
		TotalJamaah:          len(members),
		TotalJamaahAktif:     len(members),
		PerKategori:          perKategori,
		RataRataKehadiran:    rate,
		PengajianTerdekat:    terdekat,
		JamaahPerluPerhatian: attention,
		DataBelumLengkap:     incomplete,
	}, nil
}

func (s *DashboardService) buildAttentionList(
	members []model.Member,
	attByMember map[string][]model.Attendance,
	meetingDates map[string]time.Time,
	liburMeetingIDs map[string]bool,
) []AttentionItem {
	out := []AttentionItem{}
	cutoff := time.Now().AddDate(0, 0, -30)

	for _, m := range members {
		rows := attByMember[m.MemberID]

		window := make([]model.Attendance, 0, len(rows))
		for _, r := range rows {
			if liburMeetingIDs[r.MeetingID] {
				continue
			}
			if d, ok := meetingDates[r.MeetingID]; ok && d.After(cutoff) {
				window = append(window, r)
			}
		}
		sort.Slice(window, func(i, j int) bool {
			return meetingDates[window[i].MeetingID].After(meetingDates[window[j].MeetingID])
		})

		reasons := []string{}

		if len(window) >= 3 {
			hadir := 0
			for _, r := range window {
				if r.Status == "HADIR" {
					hadir++
				}
			}
			rate := (hadir * 100) / len(window)
			if rate < 50 {
				reasons = append(reasons, "Kehadiran 30 day terakhir <50% ("+strconv.Itoa(rate)+"%)")
			}
		}

		if run := maxConsecutiveRun(window, "SAKIT"); run >= 3 {
			reasons = append(reasons, "Sakit "+strconv.Itoa(run)+" pertemuan berturut-turut")
		}

		if run := maxConsecutiveRun(window, "ALPA"); run >= 5 {
			reasons = append(reasons, "Tanpa keterangan "+strconv.Itoa(run)+" pertemuan berturut-turut")
		}

		if m.UpdatedAt.Before(cutoff) {
			reasons = append(reasons, "Data belum diperbarui > 30 day")
		}
		if m.MentoringStatus == "PERLU_PERHATIAN" || m.MentoringStatus == "TIDAK_AKTIF" {
			reasons = append(reasons, "Status pembinaan: "+m.MentoringStatus)
		}

		if len(reasons) > 0 {
			out = append(out, AttentionItem{
				MemberID: m.MemberID,
				FullName: m.FullName,
				PhotoURL: m.PhotoURL,
				Reasons:  reasons,
			})
		}
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func maxConsecutiveRun(window []model.Attendance, status string) int {
	maxRun, cur := 0, 0
	for _, r := range window {
		if r.Status == status {
			cur++
			if cur > maxRun {
				maxRun = cur
			}
		} else {
			cur = 0
		}
	}
	return maxRun
}

type MyDashboard struct {
	Profile    map[string]interface{}   `json:"profile"`
	Attendance []map[string]interface{} `json:"attendance"`
	Monitoring []model.MonitoringDTO    `json:"monitoring"`
	Upcoming   []model.MeetingDTO       `json:"upcoming"`
}

func (s *DashboardService) GetMyDashboard(ctx context.Context, memberID string) (*MyDashboard, error) {
	m, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	profile := map[string]interface{}{
		"member_id":            m.MemberID,
		"full_name":            m.FullName,
		"nickname":             m.Nickname,
		"gender":               strOr(m.Gender, ""),
		"birth_place":          m.BirthPlace,
		"birth_date":           util.FormatDate(m.BirthDate),
		"photo_url":            m.PhotoURL,
		"group_label":          m.GroupLabel,
		"village":              m.Village,
		"region":               m.Region,
		"home_address":         m.HomeAddress,
		"whatsapp_number":      m.WhatsappNumber,
		"occupation":           m.Occupation,
		"hobby":                m.Hobby,
		"height":               m.Height,
		"weight":               m.Weight,
		"is_employed":          m.IsEmployed,
		"is_married":           m.IsMarried,
		"is_preacher":          m.IsPreacher,
		"joined_date":          util.FormatDate(m.JoinedDate),
		"mentoring_status":     m.MentoringStatus,
		"is_active":            m.IsActive,
		"education_level":      m.EducationLevel,
		"school":               m.School,
		"major":                m.Major,
		"education_start_year": m.EducationStartYear,
		"education_end_year":   m.EducationEndYear,
		"kategori":             util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried),
		"usia":                 util.GetAge(m.BirthDate),
		"pendidikan":           []any{},
	}

	attRows, _ := s.attendanceRepo.FindByMember(ctx, memberID)
	meetings, _ := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})
	meetingsByID := map[string]model.Meeting{}
	for _, mt := range meetings {
		meetingsByID[mt.MeetingID] = mt
	}

	attendance := []map[string]interface{}{}
	for i, a := range attRows {
		if i >= 50 {
			break
		}
		mt := meetingsByID[a.MeetingID]
		attendance = append(attendance, map[string]interface{}{
			"attendance_id":  a.AttendanceID,
			"meeting_id":     a.MeetingID,
			"status":         a.Status,
			"status_meeting": mt.Status,
			"notes":          a.Notes,
			"date":           mt.Date.Format("2006-01-02"),
			"day":            mt.Day,
			"event":          mt.Event,
			"time":           mt.Time,
			"created_at":     a.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	}

	monRows, _ := s.monitoringRepo.FindByMember(ctx, memberID)
	monitoring := []model.MonitoringDTO{}
	for i, mr := range monRows {
		if i >= 20 {
			break
		}
		monitoring = append(monitoring, toMonitoringDTO(mr))
	}

	memberKat := util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried)
	today := time.Now().Format("2006-01-02")
	upcoming := []model.MeetingDTO{}
	for _, mt := range meetings {
		if mt.Date.Format("2006-01-02") < today {
			continue
		}
		if len(mt.TargetCategories) > 0 {
			match := false
			for _, k := range mt.TargetCategories {
				if k == memberKat {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		upcoming = append(upcoming, toMeetingDTO(mt))
		if len(upcoming) >= 5 {
			break
		}
	}

	return &MyDashboard{
		Profile:    profile,
		Attendance: attendance,
		Monitoring: monitoring,
		Upcoming:   upcoming,
	}, nil
}
