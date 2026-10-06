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
	GroupID           string
	Date              string
	SermonLeader      string
	Muadzin           string
	Advisor           string
	ParkingAttendant  string
	FootwearAttendant string
	Notes             string
	CreatedBy         string
}

func (s *FridayService) Save(ctx context.Context, in SaveFridayInput) (*model.FridayScheduleDTO, error) {
	if in.Date == "" {
		return nil, errors.New("date wajib diisi")
	}
	tgl, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	if tgl.Weekday() != time.Friday {
		return nil, errors.New("date harus jatuh di day Jumat")
	}
	if in.SermonLeader == "" && in.Muadzin == "" && in.Advisor == "" &&
		in.ParkingAttendant == "" && in.FootwearAttendant == "" {
		return nil, errors.New("minimal satu petugas wajib diisi")
	}

	var groupIDPtr *string
	if in.GroupID != "" {
		groupIDPtr = &in.GroupID
	}

	f := &model.FridaySchedule{
		FridayID:          util.NewID("JMT"),
		GroupID:           groupIDPtr,
		Date:              tgl,
		SermonLeader:      util.TitleCaseID(in.SermonLeader),
		Muadzin:           util.TitleCaseID(in.Muadzin),
		Advisor:           util.TitleCaseID(in.Advisor),
		ParkingAttendant:  util.TitleCaseID(in.ParkingAttendant),
		FootwearAttendant: util.TitleCaseID(in.FootwearAttendant),
		Notes:             in.Notes,
		CreatedBy:         in.CreatedBy,
	}
	if err := s.repo.Upsert(ctx, f); err != nil {
		return nil, err
	}

	fresh, err := s.repo.FindByDate(ctx, in.GroupID, in.Date)
	if err != nil || fresh == nil {
		return nil, errors.New("gagal ambil data schedule jumat")
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

func (s *FridayService) Delete(ctx context.Context, groupID, date string) error {
	if date == "" {
		return errors.New("date wajib diisi")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	return s.repo.Delete(ctx, groupID, date)
}

func toFridayDTO(f model.FridaySchedule) model.FridayScheduleDTO {
	grpID := ""
	if f.GroupID != nil {
		grpID = *f.GroupID
	}
	return model.FridayScheduleDTO{
		FridayID:          f.FridayID,
		GroupID:           grpID,
		Date:              f.Date.Format("2006-01-02"),
		Day:               util.GetHariFromDate(&f.Date),
		SermonLeader:      f.SermonLeader,
		Muadzin:           f.Muadzin,
		Advisor:           f.Advisor,
		ParkingAttendant:  f.ParkingAttendant,
		FootwearAttendant: f.FootwearAttendant,
		Notes:             f.Notes,
		CreatedBy:         f.CreatedBy,
		CreatedAt:         f.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:         f.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}
