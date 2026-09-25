package service

import (
	"context"
	"errors"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type FridayService struct {
	repo *repository.FridayRepo
}

func NewFridayService(repo *repository.FridayRepo) *FridayService {
	return &FridayService{repo: repo}
}

type SaveFridayInput struct {
	GroupID       string
	Tanggal       string
	KhatibImam    string
	Muadzin       string
	Penasihat     string
	PetugasParkir string
	PenataSandal  string
	Catatan       string
	CreatedBy     string
}

func (s *FridayService) Save(ctx context.Context, in SaveFridayInput) (*model.FridayScheduleDTO, error) {
	if in.Tanggal == "" {
		return nil, errors.New("tanggal wajib diisi")
	}
	tgl, err := time.Parse("2006-01-02", in.Tanggal)
	if err != nil {
		return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	if tgl.Weekday() != time.Friday {
		return nil, errors.New("tanggal harus jatuh di hari Jumat")
	}
	if in.KhatibImam == "" && in.Muadzin == "" && in.Penasihat == "" &&
		in.PetugasParkir == "" && in.PenataSandal == "" {
		return nil, errors.New("minimal satu petugas wajib diisi")
	}

	var groupIDPtr *string
	if in.GroupID != "" {
		groupIDPtr = &in.GroupID
	}

	f := &model.FridaySchedule{
		FridayID:      util.NewID("JMT"),
		GroupID:       groupIDPtr,
		Tanggal:       tgl,
		KhatibImam:    util.TitleCaseID(in.KhatibImam),
		Muadzin:       util.TitleCaseID(in.Muadzin),
		Penasihat:     util.TitleCaseID(in.Penasihat),
		PetugasParkir: util.TitleCaseID(in.PetugasParkir),
		PenataSandal:  util.TitleCaseID(in.PenataSandal),
		Catatan:       in.Catatan,
		CreatedBy:     in.CreatedBy,
	}
	if err := s.repo.Upsert(ctx, f); err != nil {
		return nil, err
	}

	fresh, err := s.repo.FindByDate(ctx, in.GroupID, in.Tanggal)
	if err != nil || fresh == nil {
		return nil, errors.New("gagal ambil data jadwal jumat")
	}
	dto := toFridayDTO(*fresh)
	return &dto, nil
}

func (s *FridayService) List(ctx context.Context, groupID, from, to string) ([]model.FridayScheduleDTO, error) {
	rows, err := s.repo.FindByRange(ctx, groupID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]model.FridayScheduleDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toFridayDTO(r))
	}
	return out, nil
}

func (s *FridayService) Delete(ctx context.Context, groupID, tanggal string) error {
	if tanggal == "" {
		return errors.New("tanggal wajib diisi")
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		return errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	return s.repo.Delete(ctx, groupID, tanggal)
}

func toFridayDTO(f model.FridaySchedule) model.FridayScheduleDTO {
	grpID := ""
	if f.GroupID != nil {
		grpID = *f.GroupID
	}
	return model.FridayScheduleDTO{
		FridayID:       f.FridayID,
		GroupID:        grpID,
		Tanggal:        f.Tanggal.Format("2006-01-02"),
		Hari:           util.GetHariFromDate(&f.Tanggal),
		KhatibImam:     f.KhatibImam,
		Muadzin:        f.Muadzin,
		Penasihat:      f.Penasihat,
		PetugasParkir:  f.PetugasParkir,
		PenataSandal:   f.PenataSandal,
		Catatan:        f.Catatan,
		CreatedBy:      f.CreatedBy,
		CreatedAt:      f.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:      f.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		ReminderSentAt: formatReminderSentAt(f.ReminderSentAt),
	}
}

func formatReminderSentAt(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05.000Z07:00")
}
