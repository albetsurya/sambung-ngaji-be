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

/* ===== Templates ===== */

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
	NamaTemplate string
	Kode         string
	IsiTemplate  string
	StatusAktif  bool
}

var kodeRegex = regexp.MustCompile(`^[A-Z0-9_]{3,30}$`)

func (s *AnnouncementService) CreateTemplate(ctx context.Context, in CreateTemplateInput) (*model.AnnouncementTemplateDTO, error) {
	if strings.TrimSpace(in.NamaTemplate) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "nama template wajib diisi")
	}
	kode := strings.ToUpper(strings.TrimSpace(in.Kode))
	if !kodeRegex.MatchString(kode) {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "kode harus 3-30 karakter (huruf, angka, underscore)")
	}
	if strings.TrimSpace(in.IsiTemplate) == "" {
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
		NamaTemplate: strings.TrimSpace(in.NamaTemplate),
		Kode:         kode,
		IsiTemplate:  in.IsiTemplate,
		StatusAktif:  true,
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
	NamaTemplate *string
	Kode         *string
	IsiTemplate  *string
	StatusAktif  *bool
}

func (s *AnnouncementService) UpdateTemplate(ctx context.Context, in UpdateTemplateInput) (*model.AnnouncementTemplateDTO, error) {
	if in.TemplateID == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "template_id wajib diisi")
	}
	if _, err := s.repo.FindTemplateByID(ctx, in.TemplateID); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "template tidak ditemukan")
	}

	patch := map[string]interface{}{}
	if in.NamaTemplate != nil {
		patch["nama_template"] = strings.TrimSpace(*in.NamaTemplate)
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
	if in.IsiTemplate != nil {
		patch["isi_template"] = *in.IsiTemplate
	}
	if in.StatusAktif != nil {
		patch["status_aktif"] = *in.StatusAktif
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
		NamaTemplate: namaTemplate,
		Kode:         kode,
		IsiTemplate:  text,
		StatusAktif:  true,
	})
}

/* ===== Generate ===== */

type GenerateAnnouncementInput struct {
	TemplateID    string
	GroupID       string
	Tanggal       string
	Jam           string
	Acara         string
	Materi        string
	Catatan       string
	Penandatangan string
}

type GenerateAnnouncementResult struct {
	GeneratedText string                 `json:"generated_text"`
	Warning       string                 `json:"warning"`
	Hari          string                 `json:"hari"`
	Data          map[string]interface{} `json:"data"`
}

func (s *AnnouncementService) Generate(ctx context.Context, in GenerateAnnouncementInput) (*GenerateAnnouncementResult, error) {
	if in.TemplateID == "" || in.GroupID == "" || in.Tanggal == "" {
		return nil, errors.New("template_id, group_id, dan tanggal wajib diisi")
	}

	tpl, err := s.repo.FindTemplateByID(ctx, in.TemplateID)
	if err != nil {
		return nil, errors.New("template tidak ditemukan")
	}
	group, err := s.groupRepo.FindByID(ctx, in.GroupID)
	if err != nil {
		return nil, errors.New("kelompok tidak ditemukan")
	}

	tgl, err := time.Parse("2006-01-02", in.Tanggal)
	if err != nil {
		return nil, errors.New("tanggal tidak valid")
	}
	hari := util.GetHariFromDate(&tgl)

	warning := ""
	if !containsString(jadwalRutin, hari) {
		warning = "Tanggal ini bukan jadwal rutin pengajian (" + strings.Join(jadwalRutin, "/") + ")."
	}

	penandatangan := in.Penandatangan
	if penandatangan == "" {
		penandatangan = group.Penandatangan
	}

	data := map[string]interface{}{
		"nama_kelompok": group.GroupName,
		"hari":          hari,
		"tanggal":       formatDateShort(tgl),
		"jam":           in.Jam,
		"acara":         in.Acara,
		"materi":        in.Materi,
		"catatan":       in.Catatan,
		"penandatangan": penandatangan,
	}

	text := renderTemplate(tpl.IsiTemplate, data)

	return &GenerateAnnouncementResult{
		GeneratedText: text,
		Warning:       warning,
		Hari:          hari,
		Data:          data,
	}, nil
}

/* ===== CRUD Announcements ===== */

type CreateAnnouncementInput struct {
	TemplateID string
	MeetingID  string
	GroupID    string
	Tanggal    string
	Jam        string
	Acara      string
	Materi     string
	Catatan    string
	UserID     string
}

func (s *AnnouncementService) Create(ctx context.Context, in CreateAnnouncementInput) (*model.AnnouncementDTO, error) {
	gen, err := s.Generate(ctx, GenerateAnnouncementInput{
		TemplateID: in.TemplateID,
		GroupID:    in.GroupID,
		Tanggal:    in.Tanggal,
		Jam:        in.Jam,
		Acara:      in.Acara,
		Materi:     in.Materi,
		Catatan:    in.Catatan,
	})
	if err != nil {
		return nil, err
	}

	tgl, _ := time.Parse("2006-01-02", in.Tanggal)
	a := &model.Announcement{
		AnnouncementID: util.NewID("ANN"),
		Tanggal:        tgl,
		Hari:           gen.Hari,
		Jam:            in.Jam,
		Acara:          in.Acara,
		Materi:         in.Materi,
		Catatan:        in.Catatan,
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
	Jam            *string
	Acara          *string
	Materi         *string
	Catatan        *string
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
	if in.Jam != nil {
		patch["jam"] = *in.Jam
	}
	if in.Acara != nil {
		patch["acara"] = *in.Acara
	}
	if in.Materi != nil {
		patch["materi"] = *in.Materi
	}
	if in.Catatan != nil {
		patch["catatan"] = *in.Catatan
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
		if m.Kelompok != groupID || !m.StatusAktif {
			continue
		}
		total++
		if m.NoWA != "" {
			withWA++
		}
	}
	return map[string]int{
		"total":     total,
		"dengan_wa": withWA,
		"tanpa_wa":  total - withWA,
	}, nil
}

/* ===== Helpers ===== */

func toTemplateDTO(t model.AnnouncementTemplate) model.AnnouncementTemplateDTO {
	gid := ""
	if t.GroupID != nil {
		gid = *t.GroupID
	}
	return model.AnnouncementTemplateDTO{
		TemplateID:   t.TemplateID,
		GroupID:      gid,
		NamaTemplate: t.NamaTemplate,
		Kode:         t.Kode,
		IsiTemplate:  t.IsiTemplate,
		StatusAktif:  t.StatusAktif,
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
		Tanggal:        a.Tanggal.Format("2006-01-02"),
		Hari:           a.Hari,
		Jam:            a.Jam,
		Acara:          a.Acara,
		Materi:         a.Materi,
		Catatan:        a.Catatan,
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
	TemplateID    string
	GroupID       string
	WeekStart     string
	Jam           string
	Acara         string
	Materi        string
	Catatan       string
	Penandatangan string
}

type WeeklyGenerateResult struct {
	Hari          string                 `json:"hari"`
	Tanggal       string                 `json:"tanggal"`
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
		Hari  string
		Delta int
	}{
		{"Minggu", 0},
		{"Selasa", 2},
		{"Kamis", 4},
	}

	out := make([]WeeklyGenerateResult, 0, len(dayOffsets))
	for _, d := range dayOffsets {
		tanggal := base.AddDate(0, 0, d.Delta)
		res, err := s.Generate(ctx, GenerateAnnouncementInput{
			TemplateID:    in.TemplateID,
			GroupID:       in.GroupID,
			Tanggal:       tanggal.Format("2006-01-02"),
			Jam:           in.Jam,
			Acara:         in.Acara,
			Materi:        in.Materi,
			Catatan:       in.Catatan,
			Penandatangan: in.Penandatangan,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, WeeklyGenerateResult{
			Hari:          d.Hari,
			Tanggal:       tanggal.Format("2006-01-02"),
			GeneratedText: res.GeneratedText,
			Warning:       res.Warning,
			Data:          res.Data,
		})
	}
	return out, nil
}
