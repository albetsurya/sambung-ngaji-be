package model

import "time"

type MemberRequest struct {
	RequestID  string
	UserID     string
	Nama       string
	Status     string
	MemberID   *string
	Reason     string
	CreatedAt  time.Time
	ReviewedBy *string
	ReviewedAt *time.Time
}

type MemberRequestDTO struct {
	RequestID  string `json:"request_id"`
	UserID     string `json:"user_id"`
	Nama       string `json:"nama"`
	Status     string `json:"status"`
	MemberID   string `json:"member_id"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
	ReviewedBy string `json:"reviewed_by"`
	ReviewedAt string `json:"reviewed_at"`
}