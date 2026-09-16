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
	Pembina       string
	Penandatangan string
	Jadwal        string
	StatusAktif   *bool
}

func (s *GroupService) Save(ctx context.Context, in SaveGroupInput) (*model.GroupDTO, error) {
	if in.GroupID != "" {
		// UPDATE
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
		if in.Pembina != "" {
			patch["pembina"] = in.Pembina
		}
		if in.Penandatangan != "" {
			patch["penandatangan"] = in.Penandatangan
		}
		if in.Jadwal != "" {
			patch["jadwal"] = in.Jadwal
		}
		if in.StatusAktif != nil {
			patch["status_aktif"] = *in.StatusAktif
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

	// CREATE
	if strings.TrimSpace(in.GroupName) == "" {
		return nil, errors.New("group_name wajib diisi")
	}
	groupID := util.NewID("GRP")
	status := true
	if in.StatusAktif != nil {
		status = *in.StatusAktif
	}
	g := &model.Group{
		GroupID:       groupID,
		GroupCode:     in.GroupCode,
		GroupName:     in.GroupName,
		Pembina:       in.Pembina,
		Penandatangan: in.Penandatangan,
		Jadwal:        in.Jadwal,
		StatusAktif:   status,
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

// Pastikan import repository terpakai
var _ = repository.NewGroupRepo
