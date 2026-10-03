package model

import "time"

type Group struct {
	GroupID       string
	GroupCode     string
	GroupName     string
	Mentor       string
	Signatory string
	Schedule        string
	IsActive   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GroupDTO struct {
	GroupID       string `json:"group_id"`
	GroupCode     string `json:"group_code"`
	GroupName     string `json:"group_name"`
	Mentor       string `json:"mentor"`
	Signatory string `json:"signatory"`
	Schedule        string `json:"schedule"`
	IsActive   bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
