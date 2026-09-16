package service

import (
	"context"
	"strings"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type MemberService struct {
	repo *repository.MemberRepo
}

func NewMemberService(repo *repository.MemberRepo) *MemberService {
	return &MemberService{repo: repo}
}

func (s *MemberService) Repo() *repository.MemberRepo { return s.repo }

func (s *MemberService) GetMembers(ctx context.Context, f model.MemberListFilter) ([]model.MemberListDTO, error) {
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	filtered := s.applyFilters(all, f)

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	out := make([]model.MemberListDTO, 0, end-offset)
	for i := offset; i < end; i++ {
		out = append(out, s.toListDTO(filtered[i]))
	}
	return out, nil
}

func (s *MemberService) GetMembersPaged(ctx context.Context, f model.MemberListFilter) ([]model.MemberListDTO, int, error) {
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	filtered := s.applyFilters(all, f)
	total := len(filtered)

	limit := f.Limit
	if limit <= 0 {
		limit = 30
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]model.MemberListDTO, 0, end-offset)
	for i := offset; i < end; i++ {
		items = append(items, s.toListDTO(filtered[i]))
	}
	return items, total, nil
}

func (s *MemberService) GetAttendanceMembers(ctx context.Context, f model.MemberListFilter) ([]model.AttendanceMemberDTO, error) {
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	filtered := s.applyFilters(all, f)
	out := make([]model.AttendanceMemberDTO, 0, len(filtered))
	for _, m := range filtered {
		out = append(out, model.AttendanceMemberDTO{
			MemberID:     m.MemberID,
			NamaLengkap:  m.NamaLengkap,
			Kelompok:     m.Kelompok,
			Kategori:     s.kategori(m),
			JenisKelamin: strOr(m.JenisKelamin, ""),
		})
	}
	return out, nil
}

func (s *MemberService) GetMemberDetail(ctx context.Context, id string) (*model.MemberDetailDTO, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := s.toDetailDTO(m)
	return &dto, nil
}

func (s *MemberService) applyFilters(in []model.Member, f model.MemberListFilter) []model.Member {
	out := make([]model.Member, 0, len(in))
	search := strings.ToLower(strings.TrimSpace(f.Search))
	kelompok := strings.TrimSpace(f.Kelompok)
	jk := strings.TrimSpace(f.JenisKelamin)
	desa := strings.TrimSpace(f.Desa)
	kategori := strings.TrimSpace(f.Kategori)

	for _, m := range in {
		if !f.IncludeInactive && !m.StatusAktif {
			continue
		}
		if kelompok != "" && m.Kelompok != kelompok {
			continue
		}
		if jk != "" && strOr(m.JenisKelamin, "") != jk {
			continue
		}
		if desa != "" && m.Desa != desa {
			continue
		}
		if kategori != "" && s.kategori(m) != kategori {
			continue
		}
		if search != "" {
			hay := strings.ToLower(m.NamaLengkap) + " " + strings.ToLower(m.NamaPanggilan)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func (s *MemberService) kategori(m model.Member) string {
	return util.GetMemberCategory(m.TanggalLahir, m.JenjangPendidikan, m.IsNikah)
}

func (s *MemberService) toListDTO(m model.Member) model.MemberListDTO {
	return model.MemberListDTO{
		MemberID:      m.MemberID,
		NamaLengkap:   m.NamaLengkap,
		NamaPanggilan: m.NamaPanggilan,
		JenisKelamin:  strOr(m.JenisKelamin, ""),
		Kelompok:      m.Kelompok,
		Kategori:      s.kategori(m),
		FotoURL:       m.FotoURL,
	}
}

func (s *MemberService) toDetailDTO(m *model.Member) model.MemberDetailDTO {
	return model.MemberDetailDTO{
		MemberID:               m.MemberID,
		NamaLengkap:            m.NamaLengkap,
		NamaPanggilan:          m.NamaPanggilan,
		JenisKelamin:           strOr(m.JenisKelamin, ""),
		TempatLahir:            m.TempatLahir,
		TanggalLahir:           util.FormatDate(m.TanggalLahir),
		FotoURL:                m.FotoURL,
		NoWA:                   m.NoWA,
		AlamatRumah:            m.AlamatRumah,
		Desa:                   m.Desa,
		Daerah:                 m.Daerah,
		Kelompok:               m.Kelompok,
		IsMuballigh:            m.IsMuballigh,
		IsKerja:                m.IsKerja,
		IsNikah:                m.IsNikah,
		TinggiBadan:            m.TinggiBadan,
		BeratBadan:             m.BeratBadan,
		Hobi:                   m.Hobi,
		Pekerjaan:              m.Pekerjaan,
		StatusPembinaan:        m.StatusPembinaan,
		StatusAktif:            m.StatusAktif,
		TanggalMasuk:           util.FormatDate(m.TanggalMasuk),
		TanggalKeluar:          util.FormatDate(m.TanggalKeluar),
		JenjangPendidikan:      m.JenjangPendidikan,
		Sekolah:                m.Sekolah,
		Jurusan:                m.Jurusan,
		TahunMulaiPendidikan:   m.TahunMulaiPendidikan,
		TahunSelesaiPendidikan: m.TahunSelesaiPendidikan,
		CreatedAt:              m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:              m.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		Kategori:               s.kategori(*m),
		Usia:                   util.GetAge(m.TanggalLahir),
		Pendidikan:             []any{},
	}
}

func strOr(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}
