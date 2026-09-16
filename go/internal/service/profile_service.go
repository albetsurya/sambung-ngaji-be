package service

import (
	"context"
	"errors"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type ProfileService struct {
	memberRepo     *repository.MemberRepo
	attendanceRepo *repository.AttendanceRepo
	monitoringRepo *repository.MonitoringRepo
	meetingRepo    *repository.MeetingRepo
}

func NewProfileService(
	memberRepo *repository.MemberRepo,
	attendanceRepo *repository.AttendanceRepo,
	monitoringRepo *repository.MonitoringRepo,
	meetingRepo *repository.MeetingRepo,
) *ProfileService {
	return &ProfileService{
		memberRepo:     memberRepo,
		attendanceRepo: attendanceRepo,
		monitoringRepo: monitoringRepo,
		meetingRepo:    meetingRepo,
	}
}

func (s *ProfileService) GetMyProfile(ctx context.Context, memberID string) (map[string]interface{}, error) {
	if memberID == "" {
		return nil, errors.New("Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	m, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return nil, errors.New("data jamaah tidak ditemukan")
	}
	return profileToMap(m), nil
}

func (s *ProfileService) UpdateMyProfile(ctx context.Context, memberID string, patch map[string]interface{}) (map[string]interface{}, error) {
	if memberID == "" {
		return nil, errors.New("Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	if _, err := s.memberRepo.FindByID(ctx, memberID); err != nil {
		return nil, errors.New("data jamaah tidak ditemukan")
	}

	allowed := map[string]bool{
		"nama_panggilan":           true,
		"no_wa":                    true,
		"alamat_rumah":             true,
		"desa":                     true,
		"daerah":                   true,
		"pekerjaan":                true,
		"hobi":                     true,
		"foto_url":                 true,
		"tinggi_badan":             true,
		"berat_badan":              true,
		"is_kerja":                 true,
		"is_nikah":                 true,
		"jenjang_pendidikan":       true,
		"sekolah":                  true,
		"jurusan":                  true,
		"tahun_mulai_pendidikan":   true,
		"tahun_selesai_pendidikan": true,
	}

	filtered := map[string]interface{}{}
	for k, v := range patch {
		if allowed[k] {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("tidak ada perubahan")
	}

	if v, ok := filtered["no_wa"].(string); ok && v != "" {
		filtered["no_wa"] = util.NormalizePhone(v)
	}

	if err := s.memberRepo.Update(ctx, memberID, filtered); err != nil {
		return nil, err
	}
	fresh, _ := s.memberRepo.FindByID(ctx, memberID)
	return profileToMap(fresh), nil
}

func (s *ProfileService) GetMyAttendance(ctx context.Context, memberID string) ([]map[string]interface{}, error) {
	if memberID == "" {
		return nil, errors.New("Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	attRows, _ := s.attendanceRepo.FindByMember(ctx, memberID)
	meetings, _ := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})
	meetingsByID := map[string]model.Meeting{}
	for _, m := range meetings {
		meetingsByID[m.MeetingID] = m
	}
	out := make([]map[string]interface{}, 0, len(attRows))
	for _, a := range attRows {
		mt := meetingsByID[a.MeetingID]
		out = append(out, map[string]interface{}{
			"attendance_id": a.AttendanceID,
			"meeting_id":    a.MeetingID,
			"status":        a.Status,
			"catatan":       a.Catatan,
			"tanggal":       mt.Tanggal.Format("2006-01-02"),
			"hari":          mt.Hari,
			"acara":         mt.Acara,
			"jam":           mt.Jam,
			"created_at":    a.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	}
	return out, nil
}

func (s *ProfileService) GetMyMonitoring(ctx context.Context, memberID string) ([]model.MonitoringDTO, error) {
	if memberID == "" {
		return nil, errors.New("Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	rows, _ := s.monitoringRepo.FindByMember(ctx, memberID)
	out := make([]model.MonitoringDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMonitoringDTO(r))
	}
	return out, nil
}

func (s *ProfileService) GetUpcomingMeetings(ctx context.Context, memberID string, limit int) ([]model.MeetingDTO, error) {
	if memberID == "" {
		return nil, errors.New("Akun Anda belum terhubung ke data jamaah. Hubungi admin.")
	}
	if limit <= 0 {
		limit = 10
	}
	m, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return nil, errors.New("data jamaah tidak ditemukan")
	}
	memberKat := util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah)
	today := time.Now().Format("2006-01-02")
	meetings, _ := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})

	out := []model.MeetingDTO{}
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
		out = append(out, toMeetingDTO(mt))
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func profileToMap(m *model.Member) map[string]interface{} {
	return map[string]interface{}{
		"member_id":                m.MemberID,
		"nama_lengkap":             m.NamaLengkap,
		"nama_panggilan":           m.NamaPanggilan,
		"jenis_kelamin":            strOr(m.JenisKelamin, ""),
		"tempat_lahir":             m.TempatLahir,
		"tanggal_lahir":            util.FormatDate(m.TanggalLahir),
		"foto_url":                 m.FotoURL,
		"no_wa":                    m.NoWA,
		"alamat_rumah":             m.AlamatRumah,
		"desa":                     m.Desa,
		"daerah":                   m.Daerah,
		"kelompok":                 m.Kelompok,
		"pekerjaan":                m.Pekerjaan,
		"hobi":                     m.Hobi,
		"tinggi_badan":             m.TinggiBadan,
		"berat_badan":              m.BeratBadan,
		"is_kerja":                 m.IsKerja,
		"is_nikah":                 m.IsNikah,
		"is_muballigh":             m.IsMuballigh,
		"status_pembinaan":         m.StatusPembinaan,
		"status_aktif":             m.StatusAktif,
		"tanggal_masuk":            util.FormatDate(m.TanggalMasuk),
		"jenjang_pendidikan":       m.JenjangPendidikan,
		"sekolah":                  m.Sekolah,
		"jurusan":                  m.Jurusan,
		"tahun_mulai_pendidikan":   m.TahunMulaiPendidikan,
		"tahun_selesai_pendidikan": m.TahunSelesaiPendidikan,
		"kategori":                 util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah),
		"usia":                     util.GetAge(m.TanggalLahir),
		"pendidikan":               []any{},
	}
}
