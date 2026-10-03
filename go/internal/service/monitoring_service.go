package service

import (
	"context"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
)

type MonitoringService struct {
	repo       *repository.MonitoringRepo
	memberRepo *repository.MemberRepo
}

func NewMonitoringService(repo *repository.MonitoringRepo, memberRepo *repository.MemberRepo) *MonitoringService {
	return &MonitoringService{repo: repo, memberRepo: memberRepo}
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
		Date:      m.Date.Format("2006-01-02"),
		Type:        m.Type,
		Status:       m.Status,
		Notes:      m.Notes,
		FollowUp: m.FollowUp,
		CreatedBy:    cby,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    m.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
