package model

import "time"

type User struct {
	UserID       string
	Username     string
	PasswordHash string
	Name         string
	Role         string
	GroupID      *string
	MemberID     *string
	IsActive  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

type PublicUser struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	GroupID      string `json:"group_id"`
	MemberID     string `json:"member_id"`
	Gender string `json:"gender"`
	PhotoURL      string `json:"photo_url"`
}

type SessionClaims struct {
	SessionID string  `json:"sid"`
	UserID    string  `json:"uid"`
	Role      string  `json:"role"`
	GroupID   *string `json:"gid,omitempty"`
}
