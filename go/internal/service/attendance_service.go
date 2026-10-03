package service

import (
	"context"
	"errors"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type AttendanceService struct {
	repo        *repository.AttendanceRepo
	meetingRepo *repository.MeetingRepo
	memberSvc   *MemberService
}

func NewAttendanceService(repo *repository.AttendanceRepo, meetingRepo *repository.MeetingRepo, memberSvc *MemberService) *AttendanceService {
	return &AttendanceService{
		repo:        repo,
		meetingRepo: meetingRepo,
		memberSvc:   memberSvc,
	}
}

var validAttendanceStatuses = map[string]bool{
	"HADIR": true,
	"IZIN":  true,
	"SAKIT": true,
	"ALPA":  true,
}

func (s *AttendanceService) GetAttendance(ctx context.Context, meetingID, memberID string) ([]model.AttendanceDTO, error) {
	if meetingID != "" {
		rows, err := s.repo.FindByMeeting(ctx, meetingID)
		if err != nil {
			return nil, err
		}
		return toAttendanceDTOs(rows), nil
	}
	if memberID != "" {
		rows, err := s.repo.FindByMember(ctx, memberID)
		if err != nil {
			return nil, err
		}
		return toAttendanceDTOs(rows), nil
	}
	return nil, errors.New("meeting_id atau member_id wajib diisi")
}

func (s *AttendanceService) GetAttendancePage(ctx context.Context, meetingID string) (*model.AttendancePageDTO, error) {
	if meetingID == "" {
		return nil, errors.New("meeting_id wajib diisi")
	}
	meeting, err := s.meetingRepo.FindByID(ctx, meetingID)
	if err != nil {
		return nil, errors.New("meeting tidak ditemukan")
	}
	/* Peserta wajib = jamaah yang kelompoknya sama dengan group_label meeting.
	   Meeting tanpa group (legacy) tetap menampilkan semua jamaah. */
	groupID := ""
	if meeting.GroupID != nil {
		groupID = *meeting.GroupID
	}
	members, err := s.memberSvc.GetAttendanceMembers(ctx, model.MemberListFilter{}, groupID)
	if err != nil {
		return nil, err
	}
	attendance, err := s.repo.FindByMeeting(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	return &model.AttendancePageDTO{
		Meeting:    toMeetingDTO(*meeting),
		Members:    members,
		Attendance: toAttendanceDTOs(attendance),
	}, nil
}

type SaveAttendanceInput struct {
	MeetingID string
	MemberID  string
	Status    string
	Notes   string
	UserID    string
}

func (s *AttendanceService) SaveAttendance(ctx context.Context, in SaveAttendanceInput) (*model.AttendanceDTO, error) {
	if in.MeetingID == "" || in.MemberID == "" || in.Status == "" {
		return nil, errors.New("meeting_id, member_id, dan status wajib diisi")
	}
	if !validAttendanceStatuses[in.Status] {
		return nil, errors.New("status absensi tidak valid")
	}

	meeting, err := s.meetingRepo.FindByID(ctx, in.MeetingID)
	if err != nil {
		return nil, errors.New("meeting tidak ditemukan")
	}
	if util.IsLiburMeeting(*meeting) {
		return nil, errors.New("schedule libur tidak bisa diisi absensi")
	}

	existing, err := s.repo.FindByMeetingAndMember(ctx, in.MeetingID, in.MemberID)
	if err == nil && existing != nil {
		notes := in.Notes
		if notes == "" {
			notes = existing.Notes
		}
		if err := s.repo.Update(ctx, existing.AttendanceID, in.Status, notes); err != nil {
			return nil, err
		}
		fresh, _ := s.repo.FindByMeetingAndMember(ctx, in.MeetingID, in.MemberID)
		if fresh != nil {
			dto := toAttendanceDTO(*fresh)
			return &dto, nil
		}
	}

	a := &model.Attendance{
		AttendanceID: util.NewID("ATD"),
		MeetingID:    in.MeetingID,
		MemberID:     in.MemberID,
		Status:       in.Status,
		Notes:      in.Notes,
	}
	if in.UserID != "" {
		a.CreatedBy = &in.UserID
	}
	if err := s.repo.Insert(ctx, a); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindByMeetingAndMember(ctx, in.MeetingID, in.MemberID)
	if fresh == nil {
		return nil, errors.New("gagal ambil data setelah insert")
	}
	dto := toAttendanceDTO(*fresh)
	return &dto, nil
}

type BulkSaveInput struct {
	MeetingID string
	Items     []repository.BulkItem
	UserID    string
}

type BulkSaveResult struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
}

func (s *AttendanceService) BulkSave(ctx context.Context, in BulkSaveInput) (*BulkSaveResult, error) {
	if in.MeetingID == "" {
		return nil, errors.New("meeting_id wajib diisi")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("items wajib diisi")
	}

	meeting, err := s.meetingRepo.FindByID(ctx, in.MeetingID)
	if err != nil {
		return nil, errors.New("meeting tidak ditemukan")
	}
	if util.IsLiburMeeting(*meeting) {
		return nil, errors.New("schedule libur tidak bisa diisi absensi")
	}

	validItems := make([]repository.BulkItem, 0, len(in.Items))
	for _, it := range in.Items {
		if !validAttendanceStatuses[it.Status] {
			continue
		}
		validItems = append(validItems, it)
	}
	inserted, updated, err := s.repo.BulkUpsert(ctx, in.MeetingID, validItems, in.UserID)
	if err != nil {
		return nil, err
	}
	return &BulkSaveResult{Inserted: inserted, Updated: updated}, nil
}

func (s *AttendanceService) DeleteAttendance(ctx context.Context, meetingID, memberID string) (int64, error) {
	if meetingID == "" || memberID == "" {
		return 0, errors.New("meeting_id dan member_id wajib diisi")
	}
	existing, err := s.repo.FindByMeetingAndMember(ctx, meetingID, memberID)
	if err != nil || existing == nil {
		return 0, nil
	}
	if err := s.repo.Delete(ctx, existing.AttendanceID); err != nil {
		return 0, err
	}
	return 1, nil
}

func (s *AttendanceService) DeleteByMeeting(ctx context.Context, meetingID string) (int64, error) {
	if meetingID == "" {
		return 0, errors.New("meeting_id wajib diisi")
	}
	return s.repo.DeleteByMeeting(ctx, meetingID)
}

func (s *AttendanceService) DeleteByMember(ctx context.Context, memberID string) (int64, error) {
	if memberID == "" {
		return 0, errors.New("member_id wajib diisi")
	}
	return s.repo.DeleteByMember(ctx, memberID)
}

func toAttendanceDTOs(rows []model.Attendance) []model.AttendanceDTO {
	out := make([]model.AttendanceDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAttendanceDTO(a))
	}
	return out
}

func toAttendanceDTO(a model.Attendance) model.AttendanceDTO {
	cby := ""
	if a.CreatedBy != nil {
		cby = *a.CreatedBy
	}
	return model.AttendanceDTO{
		AttendanceID: a.AttendanceID,
		MeetingID:    a.MeetingID,
		MemberID:     a.MemberID,
		Status:       a.Status,
		Notes:      a.Notes,
		CreatedBy:    cby,
		CreatedAt:    a.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    a.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
