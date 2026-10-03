package service

import (
	"context"
	"strings"
	"time"

	apperrors "pengajian-backend/internal/errors"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

func (s *MemberService) resolveGroup(ctx context.Context, groupID, group_label string) (id, name string, err error) {
	groupID = strings.TrimSpace(groupID)
	group_label = strings.TrimSpace(group_label)
	if groupID != "" {
		if s.groupRepo == nil {
			return groupID, group_label, nil
		}
		g, err := s.groupRepo.FindByID(ctx, groupID)
		if err != nil {
			return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok tidak dikenal")
		}
		if group_label == "" {
			group_label = g.GroupName
		}
		return g.GroupID, group_label, nil
	}
	if group_label == "" {
		return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok wajib dipilih")
	}
	if s.groupRepo == nil {
		return "", group_label, nil
	}
	g, err := s.groupRepo.FindByName(ctx, group_label)
	if err != nil {
		return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok tidak dikenal: "+group_label)
	}
	return g.GroupID, g.GroupName, nil
}

type CreateMemberInput struct {
	GroupID                string
	FullName            string
	Nickname          string
	Gender           string
	BirthPlace            string
	BirthDate           string
	GroupLabel               string
	Village                   string
	Region                 string
	HomeAddress            string
	WhatsappNumber                   string
	IsPreacher            bool
	IsEmployed                bool
	IsMarried                bool
	Height            string
	Weight             string
	Hobby                   string
	Occupation              string
	PhotoURL                string
	MentoringStatus        string
	JoinedDate           string
	EducationLevel      string
	School                string
	Major                string
	EducationStartYear   string
	EducationEndYear string
}

func (s *MemberService) Create(ctx context.Context, in CreateMemberInput) (*model.MemberDetailDTO, error) {

	in.FullName = util.TitleCaseID(in.FullName)
	in.Nickname = util.TitleCaseID(in.Nickname)
	in.BirthPlace = util.TitleCaseID(in.BirthPlace)
	in.Village = util.TitleCaseID(in.Village)
	in.Region = util.TitleCaseID(in.Region)
	if strings.TrimSpace(in.FullName) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "nama lengkap wajib diisi")
	}

	jk := normalizeGender(in.Gender)
	var jkPtr *string
	if jk != "" {
		jkPtr = &jk
	}

	var tglLahir interface{}
	if in.BirthDate != "" {
		tglLahir = in.BirthDate
	}

	status := in.MentoringStatus
	if status == "" {
		status = "AKTIF"
	}
	tglMasuk := in.JoinedDate
	if tglMasuk == "" {
		tglMasuk = time.Now().Format("2006-01-02")
	}

	noWA := ""
	if in.WhatsappNumber != "" {
		noWA = util.NormalizePhone(in.WhatsappNumber)
	}

	groupID, group_label, err := s.resolveGroup(ctx, in.GroupID, in.GroupLabel)
	if err != nil {
		return nil, err
	}

	memberID := util.NewID("MBR")
	_, err = s.repo.Pool().Exec(ctx, `
		INSERT INTO members (
			member_id, full_name, nickname, gender,
			birth_place, birth_date, photo_url, whatsapp_number,
			home_address, village, region, group_label, group_id,
			is_preacher, is_employed, is_married, height, weight,
			hobby, occupation, mentoring_status, is_active, joined_date,
			education_level, school, major,
			education_start_year, education_end_year,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
			$14,$15,$16,$17,$18,$19,$20,$21,true,$22,
			$23,$24,$25,$26,$27,now(),now()
		)
	`,
		memberID, in.FullName, in.Nickname, jkPtr,
		in.BirthPlace, tglLahir, in.PhotoURL, noWA,
		in.HomeAddress, in.Village, in.Region, group_label, groupID,
		in.IsPreacher, in.IsEmployed, in.IsMarried, in.Height, in.Weight,
		in.Hobby, in.Occupation, status, tglMasuk,
		in.EducationLevel, in.School, in.Major,
		in.EducationStartYear, in.EducationEndYear,
	)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	dto := s.toDetailDTO(m)
	return &dto, nil
}

type UpdateMemberInput struct {
	MemberID               string
	GroupID                *string
	FullName            *string
	Nickname          *string
	Gender           *string
	BirthPlace            *string
	BirthDate           *string
	GroupLabel               *string
	Village                   *string
	Region                 *string
	HomeAddress            *string
	WhatsappNumber                   *string
	IsPreacher            *bool
	IsEmployed                *bool
	IsMarried                *bool
	Height            *string
	Weight             *string
	Hobby                   *string
	Occupation              *string
	PhotoURL                *string
	MentoringStatus        *string
	JoinedDate           *string
	LeftDate          *string
	EducationLevel      *string
	School                *string
	Major                *string
	EducationStartYear   *string
	EducationEndYear *string
}

func (s *MemberService) UpdateFull(ctx context.Context, in UpdateMemberInput) (*model.MemberDetailDTO, error) {

	if in.FullName != nil {
		v := util.TitleCaseID(*in.FullName)
		in.FullName = &v
	}
	if in.Nickname != nil {
		v := util.TitleCaseID(*in.Nickname)
		in.Nickname = &v
	}
	if in.BirthPlace != nil {
		v := util.TitleCaseID(*in.BirthPlace)
		in.BirthPlace = &v
	}
	if in.Village != nil {
		v := util.TitleCaseID(*in.Village)
		in.Village = &v
	}
	if in.Region != nil {
		v := util.TitleCaseID(*in.Region)
		in.Region = &v
	}
	if in.MemberID == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, in.MemberID); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}

	patch := map[string]interface{}{}
	addStr := func(col string, v *string) {
		if v != nil {
			patch[col] = *v
		}
	}
	addBool := func(col string, v *bool) {
		if v != nil {
			patch[col] = *v
		}
	}

	addStr("full_name", in.FullName)
	addStr("nickname", in.Nickname)
	if in.Gender != nil {
		if jk := normalizeGender(*in.Gender); jk != "" {
			patch["gender"] = jk
		}
	}
	addStr("birth_place", in.BirthPlace)
	if in.BirthDate != nil {
		if *in.BirthDate == "" {
			patch["birth_date"] = nil
		} else {
			patch["birth_date"] = *in.BirthDate
		}
	}
	if in.GroupLabel != nil || in.GroupID != nil {
		gid := ""
		if in.GroupID != nil {
			gid = *in.GroupID
		}
		knama := ""
		if in.GroupLabel != nil {
			knama = *in.GroupLabel
		}
		resolvedID, resolvedName, err := s.resolveGroup(ctx, gid, knama)
		if err != nil {
			return nil, err
		}
		patch["group_id"] = resolvedID
		patch["group_label"] = resolvedName
	}
	addStr("village", in.Village)
	addStr("region", in.Region)
	addStr("home_address", in.HomeAddress)
	if in.WhatsappNumber != nil && *in.WhatsappNumber != "" {
		patch["whatsapp_number"] = util.NormalizePhone(*in.WhatsappNumber)
	}
	addBool("is_preacher", in.IsPreacher)
	addBool("is_employed", in.IsEmployed)
	addBool("is_married", in.IsMarried)
	addStr("height", in.Height)
	addStr("weight", in.Weight)
	addStr("hobby", in.Hobby)
	addStr("occupation", in.Occupation)
	addStr("photo_url", in.PhotoURL)
	addStr("mentoring_status", in.MentoringStatus)
	if in.JoinedDate != nil {
		if *in.JoinedDate == "" {
			patch["joined_date"] = nil
		} else {
			patch["joined_date"] = *in.JoinedDate
		}
	}
	if in.LeftDate != nil {
		if *in.LeftDate == "" {
			patch["left_date"] = nil
		} else {
			patch["left_date"] = *in.LeftDate
		}
	}
	addStr("education_level", in.EducationLevel)
	addStr("school", in.School)
	addStr("major", in.Major)
	addStr("education_start_year", in.EducationStartYear)
	addStr("education_end_year", in.EducationEndYear)

	if len(patch) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "tidak ada perubahan")
	}
	if err := s.repo.Update(ctx, in.MemberID, patch); err != nil {
		return nil, err
	}
	if _, ok := patch["group_id"]; ok {
		_ = s.repo.SyncUsersGroupByMember(ctx, in.MemberID)
	}
	m, err := s.repo.FindByID(ctx, in.MemberID)
	if err != nil {
		return nil, err
	}
	dto := s.toDetailDTO(m)
	return &dto, nil
}

func (s *MemberService) Deactivate(ctx context.Context, memberID string) error {
	if memberID == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, memberID); err != nil {
		return apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}
	return s.repo.Deactivate(ctx, memberID)
}

func (s *MemberService) FindForExport(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.repo.FindAllIncludingInactive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		jk := ""
		if r.Gender != nil {
			jk = *r.Gender
		}
		tgl := ""
		if r.BirthDate != nil {
			tgl = r.BirthDate.Format("2006-01-02")
		}
		out = append(out, map[string]interface{}{
			"full_name":     r.FullName,
			"nickname":   r.Nickname,
			"gender":    jk,
			"birth_place":     r.BirthPlace,
			"birth_date":    tgl,
			"usia":             util.GetAge(r.BirthDate),
			"kategori":         util.GetMemberCategory(r.BirthDate, r.EducationLevel, r.IsMarried),
			"group_label":         r.GroupLabel,
			"village":             r.Village,
			"region":           r.Region,
			"home_address":     r.HomeAddress,
			"whatsapp_number":            r.WhatsappNumber,
			"occupation":        r.Occupation,
			"hobby":             r.Hobby,
			"mentoring_status": r.MentoringStatus,
		})
	}
	return out, nil
}

func normalizeGender(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "l", "laki-laki", "laki laki", "pria", "male":
		return "L"
	case "p", "perempuan", "wanita", "female":
		return "P"
	}
	return ""
}

var _ = repository.NewMemberRepo

func (s *MemberService) UpdateFotoURL(ctx context.Context, memberID, url string) error {
	return s.repo.UpdateFotoURL(ctx, memberID, url)
}

func (s *MemberService) GetFotoURL(ctx context.Context, memberID string) (string, error) {
	return s.repo.GetFotoURL(ctx, memberID)
}

func (s *MemberService) DeleteMember(ctx context.Context, memberID string) error {
	if memberID == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, memberID); err != nil {
		return apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}
	n, err := s.repo.CountUsersByMemberID(ctx, memberID)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperrors.Wrap(apperrors.ErrValidation, "member sudah punya akun user, hapus lewat Kelola Akun")
	}
	return s.repo.Delete(ctx, memberID)
}
