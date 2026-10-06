package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	apperrors "pengajian-backend/internal/errors"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type AnnouncementService struct {
	repo       *repository.AnnouncementRepo
	groupRepo  *repository.GroupRepo
	memberRepo *repository.MemberRepo
}

func NewAnnouncementService(
	repo *repository.AnnouncementRepo,
	groupRepo *repository.GroupRepo,
	memberRepo *repository.MemberRepo,
) *AnnouncementService {
	return &AnnouncementService{
		repo:       repo,
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
	}
}

var jadwalRutin = []string{"Minggu", "Selasa", "Kamis"}

func (s *AnnouncementService) GetTemplates(ctx context.Context, groupID string, includeInactive bool) ([]model.AnnouncementTemplateDTO, error) {
	tpl, err := s.repo.FindTemplates(ctx, groupID, includeInactive)
	if err != nil {
		return nil, err
	}
	out := make([]model.AnnouncementTemplateDTO, 0, len(tpl))
	for _, t := range tpl {
		out = append(out, toTemplateDTO(t))
	}
	return out, nil
}

func (s *AnnouncementService) GetTemplateDetail(ctx context.Context, id string) (*model.AnnouncementTemplateDTO, error) {
	t, err := s.repo.FindTemplateByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "template tidak ditemukan")
	}
	dto := toTemplateDTO(*t)
	return &dto, nil
}

type CreateTemplateInput struct {
	GroupID      string
	TemplateName string
	Kode         string
	TemplateBody string
	IsActive     bool
}

var kodeRegex = regexp.MustCompile(`^[A-Z0-9_]{3,30}$`)

func (s *AnnouncementService) CreateTemplate(ctx context.Context, in CreateTemplateInput) (*model.AnnouncementTemplateDTO, error) {
	if strings.TrimSpace(in.TemplateName) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "name template wajib diisi")
	}
	kode := strings.ToUpper(strings.TrimSpace(in.Kode))
	if !kodeRegex.MatchString(kode) {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "kode harus 3-30 karakter (huruf, angka, underscore)")
	}
	if strings.TrimSpace(in.TemplateBody) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "isi template wajib diisi")
	}

	if existing, _ := s.repo.FindTemplateByKode(ctx, kode); existing != nil {
		return nil, apperrors.Wrap(apperrors.ErrConflict, "kode template sudah dipakai")
	}

	var grpPtr *string
	if in.GroupID != "" {
		grpPtr = &in.GroupID
	}

	t := &model.AnnouncementTemplate{
		TemplateID:   util.NewID("TPL"),
		GroupID:      grpPtr,
		TemplateName: strings.TrimSpace(in.TemplateName),
		Kode:         kode,
		TemplateBody: in.TemplateBody,
		IsActive:     true,
	}
	if err := s.repo.InsertTemplate(ctx, t); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindTemplateByID(ctx, t.TemplateID)
	dto := toTemplateDTO(*fresh)
	return &dto, nil
}

type UpdateTemplateInput struct {
	TemplateID   string
	TemplateName *string
	Kode         *string
	TemplateBody *string
	IsActive     *bool
}

func (s *AnnouncementService) UpdateTemplate(ctx context.Context, in UpdateTemplateInput) (*model.AnnouncementTemplateDTO, error) {
	if in.TemplateID == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "template_id wajib diisi")
	}
	if _, err := s.repo.FindTemplateByID(ctx, in.TemplateID); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "template tidak ditemukan")
	}

	patch := map[string]interface{}{}
	if in.TemplateName != nil {
		patch["template_name"] = strings.TrimSpace(*in.TemplateName)
	}
	if in.Kode != nil {
		kode := strings.ToUpper(strings.TrimSpace(*in.Kode))
		if !kodeRegex.MatchString(kode) {
			return nil, apperrors.Wrap(apperrors.ErrValidation, "kode tidak valid")
		}
		if existing, _ := s.repo.FindTemplateByKode(ctx, kode); existing != nil && existing.TemplateID != in.TemplateID {
			return nil, apperrors.Wrap(apperrors.ErrConflict, "kode template sudah dipakai")
		}
		patch["kode"] = kode
	}
	if in.TemplateBody != nil {
		patch["template_body"] = *in.TemplateBody
	}
	if in.IsActive != nil {
		patch["is_active"] = *in.IsActive
	}

	if err := s.repo.UpdateTemplate(ctx, in.TemplateID, patch); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindTemplateByID(ctx, in.TemplateID)
	dto := toTemplateDTO(*fresh)
	return &dto, nil
}

func (s *AnnouncementService) DeleteTemplate(ctx context.Context, id string) error {
	if id == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "template_id wajib diisi")
	}
	if _, err := s.repo.FindTemplateByID(ctx, id); err != nil {
		return apperrors.Wrap(apperrors.ErrNotFound, "template tidak ditemukan")
	}
	return s.repo.SoftDeleteTemplate(ctx, id)
}

func (s *AnnouncementService) CreateTemplateFromAnnouncement(ctx context.Context, namaTemplate, kode, sourceAnnouncementID, isiTemplate string) (*model.AnnouncementTemplateDTO, error) {
	text := ""
	if sourceAnnouncementID != "" {
		a, err := s.repo.FindAnnouncementByID(ctx, sourceAnnouncementID)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.ErrNotFound, "pengumuman sumber tidak ditemukan")
		}
		text = a.GeneratedText
	} else if isiTemplate != "" {
		text = isiTemplate
	}
	if strings.TrimSpace(text) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "isi template kosong")
	}
	return s.CreateTemplate(ctx, CreateTemplateInput{
		TemplateName: namaTemplate,
		Kode:         kode,
		TemplateBody: text,
		IsActive:     true,
	})
}

type GenerateAnnouncementInput struct {
	TemplateID string
	GroupID    string
	Date       string
	Time       string
	Event      string
	Topic      string
	Notes      string
	Signatory  string
}

type GenerateAnnouncementResult struct {
	GeneratedText string                 `json:"generated_text"`
	Warning       string                 `json:"warning"`
	Day           string                 `json:"day"`
	Data          map[string]interface{} `json:"data"`
}

func (s *AnnouncementService) Generate(ctx context.Context, in GenerateAnnouncementInput) (*GenerateAnnouncementResult, error) {
	if in.TemplateID == "" || in.GroupID == "" || in.Date == "" {
		return nil, errors.New("template_id, group_id, dan date wajib diisi")
	}

	tpl, err := s.repo.FindTemplateByID(ctx, in.TemplateID)
	if err != nil {
		return nil, errors.New("template tidak ditemukan")
	}
	group, err := s.groupRepo.FindByID(ctx, in.GroupID)
	if err != nil {
		return nil, errors.New("kelompok tidak ditemukan")
	}

	tgl, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("tanggal tidak valid")
	}
	day := util.GetHariFromDate(&tgl)

	warning := ""
	if !containsString(jadwalRutin, day) {
		warning = "Date ini bukan schedule rutin pengajian (" + strings.Join(jadwalRutin, "/") + ")."
	}

	signatory := in.Signatory
	if signatory == "" {
		signatory = group.Signatory
	}

	data := map[string]interface{}{
		"nama_kelompok": group.GroupName,
		"day":           day,
		"date":          formatDateShort(tgl),
		"time":          in.Time,
		"event":         in.Event,
		"topic":         in.Topic,
		"notes":         in.Notes,
		"signatory":     signatory,
		// Alias kompatibel untuk template lama yang masih memakai placeholder Indonesia.
		"hari":          day,
		"tanggal":       formatDateShort(tgl),
		"jam":           in.Time,
		"acara":         in.Event,
		"materi":        in.Topic,
		"catatan":       in.Notes,
		"penandatangan": signatory,
	}

	text := renderTemplate(tpl.TemplateBody, data)

	return &GenerateAnnouncementResult{
		GeneratedText: text,
		Warning:       warning,
		Day:           day,
		Data:          data,
	}, nil
}

type CreateAnnouncementInput struct {
	TemplateID string
	MeetingID  string
	GroupID    string
	Date       string
	Time       string
	Event      string
	Topic      string
	Notes      string
	UserID     string
}

func (s *AnnouncementService) Create(ctx context.Context, in CreateAnnouncementInput) (*model.AnnouncementDTO, error) {
	gen, err := s.Generate(ctx, GenerateAnnouncementInput{
		TemplateID: in.TemplateID,
		GroupID:    in.GroupID,
		Date:       in.Date,
		Time:       in.Time,
		Event:      in.Event,
		Topic:      in.Topic,
		Notes:      in.Notes,
	})
	if err != nil {
		return nil, err
	}

	tgl, _ := time.Parse("2006-01-02", in.Date)
	a := &model.Announcement{
		AnnouncementID: util.NewID("ANN"),
		Date:           tgl,
		Day:            gen.Day,
		Time:           in.Time,
		Event:          in.Event,
		Topic:          in.Topic,
		Notes:          in.Notes,
		GeneratedText:  gen.GeneratedText,
		Status:         "DRAFT",
	}
	if in.TemplateID != "" {
		a.TemplateID = &in.TemplateID
	}
	if in.MeetingID != "" {
		a.MeetingID = &in.MeetingID
	}
	if in.GroupID != "" {
		a.GroupID = &in.GroupID
	}
	if in.UserID != "" {
		a.CreatedBy = &in.UserID
	}

	if err := s.repo.InsertAnnouncement(ctx, a); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindAnnouncementByID(ctx, a.AnnouncementID)
	dto := toAnnouncementDTO(*fresh)
	return &dto, nil
}

type UpdateAnnouncementInput struct {
	AnnouncementID string
	GeneratedText  *string
	Status         *string
	Time           *string
	Event          *string
	Topic          *string
	Notes          *string
}

func (s *AnnouncementService) Update(ctx context.Context, in UpdateAnnouncementInput) (*model.AnnouncementDTO, error) {
	if in.AnnouncementID == "" {
		return nil, errors.New("announcement_id wajib diisi")
	}
	if _, err := s.repo.FindAnnouncementByID(ctx, in.AnnouncementID); err != nil {
		return nil, errors.New("pengumuman tidak ditemukan")
	}

	patch := map[string]interface{}{}
	if in.GeneratedText != nil {
		patch["generated_text"] = *in.GeneratedText
	}
	if in.Status != nil {
		patch["status"] = *in.Status
	}
	if in.Time != nil {
		patch["time"] = *in.Time
	}
	if in.Event != nil {
		patch["event"] = *in.Event
	}
	if in.Topic != nil {
		patch["topic"] = *in.Topic
	}
	if in.Notes != nil {
		patch["notes"] = *in.Notes
	}

	if err := s.repo.UpdateAnnouncement(ctx, in.AnnouncementID, patch); err != nil {
		return nil, err
	}
	fresh, _ := s.repo.FindAnnouncementByID(ctx, in.AnnouncementID)
	dto := toAnnouncementDTO(*fresh)
	return &dto, nil
}

func (s *AnnouncementService) GetAnnouncements(ctx context.Context, groupID, status string) ([]model.AnnouncementDTO, error) {
	rows, err := s.repo.FindAnnouncements(ctx, groupID, status)
	if err != nil {
		return nil, err
	}
	out := make([]model.AnnouncementDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAnnouncementDTO(a))
	}
	return out, nil
}

func (s *AnnouncementService) GetRecipientSummary(ctx context.Context, groupID string) (map[string]int, error) {
	if groupID == "" {
		return nil, errors.New("group_id wajib diisi")
	}
	members, err := s.memberRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	total, withWA := 0, 0
	for _, m := range members {
		if m.GroupLabel != groupID || !m.IsActive {
			continue
		}
		total++
		if m.WhatsappNumber != "" {
			withWA++
		}
	}
	return map[string]int{
		"total":     total,
		"dengan_wa": withWA,
		"tanpa_wa":  total - withWA,
	}, nil
}

func toTemplateDTO(t model.AnnouncementTemplate) model.AnnouncementTemplateDTO {
	gid := ""
	if t.GroupID != nil {
		gid = *t.GroupID
	}
	return model.AnnouncementTemplateDTO{
		TemplateID:   t.TemplateID,
		GroupID:      gid,
		TemplateName: t.TemplateName,
		Kode:         t.Kode,
		TemplateBody: t.TemplateBody,
		IsActive:     t.IsActive,
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    t.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

func toAnnouncementDTO(a model.Announcement) model.AnnouncementDTO {
	tid, mid, gid, cby := "", "", "", ""
	if a.TemplateID != nil {
		tid = *a.TemplateID
	}
	if a.MeetingID != nil {
		mid = *a.MeetingID
	}
	if a.GroupID != nil {
		gid = *a.GroupID
	}
	if a.CreatedBy != nil {
		cby = *a.CreatedBy
	}
	return model.AnnouncementDTO{
		AnnouncementID: a.AnnouncementID,
		TemplateID:     tid,
		MeetingID:      mid,
		GroupID:        gid,
		Date:           a.Date.Format("2006-01-02"),
		Day:            a.Day,
		Time:           a.Time,
		Event:          a.Event,
		Topic:          a.Topic,
		Notes:          a.Notes,
		GeneratedText:  a.GeneratedText,
		Status:         a.Status,
		CreatedBy:      cby,
		CreatedAt:      a.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:      a.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

var tplRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func renderTemplate(text string, data map[string]interface{}) string {
	return tplRegex.ReplaceAllStringFunc(text, func(match string) string {
		sub := tplRegex.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := sub[1]
		if v, ok := data[key]; ok && v != nil {
			return toStringValue(v)
		}
		return ""
	})
}

func toStringValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	}
	return ""
}

func formatDateShort(t time.Time) string {
	bulan := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return strconv.Itoa(t.Day()) + " " + bulan[int(t.Month())-1] + " " + strconv.Itoa(t.Year())
}

func containsString(arr []string, s string) bool {
	for _, x := range arr {
		if x == s {
			return true
		}
	}
	return false
}

type WeeklyGenerateInput struct {
	TemplateID string
	GroupID    string
	WeekStart  string
	Time       string
	Event      string
	Topic      string
	Notes      string
	Signatory  string
}

type WeeklyGenerateResult struct {
	Day           string                 `json:"day"`
	Date          string                 `json:"date"`
	GeneratedText string                 `json:"generated_text"`
	Warning       string                 `json:"warning"`
	Data          map[string]interface{} `json:"data"`
}

func (s *AnnouncementService) GenerateWeekly(ctx context.Context, in WeeklyGenerateInput) ([]WeeklyGenerateResult, error) {
	if in.TemplateID == "" || in.GroupID == "" || in.WeekStart == "" {
		return nil, errors.New("template_id, group_id, dan week_start wajib diisi")
	}

	base, err := time.Parse("2006-01-02", in.WeekStart)
	if err != nil {
		return nil, errors.New("week_start tidak valid (YYYY-MM-DD)")
	}

	dayOffsets := []struct {
		Day   string
		Delta int
	}{
		{"Minggu", 0},
		{"Selasa", 2},
		{"Kamis", 4},
	}

	out := make([]WeeklyGenerateResult, 0, len(dayOffsets))
	for _, d := range dayOffsets {
		date := base.AddDate(0, 0, d.Delta)
		res, err := s.Generate(ctx, GenerateAnnouncementInput{
			TemplateID: in.TemplateID,
			GroupID:    in.GroupID,
			Date:       date.Format("2006-01-02"),
			Time:       in.Time,
			Event:      in.Event,
			Topic:      in.Topic,
			Notes:      in.Notes,
			Signatory:  in.Signatory,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, WeeklyGenerateResult{
			Day:           d.Day,
			Date:          date.Format("2006-01-02"),
			GeneratedText: res.GeneratedText,
			Warning:       res.Warning,
			Data:          res.Data,
		})
	}
	return out, nil
}
