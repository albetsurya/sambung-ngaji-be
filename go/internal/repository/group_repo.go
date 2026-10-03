package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type GroupRepo struct {
	pool *pgxpool.Pool
}

func NewGroupRepo(pool *pgxpool.Pool) *GroupRepo {
	return &GroupRepo{pool: pool}
}

const groupSelectCols = `
	group_id, group_code, group_name, mentor, signatory,
	schedule, is_active, created_at, updated_at`

func (r *GroupRepo) FindAll(ctx context.Context, includeInactive bool) ([]model.Group, error) {
	q := `SELECT ` + groupSelectCols + ` FROM groups`
	if !includeInactive {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY group_name`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Group
	for rows.Next() {
		var g model.Group
		if err := rows.Scan(
			&g.GroupID, &g.GroupCode, &g.GroupName, &g.Mentor, &g.Signatory,
			&g.Schedule, &g.IsActive, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *GroupRepo) FindByID(ctx context.Context, id string) (*model.Group, error) {
	var g model.Group
	err := r.pool.QueryRow(ctx,
		`SELECT `+groupSelectCols+` FROM groups WHERE group_id = $1`, id,
	).Scan(
		&g.GroupID, &g.GroupCode, &g.GroupName, &g.Mentor, &g.Signatory,
		&g.Schedule, &g.IsActive, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupRepo) FindByName(ctx context.Context, name string) (*model.Group, error) {
	var g model.Group
	err := r.pool.QueryRow(ctx,
		`SELECT `+groupSelectCols+` FROM groups WHERE LOWER(TRIM(group_name)) = LOWER(TRIM($1)) ORDER BY is_active DESC LIMIT 1`, name,
	).Scan(
		&g.GroupID, &g.GroupCode, &g.GroupName, &g.Mentor, &g.Signatory,
		&g.Schedule, &g.IsActive, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
