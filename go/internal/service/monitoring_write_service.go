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
	Tanggal      string
	Jenis        string
	Status       string
	Catatan      string
	TindakLanjut string
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
	if in.Tanggal != "" {
		t, err := time.Parse("2006-01-02", in.Tanggal)
		if err != nil {
			return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
		}
		tgl = t
	} else {
		tgl = time.Now()
	}

	jenis := in.Jenis
	if jenis == "" {
		jenis = "UMUM"
	}

	m := &model.Monitoring{
		MonitoringID: util.NewID("MON"),
		MemberID:     in.MemberID,
		Tanggal:      tgl,
		Jenis:        jenis,
		Status:       in.Status,
		Catatan:      in.Catatan,
		TindakLanjut: in.TindakLanjut,
	}
	if in.UserID != "" {
		m.CreatedBy = &in.UserID
	}

	if err := s.repo.Insert(ctx, m); err != nil {
		return nil, err
	}

	// Update status_pembinaan di members
	if err := s.memberRepo.Update(ctx, in.MemberID, map[string]interface{}{
		"status_pembinaan": in.Status,
	}); err != nil {
		// tidak fatal, log saja
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
	Catatan      *string
	TindakLanjut *string
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
	if in.Catatan != nil {
		patch["catatan"] = *in.Catatan
	}
	if in.TindakLanjut != nil {
		patch["tindak_lanjut"] = *in.TindakLanjut
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
