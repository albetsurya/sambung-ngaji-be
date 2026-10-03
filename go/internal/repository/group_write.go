package repository

import (
	"context"

	"pengajian-backend/internal/model"
)

func (r *GroupRepo) Insert(ctx context.Context, g *model.Group) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO groups
		(group_id, group_code, group_name, mentor, signatory,
		 schedule, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now())
	`, g.GroupID, g.GroupCode, g.GroupName, g.Mentor, g.Signatory,
		g.Schedule, g.IsActive)
	return err
}

func (r *GroupRepo) Update(ctx context.Context, id string, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	q := `UPDATE groups SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE group_id = $` + itoa(n)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
