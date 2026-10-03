package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MemberRepo struct {
	pool *pgxpool.Pool
}

func NewMemberRepo(pool *pgxpool.Pool) *MemberRepo {
	return &MemberRepo{pool: pool}
}

const memberSelectCols = `
	member_id, group_id, full_name, nickname, gender,
	birth_place, birth_date, photo_url, whatsapp_number,
	home_address, village, region, group_label,
	is_preacher, is_employed, is_married, height, weight,
	hobby, occupation, mentoring_status, is_active,
	joined_date, left_date,
	education_level, school, major,
	education_start_year, education_end_year,
	created_at, updated_at`

func (r *MemberRepo) FindAll(ctx context.Context) ([]model.Member, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	rows, err := r.pool.Query(ctx,
		`SELECT `+memberSelectCols+` FROM members WHERE is_active = true ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMembers(rows)
}

func (r *MemberRepo) FindAllByGroup(ctx context.Context, groupID string) ([]model.Member, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	if groupID == "" {
		return r.FindAll(ctx)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+memberSelectCols+` FROM members WHERE is_active = true AND group_id = $1 ORDER BY full_name`,
		groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMembers(rows)
}

func (r *MemberRepo) FindByID(ctx context.Context, id string) (*model.Member, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	row := r.pool.QueryRow(ctx,
		`SELECT `+memberSelectCols+` FROM members WHERE member_id = $1`, id)
	return scanMember(row)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMember(s rowScanner) (*model.Member, error) {
	var m model.Member
	err := s.Scan(
		&m.MemberID, &m.GroupID, &m.FullName, &m.Nickname, &m.Gender,
		&m.BirthPlace, &m.BirthDate, &m.PhotoURL, &m.WhatsappNumber,
		&m.HomeAddress, &m.Village, &m.Region, &m.GroupLabel,
		&m.IsPreacher, &m.IsEmployed, &m.IsMarried, &m.Height, &m.Weight,
		&m.Hobby, &m.Occupation, &m.MentoringStatus, &m.IsActive,
		&m.JoinedDate, &m.LeftDate,
		&m.EducationLevel, &m.School, &m.Major,
		&m.EducationStartYear, &m.EducationEndYear,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type rowsScanner interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

func scanMembers(rows rowsScanner) ([]model.Member, error) {
	var out []model.Member
	for rows.Next() {
		var m model.Member
		err := rows.Scan(
			&m.MemberID, &m.GroupID, &m.FullName, &m.Nickname, &m.Gender,
			&m.BirthPlace, &m.BirthDate, &m.PhotoURL, &m.WhatsappNumber,
			&m.HomeAddress, &m.Village, &m.Region, &m.GroupLabel,
			&m.IsPreacher, &m.IsEmployed, &m.IsMarried, &m.Height, &m.Weight,
			&m.Hobby, &m.Occupation, &m.MentoringStatus, &m.IsActive,
			&m.JoinedDate, &m.LeftDate,
			&m.EducationLevel, &m.School, &m.Major,
			&m.EducationStartYear, &m.EducationEndYear,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi rows: %w", err)
	}
	return out, nil
}
