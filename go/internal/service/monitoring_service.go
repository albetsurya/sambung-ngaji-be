package service

import (
	"context"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
)

type MonitoringService struct {
	repo *repository.MonitoringRepo
}

func NewMonitoringService(repo *repository.MonitoringRepo) *MonitoringService {
	return &MonitoringService{repo: repo}
}

func (s *MonitoringService) FindByMember(ctx context.Context, memberID string) ([]model.MonitoringDTO, error) {
	rows, err := s.repo.FindByMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	out := make([]model.MonitoringDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMonitoringDTO(r))
	}
	return out, nil
}

func toMonitoringDTO(m model.Monitoring) model.MonitoringDTO {
	cby := ""
	if m.CreatedBy != nil {
		cby = *m.CreatedBy
	}
	return model.MonitoringDTO{
		MonitoringID: m.MonitoringID,
		MemberID:     m.MemberID,
		Tanggal:      m.Tanggal.Format("2006-01-02"),
		Jenis:        m.Jenis,
		Status:       m.Status,
		Catatan:      m.Catatan,
		TindakLanjut: m.TindakLanjut,
		CreatedBy:    cby,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    m.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
