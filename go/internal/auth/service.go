package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUserInactive       = errors.New("akun tidak aktif")
)

type Service struct {
	pool *pgxpool.Pool
	jwt  *JWTManager
}

func NewService(pool *pgxpool.Pool, jwt *JWTManager) *Service {
	return &Service{pool: pool, jwt: jwt}
}

// Login: cari user, verifikasi password (SHA-256 legacy atau bcrypt),
// auto-upgrade ke bcrypt kalau masih SHA-256, buat session + JWT.
func (s *Service) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	u, err := s.findByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}
	if !u.StatusAktif {
		return "", nil, ErrUserInactive
	}

	match, needsRehash := VerifyPassword(password, u.PasswordHash)
	if !match {
		return "", nil, ErrInvalidCredentials
	}

	if needsRehash {
		newHash, err := HashPassword(password)
		if err == nil {
			_, _ = s.pool.Exec(ctx,
				`UPDATE users SET password_hash=$1, updated_at=now() WHERE user_id=$2`,
				newHash, u.UserID)
			u.PasswordHash = newHash
		}
	}

	sessionID := uuid.NewString()
	expiresAt := time.Now().Add(s.jwt.TTL())

	_, err = s.pool.Exec(ctx, `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES ($1, $2, now(), $3)
	`, sessionID, u.UserID, expiresAt)
	if err != nil {
		return "", nil, fmt.Errorf("simpan session: %w", err)
	}

	_, _ = s.pool.Exec(ctx,
		`UPDATE users SET last_login_at = now() WHERE user_id = $1`, u.UserID)

	token, err := s.jwt.Generate(sessionID, u.UserID, u.Role, ptrToString(u.GroupID))
	if err != nil {
		return "", nil, fmt.Errorf("generate jwt: %w", err)
	}

	return token, u, nil
}

// Logout: hapus session (revoke).
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, sessionID)
	return err
}

// ValidateSession: cek JWT lalu cek session masih ada di DB.
func (s *Service) ValidateSession(ctx context.Context, token string) (*model.User, *model.SessionClaims, error) {
	claims, err := s.jwt.Parse(token)
	if err != nil {
		return nil, nil, err
	}

	var expiresAt time.Time
	err = s.pool.QueryRow(ctx,
		`SELECT expires_at FROM sessions WHERE token = $1`, claims.SessionID,
	).Scan(&expiresAt)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	if time.Now().After(expiresAt) {
		_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, claims.SessionID)
		return nil, nil, ErrInvalidToken
	}

	u, err := s.findByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	if !u.StatusAktif {
		return nil, nil, ErrInvalidToken
	}
	return u, claims, nil
}

// InvalidateUserSessions: hapus semua session user (untuk change password,
// ubah role, atau ubah status aktif).
func (s *Service) InvalidateUserSessions(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (s *Service) findByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	var gid *string
	err := s.pool.QueryRow(ctx, `
		SELECT user_id, username, password_hash, nama, role, group_id, member_id,
		       status_aktif, created_at, updated_at, last_login_at
		FROM users WHERE username = $1
	`, username).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&gid, &u.MemberID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	u.GroupID = gid
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) findByID(ctx context.Context, userID string) (*model.User, error) {
	var u model.User
	var gid *string
	err := s.pool.QueryRow(ctx, `
		SELECT user_id, username, password_hash, nama, role, group_id, member_id,
		       status_aktif, created_at, updated_at, last_login_at
		FROM users WHERE user_id = $1
	`, userID).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role,
		&gid, &u.MemberID, &u.StatusAktif, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	u.GroupID = gid
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ToPublic mengubah User ke bentuk response frontend.
// jenis_kelamin diambil dari members kalau member_id ada.
func (s *Service) ToPublic(ctx context.Context, u *model.User) model.PublicUser {
	p := model.PublicUser{
		UserID:   u.UserID,
		Username: u.Username,
		Nama:     u.Nama,
		Role:     u.Role,
		GroupID:  ptrToString(u.GroupID),
	}
	if u.MemberID != nil && *u.MemberID != "" {
		p.MemberID = *u.MemberID
		var jk, foto *string
		_ = s.pool.QueryRow(ctx,
			`SELECT jenis_kelamin, foto_url FROM members WHERE member_id = $1`, *u.MemberID,
		).Scan(&jk, &foto)
		if jk != nil {
			p.JenisKelamin = *jk
		}
		if foto != nil {
			p.FotoURL = *foto
		}
	}
	return p
}
