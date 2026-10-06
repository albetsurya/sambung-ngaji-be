package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberExportRow struct {
	MemberID        string
	FullName        string
	Nickname        string
	Gender          *string
	BirthPlace      string
	BirthDate       *time.Time
	Village         string
	Region          string
	HomeAddress     string
	WhatsappNumber  string
	Occupation      string
	Hobby           string
	MentoringStatus string
	GroupLabel      string
	EducationLevel  string
	IsMarried       bool
	IsActive        bool
}

func (r *MemberRepo) Deactivate(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE members
		SET is_active = false, left_date = $1, updated_at = now()
		WHERE member_id = $2
	`, time.Now().Format("2006-01-02"), id)
	return err
}

func (r *MemberRepo) FindAllIncludingInactive(ctx context.Context) ([]MemberExportRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT member_id, full_name, nickname, gender,
		       birth_place, birth_date, village, region, home_address,
		       whatsapp_number, occupation, hobby, mentoring_status, group_label,
		       education_level, is_married, is_active
		FROM members
		ORDER BY full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MemberExportRow
	for rows.Next() {
		var m MemberExportRow
		if err := rows.Scan(
			&m.MemberID, &m.FullName, &m.Nickname, &m.Gender,
			&m.BirthPlace, &m.BirthDate, &m.Village, &m.Region, &m.HomeAddress,
			&m.WhatsappNumber, &m.Occupation, &m.Hobby, &m.MentoringStatus, &m.GroupLabel,
			&m.EducationLevel, &m.IsMarried, &m.IsActive,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRepo) Pool() *pgxpool.Pool { return r.pool }
func (r *MemberRepo) UpdateFotoURL(ctx context.Context, memberID, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE members SET photo_url = $1, updated_at = now() WHERE member_id = $2
	`, url, memberID)
	return err
}

func (r *MemberRepo) GetFotoURL(ctx context.Context, memberID string) (string, error) {
	var url string
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(photo_url, '') FROM members WHERE member_id = $1`, memberID,
	).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (r *MemberRepo) MemberIDsWithUsers(ctx context.Context) (map[string]bool, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT member_id FROM users
		WHERE member_id IS NOT NULL AND member_id <> '' AND is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (r *MemberRepo) CountUsersByMemberID(ctx context.Context, memberID string) (int, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE member_id = $1 AND is_active = true`, memberID).Scan(&n)
	return n, err
}

func (r *MemberRepo) MemberHasActiveUser(ctx context.Context, memberID string) (bool, error) {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE member_id = $1 AND is_active = true)`, memberID).Scan(&exists)
	return exists, err
}

func (r *MemberRepo) Delete(ctx context.Context, id string) error {
	ctx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	_, err := r.pool.Exec(ctx, `DELETE FROM members WHERE member_id = $1`, id)
	return err
}
