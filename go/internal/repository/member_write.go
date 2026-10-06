package repository

import (
	"context"

	"pengajian-backend/internal/model"
)

type NewMemberInput struct {
	MemberID           string
	GroupID            *string
	FullName           string
	Nickname           string
	Gender             *string
	BirthPlace         string
	BirthDate          interface{}
	PhotoURL           string
	WhatsappNumber     string
	HomeAddress        string
	Village            string
	Region             string
	GroupLabel         string
	IsPreacher         bool
	IsEmployed         bool
	IsMarried          bool
	Hobby              string
	Occupation         string
	MentoringStatus    string
	EducationLevel     string
	School             string
	Major              string
	EducationStartYear string
	EducationEndYear   string
}

func (r *MemberRepo) Insert(ctx context.Context, in NewMemberInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO members (
			member_id, group_id, full_name, nickname, gender,
			birth_place, birth_date, photo_url, whatsapp_number,
			home_address, village, region, group_label,
			is_preacher, is_employed, is_married,
			hobby, occupation, mentoring_status, is_active,
			joined_date, education_level, school, major,
			education_start_year, education_end_year,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,true,
			now(),$19,$20,$21,$22,$23,now(),now()
		)
	`, in.MemberID, in.GroupID, in.FullName, in.Nickname, in.Gender,
		in.BirthPlace, in.BirthDate, in.PhotoURL, in.WhatsappNumber,
		in.HomeAddress, in.Village, in.Region, in.GroupLabel,
		in.IsPreacher, in.IsEmployed, in.IsMarried,
		in.Hobby, in.Occupation, in.MentoringStatus,
		in.EducationLevel, in.School, in.Major,
		in.EducationStartYear, in.EducationEndYear)
	return err
}

var _ = model.Member{}
