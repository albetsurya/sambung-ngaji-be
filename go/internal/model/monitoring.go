package model

import "time"

type Monitoring struct {
	MonitoringID string
	MemberID     string
	Date      time.Time
	Type        string
	Status       string
	Notes      string
	FollowUp string
	CreatedBy    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MonitoringDTO struct {
	MonitoringID string `json:"monitoring_id"`
	MemberID     string `json:"member_id"`
	Date      string `json:"date"`
	Type        string `json:"type"`
	Status       string `json:"status"`
	Notes      string `json:"notes"`
	FollowUp string `json:"follow_up"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
