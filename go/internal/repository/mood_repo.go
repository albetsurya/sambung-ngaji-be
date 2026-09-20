package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MoodRepo struct {
	pool *pgxpool.Pool
}

func NewMoodRepo(pool *pgxpool.Pool) *MoodRepo {
	return &MoodRepo{pool: pool}
}

const moodSelectCols = `mood_id, member_id, mood_key, tanggal, created_at`

func (r *MoodRepo) FindByMember(ctx context.Context, memberID string, limit int) ([]model.MemberMood, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+moodSelectCols+` FROM member_moods WHERE member_id=$1 ORDER BY tanggal DESC LIMIT $2`,
		memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMoods(rows)
}

func (r *MoodRepo) FindByMemberAndDate(ctx context.Context, memberID, tanggal string) (*model.MemberMood, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+moodSelectCols+` FROM member_moods WHERE member_id=$1 AND tanggal=$2`,
		memberID, tanggal)
	var m model.MemberMood
	if err := row.Scan(&m.MoodID, &m.MemberID, &m.MoodKey, &m.Tanggal, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MoodRepo) Upsert(ctx context.Context, m *model.MemberMood) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO member_moods (mood_id, member_id, mood_key, tanggal, created_at)
		 VALUES ($1,$2,$3,$4, now())
		 ON CONFLICT (member_id, tanggal)
		 DO UPDATE SET mood_key = EXCLUDED.mood_key`,
		m.MoodID, m.MemberID, m.MoodKey, m.Tanggal)
	return err
}

func scanMoods(rows rowsScanner) ([]model.MemberMood, error) {
	var out []model.MemberMood
	for rows.Next() {
		var m model.MemberMood
		if err := rows.Scan(&m.MoodID, &m.MemberID, &m.MoodKey, &m.Tanggal, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
