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
	MemberID    string   `json:"member_id"`
	NamaLengkap string   `json:"nama_lengkap"`
	FotoURL     string   `json:"foto_url"`
	Reasons     []string `json:"reasons"`
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

// GetGeneral: total 4 query (members, meetings, attendance-by-meeting, attendance-by-member).
// Sebelumnya N+1 — sekarang O(1) query.
func (s *DashboardService) GetGeneral(ctx context.Context) (*GeneralDashboard, error) {
	members, err := s.memberRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	meetings, err := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})
	if err != nil {
		return nil, err
	}

	// ===== Batch fetch attendance =====

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

	// Group attendance per member (urutan tetap DESC by created_at).
	attGrouped := make(map[string][]model.Attendance, len(members))
	for _, a := range attByMember {
		attGrouped[a.MemberID] = append(attGrouped[a.MemberID], a)
	}

	// ===== Kategori =====

	perKategori := map[string]int{
		util.KatBalita: 0, util.KatCaberawit: 0, util.KatPraRemaja: 0,
		util.KatRemaja: 0, util.KatPraNikah: 0, util.KatDewasa: 0, util.KatIstimewa: 0,
	}
	for _, m := range members {
		k := util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah)
		perKategori[k]++
	}

	// ===== Pengajian terdekat =====

	today := time.Now().Format("2006-01-02")
	var upcoming *model.Meeting
	var futureMeetings []model.Meeting
	for _, m := range meetings {
		if m.Tanggal.Format("2006-01-02") >= today {
			futureMeetings = append(futureMeetings, m)
		}
	}
	if len(futureMeetings) > 0 {
		sort.Slice(futureMeetings, func(i, j int) bool {
			return futureMeetings[i].Tanggal.Before(futureMeetings[j].Tanggal)
		})
		upcoming = &futureMeetings[0]
	}

	// ===== Rata-rata kehadiran (pakai attByMeeting flat) =====

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

	// ===== Attention list (pakai attGrouped) =====

	meetingDates := make(map[string]time.Time, len(meetings))
	for _, m := range meetings {
		meetingDates[m.MeetingID] = m.Tanggal
	}

	attention := s.buildAttentionList(members, attGrouped, meetingDates, liburMeetingIDs)

	// ===== Data belum lengkap =====

	incomplete := 0
	for _, m := range members {
		if m.TanggalLahir == nil || m.Kelompok == "" || m.Desa == "" || m.AlamatRumah == "" {
			incomplete++
		}
	}

	var terdekat map[string]interface{}
	if upcoming != nil {
		terdekat = map[string]interface{}{
			"meeting_id":      upcoming.MeetingID,
			"tanggal":         upcoming.Tanggal.Format("2006-01-02"),
			"hari":            upcoming.Hari,
			"acara":           upcoming.Acara,
			"kategori_target": upcoming.KategoriTarget,
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

// buildAttentionList: terima attendance yang sudah di-preload per member + map tanggal meeting.
// Aturan flag (window rolling 30 hari terakhir dari hari ini):
//   - Kehadiran < 50% dalam 30 hari (minimal 3 absensi di window)
//   - SAKIT >= 3 pertemuan berturut-turut (run maksimal di window)
//   - ALPA >= 5 pertemuan berturut-turut (run maksimal di window)
//
// Plus status pembinaan & data lama (dari logika sebelumnya).
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

		// Ambil absensi dalam window 30 hari + urutkan tanggal meeting terbaru dulu
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

		// 1. Kehadiran < 50% dalam 30 hari
		if len(window) >= 3 {
			hadir := 0
			for _, r := range window {
				if r.Status == "HADIR" {
					hadir++
				}
			}
			rate := (hadir * 100) / len(window)
			if rate < 50 {
				reasons = append(reasons, "Kehadiran 30 hari terakhir <50% ("+strconv.Itoa(rate)+"%)")
			}
		}

		// 2. SAKIT >= 3 berturut dalam window
		if run := maxConsecutiveRun(window, "SAKIT"); run >= 3 {
			reasons = append(reasons, "Sakit "+strconv.Itoa(run)+" pertemuan berturut-turut")
		}

		// 3. ALPA >= 5 berturut dalam window
		if run := maxConsecutiveRun(window, "ALPA"); run >= 5 {
			reasons = append(reasons, "Tanpa keterangan "+strconv.Itoa(run)+" pertemuan berturut-turut")
		}

		if m.UpdatedAt.Before(cutoff) {
			reasons = append(reasons, "Data belum diperbarui > 30 hari")
		}
		if m.StatusPembinaan == "PERLU_PERHATIAN" || m.StatusPembinaan == "TIDAK_AKTIF" {
			reasons = append(reasons, "Status pembinaan: "+m.StatusPembinaan)
		}

		if len(reasons) > 0 {
			out = append(out, AttentionItem{
				MemberID:    m.MemberID,
				NamaLengkap: m.NamaLengkap,
				FotoURL:     m.FotoURL,
				Reasons:     reasons,
			})
		}
		if len(out) >= 10 {
			break
		}
	}
	return out
}

// maxConsecutiveRun: panjang run terpanjang status tertentu pada window yang
// sudah diurutkan tanggal terbaru→terlama. "Berturut-turut" = record absensi
// yang urut (pertemuan berurutan), bukan adjacency hari kalender (pengajian
// tidak harian).
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

/* ===== My Dashboard (MEMBER) ===== */

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
		"member_id":          m.MemberID,
		"nama_lengkap":       m.NamaLengkap,
		"nama_panggilan":     m.NamaPanggilan,
		"jenis_kelamin":      strOr(m.JenisKelamin, ""),
		"tempat_lahir":       m.TempatLahir,
		"tanggal_lahir":      util.FormatDate(m.TanggalLahir),
		"foto_url":           m.FotoURL,
		"kelompok":           m.Kelompok,
		"desa":               m.Desa,
		"daerah":             m.Daerah,
		"alamat_rumah":       m.AlamatRumah,
		"no_wa":              m.NoWA,
		"pekerjaan":          m.Pekerjaan,
		"hobi":               m.Hobi,
		"status_pembinaan":   m.StatusPembinaan,
		"status_aktif":       m.StatusAktif,
		"jenjang_pendidikan": m.JenjangPendidikan,
		"sekolah":            m.Sekolah,
		"jurusan":            m.Jurusan,
		"kategori":           util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah),
		"usia":               util.GetAge(m.TanggalLahir),
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
			"catatan":        a.Catatan,
			"tanggal":        mt.Tanggal.Format("2006-01-02"),
			"hari":           mt.Hari,
			"acara":          mt.Acara,
			"jam":            mt.Jam,
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

	memberKat := util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah)
	today := time.Now().Format("2006-01-02")
	upcoming := []model.MeetingDTO{}
	for _, mt := range meetings {
		if mt.Tanggal.Format("2006-01-02") < today {
			continue
		}
		if len(mt.KategoriTarget) > 0 {
			match := false
			for _, k := range mt.KategoriTarget {
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
