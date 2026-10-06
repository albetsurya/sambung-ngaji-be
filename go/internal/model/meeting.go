package model

import "time"

type Meeting struct {
	MeetingID        string
	Date             time.Time
	Day              string
	Time             string
	StartTime        string
	GroupID          *string
	Event            string
	Topic            string
	Status           string
	Notes            string
	TargetCategories []string
	GenderTarget     *string
	CreatedBy        *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MeetingDTO struct {
	MeetingID        string   `json:"meeting_id"`
	Date             string   `json:"date"`
	Day              string   `json:"day"`
	Time             string   `json:"time"`
	StartTime        string   `json:"start_time"`
	GroupID          string   `json:"group_id"`
	Event            string   `json:"event"`
	Topic            string   `json:"topic"`
	Status           string   `json:"status"`
	Notes            string   `json:"notes"`
	TargetCategories []string `json:"target_categories"`
	GenderTarget     string   `json:"gender_target"`
	CreatedBy        string   `json:"created_by"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type MeetingListFilter struct {
	From    string
	To      string
	GroupID string
}
