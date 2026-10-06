package model

import "time"

type AuditLog struct {
	LogID      string
	UserID     *string
	UserName   string
	Action     string
	TargetType string
	TargetID   string
	Timestamp  time.Time
}

type AuditLogDTO struct {
	LogID      string `json:"log_id"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Timestamp  string `json:"timestamp"`
}

type UserDTO struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	GroupID     string `json:"group_id"`
	MemberID    string `json:"member_id"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastLoginAt string `json:"last_login_at"`
}
