package repository

import (
	"context"
)

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

// SyncUsersGroupByMember menyamakan users.group_id dengan members.group_id
// untuk semua akun yang terhubung ke member tersebut. User tanpa member_id
// (NULL) tidak tersentuh. Dipakai sebagai jaring pengaman aplikasi selain
// trigger DB trg_sync_user_group.
func (r *MemberRepo) SyncUsersGroupByMember(ctx context.Context, memberID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users u SET group_id = m.group_id, updated_at = NOW()
		FROM members m
		WHERE u.member_id = m.member_id
		  AND m.member_id = $1
		  AND m.group_id IS NOT NULL
		  AND (u.group_id IS DISTINCT FROM m.group_id)`, memberID)
	return err
}
