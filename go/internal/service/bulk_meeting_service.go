package service

import (
	"context"
	"time"

	apperrors "pengajian-backend/internal/errors"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type BulkMeetingService struct {
	meetingRepo *repository.MeetingRepo
	groupRepo   *repository.GroupRepo
	annRepo     *repository.AnnouncementRepo
}

func NewBulkMeetingService(
	meetingRepo *repository.MeetingRepo,
	groupRepo *repository.GroupRepo,
	annRepo *repository.AnnouncementRepo,
) *BulkMeetingService {
	return &BulkMeetingService{
		meetingRepo: meetingRepo,
		groupRepo:   groupRepo,
		annRepo:     annRepo,
	}
}

var hariIndex = map[string]int{
	"Minggu": 0, "Senin": 1, "Selasa": 2, "Rabu": 3,
	"Kamis": 4, "Jumat": 5, "Sabtu": 6,
}

type BulkParams struct {
	Tahun            int
	Bulan            int
	Day              []string
	Time             string
	Event            string
	GroupID          string
	Topic            string
	Notes            string
	TargetCategories []string
}

func validateBulk(p BulkParams) error {
	if p.Tahun == 0 || p.Bulan == 0 {
		return apperrors.Wrap(apperrors.ErrValidation, "tahun dan bulan wajib diisi")
	}
	if p.Tahun < 2020 || p.Tahun > 2100 {
		return apperrors.Wrap(apperrors.ErrValidation, "tahun tidak valid")
	}
	if p.Bulan < 1 || p.Bulan > 12 {
		return apperrors.Wrap(apperrors.ErrValidation, "bulan tidak valid")
	}
	if len(p.Day) == 0 {
		return apperrors.Wrap(apperrors.ErrValidation, "day wajib dipilih minimal 1")
	}
	if p.Time == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "time wajib diisi")
	}
	if p.Event == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "event wajib diisi")
	}
	if p.GroupID == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "kelompok wajib dipilih")
	}
	return nil
}

func buildDateList(p BulkParams) []time.Time {
	dayIdx := []int{}
	for _, h := range p.Day {
		if idx, ok := hariIndex[h]; ok {
			dayIdx = append(dayIdx, idx)
		}
	}
	if len(dayIdx) == 0 {
		return nil
	}

	daysInMonth := time.Date(p.Tahun, time.Month(p.Bulan)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	out := []time.Time{}
	for d := 1; d <= daysInMonth; d++ {
		t := time.Date(p.Tahun, time.Month(p.Bulan), d, 0, 0, 0, 0, time.UTC)
		wd := int(t.Weekday())
		for _, idx := range dayIdx {
			if idx == wd {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

type BulkPreviewItem struct {
	Date           string `json:"date"`
	TanggalDisplay string `json:"tanggal_display"`
	Day            string `json:"day"`
	SudahAda       bool   `json:"sudah_ada"`
}

type BulkPreviewResult struct {
	TotalDates    int               `json:"total_dates"`
	TotalNew      int               `json:"total_new"`
	TotalExisting int               `json:"total_existing"`
	Meetings      []BulkPreviewItem `json:"meetings"`
}

func (s *BulkMeetingService) Preview(ctx context.Context, p BulkParams) (*BulkPreviewResult, error) {
	if err := validateBulk(p); err != nil {
		return nil, err
	}
	dates := buildDateList(p)
	if len(dates) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "tidak ada date valid di bulan ini untuk day yang dipilih")
	}

	existing, err := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})
	if err != nil {
		return nil, err
	}
	existingKey := map[string]bool{}
	for _, m := range existing {
		gid := ""
		if m.GroupID != nil {
			gid = *m.GroupID
		}
		existingKey[m.Date.Format("2006-01-02")+"|"+gid] = true
	}

	items := make([]BulkPreviewItem, 0, len(dates))
	totalNew := 0
	for _, d := range dates {
		iso := d.Format("2006-01-02")
		sudahAda := existingKey[iso+"|"+p.GroupID]
		if !sudahAda {
			totalNew++
		}
		items = append(items, BulkPreviewItem{
			Date:           iso,
			TanggalDisplay: formatDateShort(d),
			Day:            util.GetHariFromDate(&d),
			SudahAda:       sudahAda,
		})
	}

	return &BulkPreviewResult{
		TotalDates:    len(dates),
		TotalNew:      totalNew,
		TotalExisting: len(dates) - totalNew,
		Meetings:      items,
	}, nil
}

type BulkCreateResult struct {
	Created  int               `json:"created"`
	Skipped  int               `json:"skipped"`
	Meetings []BulkCreatedItem `json:"meetings"`
}

type BulkCreatedItem struct {
	MeetingID      string `json:"meeting_id"`
	Date           string `json:"date"`
	TanggalDisplay string `json:"tanggal_display"`
	Day            string `json:"day"`
}

func (s *BulkMeetingService) BulkCreate(ctx context.Context, p BulkParams, userID string) (*BulkCreateResult, error) {
	if err := validateBulk(p); err != nil {
		return nil, err
	}
	dates := buildDateList(p)

	existing, _ := s.meetingRepo.FindAll(ctx, model.MeetingListFilter{})
	existingKey := map[string]bool{}
	for _, m := range existing {
		gid := ""
		if m.GroupID != nil {
			gid = *m.GroupID
		}
		existingKey[m.Date.Format("2006-01-02")+"|"+gid] = true
	}

	var toCreate []time.Time
	for _, d := range dates {
		if !existingKey[d.Format("2006-01-02")+"|"+p.GroupID] {
			toCreate = append(toCreate, d)
		}
	}
	if len(toCreate) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrConflict, "semua date sudah ada. tidak ada yang dibuat")
	}

	created := []BulkCreatedItem{}
	for _, d := range toCreate {
		mid := util.NewID("MTG")
		gid := p.GroupID
		m := &model.Meeting{
			MeetingID:        mid,
			Date:             d,
			Day:              util.GetHariFromDate(&d),
			Time:             p.Time,
			GroupID:          &gid,
			Event:            p.Event,
			Topic:            p.Topic,
			Status:           "SCHEDULED",
			Notes:            p.Notes,
			TargetCategories: p.TargetCategories,
		}
		if userID != "" {
			m.CreatedBy = &userID
		}
		if m.TargetCategories == nil {
			m.TargetCategories = []string{}
		}
		if err := s.meetingRepo.Create(ctx, m); err != nil {
			return nil, err
		}
		created = append(created, BulkCreatedItem{
			MeetingID:      mid,
			Date:           d.Format("2006-01-02"),
			TanggalDisplay: formatDateShort(d),
			Day:            util.GetHariFromDate(&d),
		})
	}

	return &BulkCreateResult{
		Created:  len(created),
		Skipped:  len(dates) - len(toCreate),
		Meetings: created,
	}, nil
}

type BulkTemplateDTO struct {
	TemplateID   string `json:"template_id"`
	TemplateName string `json:"template_name"`
	Kode         string `json:"kode"`
	TemplateBody string `json:"template_body"`
}

func (s *BulkMeetingService) GetTemplates(ctx context.Context, groupID string) ([]BulkTemplateDTO, error) {
	tpls, err := s.annRepo.FindTemplates(ctx, groupID, false)
	if err != nil {
		return nil, err
	}
	out := make([]BulkTemplateDTO, 0, len(tpls))
	for _, t := range tpls {
		out = append(out, BulkTemplateDTO{
			TemplateID:   t.TemplateID,
			TemplateName: t.TemplateName,
			Kode:         t.Kode,
			TemplateBody: t.TemplateBody,
		})
	}
	return out, nil
}
