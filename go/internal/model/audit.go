package model

import "time"

type AuditLog struct {
	LogID      string
	UserID     *string
	UserNama   string
	Action     string
	TargetType string
	TargetID   string
	Timestamp  time.Time
}

type AuditLogDTO struct {
	LogID      string `json:"log_id"`
	UserID     string `json:"user_id"`
	UserNama   string `json:"user_nama"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Timestamp  string `json:"timestamp"`
}

type UserDTO struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Nama        string `json:"nama"`
	Role        string `json:"role"`
	GroupID     string `json:"group_id"`
	MemberID    string `json:"member_id"`
	StatusAktif bool   `json:"status_aktif"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastLoginAt string `json:"last_login_at"`
}
