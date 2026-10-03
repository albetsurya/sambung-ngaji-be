package service

import (
	"context"
	"errors"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

type CreateMonitoringInput struct {
	MemberID     string
	Date      string
	Type        string
	Status       string
	Notes      string
	FollowUp string
	UserID       string
}

var validMonitoringStatus = map[string]bool{
	"AKTIF":           true,
	"PERLU_PERHATIAN": true,
	"KURANG_AKTIF":    true,
	"TIDAK_AKTIF":     true,
}

func (s *MonitoringService) Create(ctx context.Context, in CreateMonitoringInput) (*model.MonitoringDTO, error) {
	if in.MemberID == "" || in.Status == "" {
		return nil, errors.New("member_id dan status wajib diisi")
	}
	if !validMonitoringStatus[in.Status] {
		return nil, errors.New("status monitoring tidak valid")
	}
	if _, err := s.memberRepo.FindByID(ctx, in.MemberID); err != nil {
		return nil, errors.New("jamaah tidak ditemukan")
	}

	var tgl time.Time
	if in.Date != "" {
		t, err := time.Parse("2006-01-02", in.Date)
		if err != nil {
			return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
		}
		tgl = t
	} else {
		tgl = time.Now()
	}

	mtype := in.Type
	if mtype == "" {
		mtype = "UMUM"
	}

	m := &model.Monitoring{
		MonitoringID: util.NewID("MON"),
		MemberID:     in.MemberID,
		Date:      tgl,
		Type:        mtype,
		Status:       in.Status,
		Notes:      in.Notes,
		FollowUp: in.FollowUp,
	}
	if in.UserID != "" {
		m.CreatedBy = &in.UserID
	}

	if err := s.repo.Insert(ctx, m); err != nil {
		return nil, err
	}

	if err := s.memberRepo.Update(ctx, in.MemberID, map[string]interface{}{
		"mentoring_status": in.Status,
	}); err != nil {
		_ = err
	}

	fresh, _ := s.repo.FindByID(ctx, m.MonitoringID)
	if fresh == nil {
		return nil, errors.New("gagal ambil data monitoring")
	}
	dto := toMonitoringDTO(*fresh)
	return &dto, nil
}

type UpdateMonitoringInput struct {
	MonitoringID string
	Notes      *string
	FollowUp *string
}

func (s *MonitoringService) UpdateEntry(ctx context.Context, in UpdateMonitoringInput) (*model.MonitoringDTO, error) {
	if in.MonitoringID == "" {
		return nil, errors.New("monitoring_id wajib diisi")
	}
	existing, err := s.repo.FindByID(ctx, in.MonitoringID)
	if err != nil || existing == nil {
		return nil, errors.New("data monitoring tidak ditemukan")
	}

	patch := map[string]interface{}{}
	if in.Notes != nil {
		patch["notes"] = *in.Notes
	}
	if in.FollowUp != nil {
		patch["follow_up"] = *in.FollowUp
	}
	if len(patch) == 0 {
		return nil, errors.New("tidak ada perubahan")
	}

	if err := s.repo.Update(ctx, in.MonitoringID, patch); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindByID(ctx, in.MonitoringID)
	dto := toMonitoringDTO(*fresh)
	return &dto, nil
}
