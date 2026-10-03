package service

import (
	"context"
	"errors"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

var validMoodKeys = map[string]bool{
	"sedih":     true,
	"cemas":     true,
	"syukur":    true,
	"marah":     true,
	"lelah":     true,
	"takut":     true,
	"putus-asa": true,
	"tenang":    true,
}

type MoodService struct {
	repo       *repository.MoodRepo
	memberRepo *repository.MemberRepo
}

func NewMoodService(repo *repository.MoodRepo, memberRepo *repository.MemberRepo) *MoodService {
	return &MoodService{repo: repo, memberRepo: memberRepo}
}

type SaveMoodInput struct {
	MemberID string
	MoodKey  string
	Date  string
}

func (s *MoodService) Save(ctx context.Context, in SaveMoodInput) (*model.MemberMoodDTO, error) {
	if in.MemberID == "" || in.MoodKey == "" {
		return nil, errors.New("member_id dan mood wajib diisi")
	}
	if !validMoodKeys[in.MoodKey] {
		return nil, errors.New("mood tidak valid")
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

	m := &model.MemberMood{
		MoodID:   util.NewID("MOD"),
		MemberID: in.MemberID,
		MoodKey:  in.MoodKey,
		Date:  tgl,
	}
	if err := s.repo.Upsert(ctx, m); err != nil {
		return nil, err
	}

	fresh, err := s.repo.FindByMemberAndDate(ctx, in.MemberID, tgl.Format("2006-01-02"))
	if err != nil || fresh == nil {
		return nil, errors.New("gagal ambil data mood")
	}
	dto := toMoodDTO(*fresh)
	return &dto, nil
}

func (s *MoodService) GetByMember(ctx context.Context, memberID string, limit int) ([]model.MemberMoodDTO, error) {
	rows, err := s.repo.FindByMember(ctx, memberID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]model.MemberMoodDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMoodDTO(r))
	}
	return out, nil
}

func toMoodDTO(m model.MemberMood) model.MemberMoodDTO {
	return model.MemberMoodDTO{
		MoodID:    m.MoodID,
		MemberID:  m.MemberID,
		MoodKey:   m.MoodKey,
		Date:   m.Date.Format("2006-01-02"),
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
