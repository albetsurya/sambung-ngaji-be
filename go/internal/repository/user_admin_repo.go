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
	user_id, username, password_hash, nama, role, member_id,
	status_aktif, created_at, updated_at, last_login_at`

func (r *UserAdminRepo) FindAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userAdminSelectCols+` FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
			&u.MemberID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
		); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserAdminRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT `+userAdminSelectCols+` FROM users WHERE user_id = $1`, id,
	).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&u.MemberID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
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
	err := r.pool.QueryRow(ctx,
		`SELECT `+userAdminSelectCols+` FROM users WHERE member_id = $1 AND status_aktif = true LIMIT 1`,
		memberID,
	).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&u.MemberID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserAdminRepo) DeleteAllSessions(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// DeleteAllSessionsExcept — hapus semua sesi user kecuali sesi berjalan (by token).
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

// DeleteUserAndMember — hapus user + member permanen dalam transaction.
// Tolak jika user ber-role SUPER_ADMIN.
// Cascade otomatis: sessions, member_requests (via users),
// attendance, monitoring, member_moods (via members).
func (r *UserAdminRepo) DeleteUserAndMember(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user_id wajib diisi")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Cek role + member_id dulu
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

	// Hapus user (cascade ke sessions, member_requests)
	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, userID); err != nil {
		return err
	}

	// Hapus member (cascade ke attendance, monitoring, member_moods)
	if memberID != nil && *memberID != "" {
		if _, err := tx.Exec(ctx, `DELETE FROM members WHERE member_id = $1`, *memberID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
