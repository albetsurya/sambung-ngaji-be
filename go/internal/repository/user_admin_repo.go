package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type UserAdminRepo struct {
	pool *pgxpool.Pool
}

func NewUserAdminRepo(pool *pgxpool.Pool) *UserAdminRepo {
	return &UserAdminRepo{pool: pool}
}

const userAdminSelectCols = `
	u.user_id, u.username, u.password_hash, u.nama, u.role, u.member_id,
	COALESCE(m.group_id, u.group_id, '') AS group_id,
	u.status_aktif, u.created_at, u.updated_at, u.last_login_at`

func (r *UserAdminRepo) FindAll(ctx context.Context) ([]model.User, error) {
	q := `SELECT ` + userAdminSelectCols + ` FROM users u LEFT JOIN members m ON u.member_id = m.member_id ORDER BY u.username`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		var u model.User
		var groupID string
		if err := rows.Scan(
			&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
			&u.MemberID, &groupID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
		); err != nil {
			return nil, err
		}
		if groupID != "" {
			g := groupID
			u.GroupID = &g
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserAdminRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	var groupID string
	q := `SELECT ` + userAdminSelectCols + ` FROM users u LEFT JOIN members m ON u.member_id = m.member_id WHERE u.user_id = $1`
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&u.MemberID, &groupID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	if groupID != "" {
		g := groupID
		u.GroupID = &g
	}
	return &u, nil
}

func (r *UserAdminRepo) Update(ctx context.Context, id string, patch map[string]interface{}) error {
	q := `UPDATE users SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE user_id = $` + itoa(n)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *UserAdminRepo) CountActiveSuperAdminsExcept(ctx context.Context, exceptID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE role = 'SUPER_ADMIN' AND status_aktif = true AND user_id <> $1
	`, exceptID).Scan(&n)
	return n, err
}

func (r *UserAdminRepo) FindActiveUserByMemberID(ctx context.Context, memberID string) (*model.User, error) {
	var u model.User
	var groupID string
	q := `SELECT ` + userAdminSelectCols + ` FROM users u LEFT JOIN members m ON u.member_id = m.member_id WHERE u.member_id = $1 AND u.status_aktif = true LIMIT 1`
	err := r.pool.QueryRow(ctx, q, memberID).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&u.MemberID, &groupID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	if groupID != "" {
		g := groupID
		u.GroupID = &g
	}
	return &u, nil
}

func (r *UserAdminRepo) DeleteAllSessions(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (r *UserAdminRepo) DeleteAllSessionsExcept(ctx context.Context, userID, keepToken string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1 AND token <> $2`, userID, keepToken)
	return err
}

func (r *UserAdminRepo) UpdatePasswordHash(ctx context.Context, userID, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = now() WHERE user_id = $2
	`, hash, userID)
	return err
}

func (r *UserAdminRepo) DeleteUserAndMember(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user_id wajib diisi")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var role string
	var memberID *string
	err = tx.QueryRow(ctx,
		`SELECT role, member_id FROM users WHERE user_id = $1`, userID,
	).Scan(&role, &memberID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}
	if role == "SUPER_ADMIN" {
		return errors.New("tidak bisa menghapus user SUPER_ADMIN")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, userID); err != nil {
		return err
	}

	if memberID != nil && *memberID != "" {
		if _, err := tx.Exec(ctx, `DELETE FROM members WHERE member_id = $1`, *memberID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
