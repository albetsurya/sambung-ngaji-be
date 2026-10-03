package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type MeetingService struct {
	repo *repository.MeetingRepo
}

func NewMeetingService(repo *repository.MeetingRepo) *MeetingService {
	return &MeetingService{repo: repo}
}

/* Ambil satu meeting (untuk guard akses per-record). */
func (s *MeetingService) GetMeeting(ctx context.Context, id string) (*model.Meeting, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *MeetingService) GetMeetings(ctx context.Context, f model.MeetingListFilter) ([]model.MeetingDTO, error) {
	meetings, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]model.MeetingDTO, 0, len(meetings))
	for _, m := range meetings {
		out = append(out, toMeetingDTO(m))
	}
	return out, nil
}

type CreateMeetingInput struct {
	Date        string
	Time            string
	StartTime       string
	GroupID        string
	Event          string
	Topic         string
	Status         string
	Notes        string
	TargetCategories []string
	GenderTarget   string
	CreatedBy      string
}

func (s *MeetingService) CreateMeeting(ctx context.Context, in CreateMeetingInput) (*model.MeetingDTO, error) {
	if in.Date == "" {
		return nil, errors.New("date wajib diisi")
	}
	if in.Event == "" {
		return nil, errors.New("event wajib diisi")
	}
	tgl, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}

	m := &model.Meeting{
		MeetingID:      util.NewID("MTG"),
		Date:        tgl,
		Day:           util.GetHariFromDate(&tgl),
		Time:            in.Time,
		StartTime:       in.StartTime,
		Event:          in.Event,
		Topic:         in.Topic,
		Status:         strDef(in.Status, "SCHEDULED"),
		Notes:        in.Notes,
		TargetCategories: in.TargetCategories,
	}
	if in.GenderTarget == "L" || in.GenderTarget == "P" {
		gt := in.GenderTarget
		m.GenderTarget = &gt
	}
	if in.GroupID != "" {
		m.GroupID = &in.GroupID
	}
	if in.CreatedBy != "" {
		m.CreatedBy = &in.CreatedBy
	}
	if m.TargetCategories == nil {
		m.TargetCategories = []string{}
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	fresh, err := s.repo.FindByID(ctx, m.MeetingID)
	if err != nil {
		return nil, err
	}
	dto := toMeetingDTO(*fresh)
	return &dto, nil
}

type UpdateMeetingInput struct {
	MeetingID      string
	Date        string
	Time            string
	StartTime       string
	GroupID        string
	Event          string
	Topic         string
	Status         string
	Notes        string
	TargetCategories *[]string
	GenderTarget   *string
}

func (s *MeetingService) UpdateMeeting(ctx context.Context, in UpdateMeetingInput) (*model.MeetingDTO, error) {
	if in.MeetingID == "" {
		return nil, errors.New("meeting_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, in.MeetingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("jadwal tidak ditemukan")
		}
		return nil, err
	}

	var patch repository.MeetingPatch
	if in.Date != "" {
		tgl, err := time.Parse("2006-01-02", in.Date)
		if err != nil {
			return nil, errors.New("tanggal tidak valid")
		}
		tglStr := tgl.Format("2006-01-02")
		hariStr := util.GetHariFromDate(&tgl)
		patch.Date = &tglStr
		patch.Day = &hariStr
	}
	if in.Time != "" {
		patch.Time = &in.Time
	}
	if in.StartTime != "" {
		patch.StartTime = &in.StartTime
	}
	if in.GroupID != "" {
		patch.GroupID = &in.GroupID
	}
	if in.Event != "" {
		patch.Event = &in.Event
	}
	if in.Topic != "" {
		patch.Topic = &in.Topic
	}
	if in.Status != "" {
		patch.Status = &in.Status
	}
	if in.Notes != "" {
		patch.Notes = &in.Notes
	}
	patch.TargetCategories = in.TargetCategories
	if in.GenderTarget != nil {
		val := *in.GenderTarget
		if val == "L" || val == "P" || val == "" {
			patch.GenderTarget = &val
		}
	}
	if err := s.repo.Update(ctx, in.MeetingID, patch); err != nil {
		return nil, err
	}

	fresh, err := s.repo.FindByID(ctx, in.MeetingID)
	if err != nil {
		return nil, err
	}
	dto := toMeetingDTO(*fresh)
	return &dto, nil
}

type DeleteMeetingsBulkInput struct {
	MeetingIDs []string
}

type DeleteMeetingsBulkResult struct {
	Requested int   `json:"requested"`
	Deleted   int64 `json:"deleted"`
}

func (s *MeetingService) DeleteMeetingBulk(ctx context.Context, in DeleteMeetingsBulkInput) (*DeleteMeetingsBulkResult, error) {
	seen := map[string]bool{}
	ids := make([]string, 0, len(in.MeetingIDs))
	for _, id := range in.MeetingIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("meeting_ids wajib diisi")
	}
	if len(ids) > 100 {
		return nil, errors.New("maksimal 100 jadwal per request")
	}

	deleted, err := s.repo.DeleteMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &DeleteMeetingsBulkResult{
		Requested: len(ids),
		Deleted:   deleted,
	}, nil
}

type DeleteMeetingResult struct {
	MeetingID         string `json:"meeting_id"`
	DeletedAttendance int64  `json:"deleted_attendance"`
}

func (s *MeetingService) DeleteMeeting(ctx context.Context, id string) (*DeleteMeetingResult, error) {
	if id == "" {
		return nil, errors.New("meeting_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("jadwal tidak ditemukan")
		}
		return nil, err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &DeleteMeetingResult{MeetingID: id}, nil
}

func toMeetingDTO(m model.Meeting) model.MeetingDTO {
	gid := ""
	if m.GroupID != nil {
		gid = *m.GroupID
	}
	cby := ""
	if m.CreatedBy != nil {
		cby = *m.CreatedBy
	}
	gt := ""
	if m.GenderTarget != nil {
		gt = *m.GenderTarget
	}
	kat := m.TargetCategories
	if kat == nil {
		kat = []string{}
	}
	return model.MeetingDTO{
		MeetingID:      m.MeetingID,
		Date:        m.Date.Format("2006-01-02"),
		Day:           m.Day,
		Time:            m.Time,
		StartTime:       m.StartTime,
		GroupID:        gid,
		Event:          m.Event,
		Topic:         m.Topic,
		Status:         m.Status,
		Notes:        m.Notes,
		TargetCategories: kat,
		GenderTarget:   gt,
		CreatedBy:      cby,
		CreatedAt:      m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:      m.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

func strDef(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
