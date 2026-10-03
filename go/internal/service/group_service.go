package service

import (
	"context"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
)

type GroupService struct {
	repo *repository.GroupRepo
}

func NewGroupService(repo *repository.GroupRepo) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) GetGroups(ctx context.Context, includeInactive bool) ([]model.GroupDTO, error) {
	groups, err := s.repo.FindAll(ctx, includeInactive)
	if err != nil {
		return nil, err
	}
	out := make([]model.GroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, toGroupDTO(g))
	}
	return out, nil
}

func toGroupDTO(g model.Group) model.GroupDTO {
	return model.GroupDTO{
		GroupID:       g.GroupID,
		GroupCode:     g.GroupCode,
		GroupName:     g.GroupName,
		Mentor:       g.Mentor,
		Signatory: g.Signatory,
		Schedule:        g.Schedule,
		IsActive:   g.IsActive,
		CreatedAt:     g.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:     g.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
