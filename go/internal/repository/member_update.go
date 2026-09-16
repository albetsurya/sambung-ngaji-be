package repository

import (
	"context"
)

// Update: patch kolom members sesuai map (whitelist di service).
// Key harus nama kolom Postgres yang valid.
func (r *MemberRepo) Update(ctx context.Context, id string, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	q := `UPDATE members SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE member_id = $` + itoa(n)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
