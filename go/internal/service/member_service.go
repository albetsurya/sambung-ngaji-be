package service

import (
	"context"
	"strings"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type MemberService struct {
	repo      *repository.MemberRepo
	groupRepo *repository.GroupRepo
}

func NewMemberService(repo *repository.MemberRepo) *MemberService {
	return &MemberService{repo: repo}
}

func (s *MemberService) Repo() *repository.MemberRepo { return s.repo }

func (s *MemberService) SetGroupRepo(gr *repository.GroupRepo) { s.groupRepo = gr }
func (s *MemberService) GroupRepo() *repository.GroupRepo      { return s.groupRepo }

func (s *MemberService) GetMembers(ctx context.Context, f model.MemberListFilter, groupID string) ([]model.MemberListDTO, error) {
	all, err := s.repo.FindAllByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	filtered := s.applyFilters(all, f)
	withUsers, err := s.repo.MemberIDsWithUsers(ctx)
	if err != nil {
		return nil, err
	}

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
		out = append(out, s.toListDTO(filtered[i], withUsers[filtered[i].MemberID]))
	}
	return out, nil
}

func (s *MemberService) GetMembersPaged(ctx context.Context, f model.MemberListFilter, groupID string) ([]model.MemberListDTO, int, error) {
	all, err := s.repo.FindAllByGroup(ctx, groupID)
	if err != nil {
		return nil, 0, err
	}
	filtered := s.applyFilters(all, f)
	total := len(filtered)
	withUsers, err := s.repo.MemberIDsWithUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

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
		items = append(items, s.toListDTO(filtered[i], withUsers[filtered[i].MemberID]))
	}
	return items, total, nil
}

func (s *MemberService) GetAttendanceMembers(ctx context.Context, f model.MemberListFilter, groupID string) ([]model.AttendanceMemberDTO, error) {
	all, err := s.repo.FindAllByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	filtered := s.applyFilters(all, f)
	out := make([]model.AttendanceMemberDTO, 0, len(filtered))
	for _, m := range filtered {
		gid := ""
		if m.GroupID != nil {
			gid = *m.GroupID
		}
		out = append(out, model.AttendanceMemberDTO{
			MemberID:     m.MemberID,
			GroupID:      gid,
			FullName:  m.FullName,
			GroupLabel:     m.GroupLabel,
			Kategori:     s.kategori(m),
			Gender: strOr(m.Gender, ""),
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
	hasUser, err := s.repo.MemberHasActiveUser(ctx, id)
	if err != nil {
		return nil, err
	}
	dto.HasUser = hasUser
	return &dto, nil
}

func (s *MemberService) applyFilters(in []model.Member, f model.MemberListFilter) []model.Member {
	out := make([]model.Member, 0, len(in))
	search := strings.ToLower(strings.TrimSpace(f.Search))
	group_label := strings.TrimSpace(f.GroupLabel)
	jk := strings.TrimSpace(f.Gender)
	village := strings.TrimSpace(f.Village)
	kategori := strings.TrimSpace(f.Kategori)

	for _, m := range in {
		if !f.IncludeInactive && !m.IsActive {
			continue
		}
		if group_label != "" && m.GroupLabel != group_label && (m.GroupID == nil || *m.GroupID != group_label) {
			continue
		}
		if jk != "" && strOr(m.Gender, "") != jk {
			continue
		}
		if village != "" && m.Village != village {
			continue
		}
		if kategori != "" && s.kategori(m) != kategori {
			continue
		}
		if search != "" {
			hay := strings.ToLower(m.FullName) + " " + strings.ToLower(m.Nickname)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func (s *MemberService) kategori(m model.Member) string {
	return util.GetMemberCategory(m.BirthDate, m.EducationLevel, m.IsMarried)
}

func (s *MemberService) toListDTO(m model.Member, hasUser bool) model.MemberListDTO {
	gid := ""
	if m.GroupID != nil {
		gid = *m.GroupID
	}
	return model.MemberListDTO{
		MemberID:      m.MemberID,
		GroupID:       gid,
		FullName:   m.FullName,
		Nickname: m.Nickname,
		Gender:  strOr(m.Gender, ""),
		GroupLabel:      m.GroupLabel,
		GroupName:     m.GroupLabel,
		Kategori:      s.kategori(m),
		PhotoURL:       m.PhotoURL,
		HasUser:       hasUser,
	}
}

func (s *MemberService) toDetailDTO(m *model.Member) model.MemberDetailDTO {
	gid := ""
	if m.GroupID != nil {
		gid = *m.GroupID
	}
	return model.MemberDetailDTO{
		MemberID:               m.MemberID,
		GroupID:                gid,
		FullName:            m.FullName,
		Nickname:          m.Nickname,
		Gender:           strOr(m.Gender, ""),
		BirthPlace:            m.BirthPlace,
		BirthDate:           util.FormatDate(m.BirthDate),
		PhotoURL:                m.PhotoURL,
		WhatsappNumber:                   m.WhatsappNumber,
		HomeAddress:            m.HomeAddress,
		Village:                   m.Village,
		Region:                 m.Region,
		GroupLabel:               m.GroupLabel,
		GroupName:              m.GroupLabel,
		IsPreacher:            m.IsPreacher,
		IsEmployed:                m.IsEmployed,
		IsMarried:                m.IsMarried,
		Height:            m.Height,
		Weight:             m.Weight,
		Hobby:                   m.Hobby,
		Occupation:              m.Occupation,
		MentoringStatus:        m.MentoringStatus,
		IsActive:            m.IsActive,
		JoinedDate:           util.FormatDate(m.JoinedDate),
		LeftDate:          util.FormatDate(m.LeftDate),
		EducationLevel:      m.EducationLevel,
		School:                m.School,
		Major:                m.Major,
		EducationStartYear:   m.EducationStartYear,
		EducationEndYear: m.EducationEndYear,
		CreatedAt:              m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:              m.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		Kategori:               s.kategori(*m),
		Usia:                   util.GetAge(m.BirthDate),
		Pendidikan:             []any{},
	}
}

func strOr(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}
