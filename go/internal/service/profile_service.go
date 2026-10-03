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
		"nickname":           true,
		"whatsapp_number":                    true,
		"home_address":             true,
		"village":                     true,
		"region":                   true,
		"occupation":                true,
		"hobby":                     true,
		"photo_url":                 true,
		"height":             true,
		"weight":              true,
		"is_employed":                 true,
		"is_married":                 true,
		"is_preacher":             true,
		"education_level":       true,
		"school":                  true,
		"major":                  true,
		"education_start_year":   true,
		"education_end_year": true,
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

	if v, ok := filtered["whatsapp_number"].(string); ok && v != "" {
		filtered["whatsapp_number"] = util.NormalizePhone(v)
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
			"attendance_id":  a.AttendanceID,
			"meeting_id":     a.MeetingID,
			"status":         a.Status,
			"status_meeting": mt.Status,
			"notes":        a.Notes,
			"date":        mt.Date.Format("2006-01-02"),
			"day":           mt.Day,
			"event":          mt.Event,
			"time":            mt.Time,
			"created_at":     a.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
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
	memberKat := util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried)
	today := time.Now().Format("2006-01-02")
	meetings, _ := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})

	out := []model.MeetingDTO{}
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
		"full_name":             m.FullName,
		"nickname":           m.Nickname,
		"gender":            strOr(m.Gender, ""),
		"birth_place":             m.BirthPlace,
		"birth_date":            util.FormatDate(m.BirthDate),
		"photo_url":                 m.PhotoURL,
		"whatsapp_number":                    m.WhatsappNumber,
		"home_address":             m.HomeAddress,
		"village":                     m.Village,
		"region":                   m.Region,
		"group_label":                 m.GroupLabel,
		"occupation":                m.Occupation,
		"hobby":                     m.Hobby,
		"height":             m.Height,
		"weight":              m.Weight,
		"is_employed":                 m.IsEmployed,
		"is_married":                 m.IsMarried,
		"is_preacher":             m.IsPreacher,
		"mentoring_status":         m.MentoringStatus,
		"is_active":             m.IsActive,
		"joined_date":            util.FormatDate(m.JoinedDate),
		"education_level":       m.EducationLevel,
		"school":                  m.School,
		"major":                  m.Major,
		"education_start_year":   m.EducationStartYear,
		"education_end_year": m.EducationEndYear,
		"kategori":                 util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried),
		"usia":                     util.GetAge(m.BirthDate),
		"pendidikan":               []any{},
	}
}
