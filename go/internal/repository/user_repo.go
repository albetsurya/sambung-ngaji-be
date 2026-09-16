package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepo) MemberHasActiveUser(ctx context.Context, memberID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE member_id = $1 AND status_aktif = true)`, memberID,
	).Scan(&exists)
	return exists, err
}

type NewUserInput struct {
	UserID       string
	Username     string
	PasswordHash string
	Nama         string
	Role         string
	MemberID     string
}

func (r *UserRepo) Insert(ctx context.Context, in NewUserInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users
		(user_id, username, password_hash, nama, role, member_id, status_aktif, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,true,now(),now())
	`, in.UserID, in.Username, in.PasswordHash, in.Nama, in.Role, in.MemberID)
	return err
}
