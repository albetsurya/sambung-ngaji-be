package model

import "time"

type User struct {
	UserID       string
	Username     string
	PasswordHash string
	Nama         string
	Role         string
	GroupID      *string
	MemberID     *string
	StatusAktif  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

type PublicUser struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Nama         string `json:"nama"`
	Role         string `json:"role"`
	GroupID      string `json:"group_id"`
	MemberID     string `json:"member_id"`
	JenisKelamin string `json:"jenis_kelamin"`
	FotoURL      string `json:"foto_url"`
}

type SessionClaims struct {
	SessionID string  `json:"sid"`
	UserID    string  `json:"uid"`
	Role      string  `json:"role"`
	GroupID   *string `json:"gid,omitempty"`
}
