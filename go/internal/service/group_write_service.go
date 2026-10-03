package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type SaveGroupInput struct {
	GroupID       string
	GroupCode     string
	GroupName     string
	Mentor       string
	Signatory string
	Schedule        string
	IsActive   *bool
}

func (s *GroupService) Save(ctx context.Context, in SaveGroupInput) (*model.GroupDTO, error) {
	if in.GroupID != "" {
		existing, err := s.repo.FindByID(ctx, in.GroupID)
		if err != nil || existing == nil {
			return nil, errors.New("kelompok tidak ditemukan")
		}
		patch := map[string]interface{}{}
		if in.GroupCode != "" {
			patch["group_code"] = in.GroupCode
		}
		if in.GroupName != "" {
			patch["group_name"] = in.GroupName
		}
		if in.Mentor != "" {
			patch["mentor"] = in.Mentor
		}
		if in.Signatory != "" {
			patch["signatory"] = in.Signatory
		}
		if in.Schedule != "" {
			patch["schedule"] = in.Schedule
		}
		if in.IsActive != nil {
			patch["is_active"] = *in.IsActive
		}
		if len(patch) == 0 {
			return nil, errors.New("tidak ada perubahan")
		}
		if err := s.repo.Update(ctx, in.GroupID, patch); err != nil {
			return nil, err
		}
		fresh, _ := s.repo.FindByID(ctx, in.GroupID)
		dto := toGroupDTO(*fresh)
		return &dto, nil
	}

	if strings.TrimSpace(in.GroupName) == "" {
		return nil, errors.New("group_name wajib diisi")
	}
	groupID := util.NewID("GRP")
	status := true
	if in.IsActive != nil {
		status = *in.IsActive
	}
	g := &model.Group{
		GroupID:       groupID,
		GroupCode:     in.GroupCode,
		GroupName:     in.GroupName,
		Mentor:       in.Mentor,
		Signatory: in.Signatory,
		Schedule:        in.Schedule,
		IsActive:   status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := s.repo.Insert(ctx, g); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindByID(ctx, groupID)
	dto := toGroupDTO(*fresh)
	return &dto, nil
}

var _ = repository.NewGroupRepo
