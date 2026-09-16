package service

import (
	"context"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type AuditService struct {
	repo     *repository.AuditRepo
	userRepo *repository.UserAdminRepo
}

func NewAuditService(repo *repository.AuditRepo, userRepo *repository.UserAdminRepo) *AuditService {
	return &AuditService{repo: repo, userRepo: userRepo}
}

func (s *AuditService) GetAuditLogs(ctx context.Context, userID, targetType string, limit int) ([]model.AuditLogDTO, error) {
	rows, err := s.repo.FindAll(ctx, userID, targetType, limit)
	if err != nil {
		return nil, err
	}
	out := make([]model.AuditLogDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAuditDTO(a))
	}
	return out, nil
}

// Log: helper untuk dipakai service lain.
// Fire-and-forget — error di-log tapi tidak menggagalkan operasi utama.
func (s *AuditService) Log(ctx context.Context, userID, action, targetType, targetID string) {
	userNama := ""
	var uidPtr *string
	if userID != "" {
		uidPtr = &userID
		if u, err := s.userRepo.FindByID(ctx, userID); err == nil {
			userNama = u.Nama
		}
	}
	_ = s.repo.Insert(ctx, util.NewID("LOG"), uidPtr, userNama, action, targetType, targetID)
}

func toAuditDTO(a model.AuditLog) model.AuditLogDTO {
	uid := ""
	if a.UserID != nil {
		uid = *a.UserID
	}
	return model.AuditLogDTO{
		LogID:      a.LogID,
		UserID:     uid,
		UserNama:   a.UserNama,
		Action:     a.Action,
		TargetType: a.TargetType,
		TargetID:   a.TargetID,
		Timestamp:  a.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
