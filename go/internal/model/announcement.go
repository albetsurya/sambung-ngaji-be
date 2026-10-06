package model

import "time"

type AnnouncementTemplate struct {
	TemplateID   string
	GroupID      *string
	TemplateName string
	Kode         string
	TemplateBody string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AnnouncementTemplateDTO struct {
	TemplateID   string `json:"template_id"`
	GroupID      string `json:"group_id"`
	TemplateName string `json:"template_name"`
	Kode         string `json:"kode"`
	TemplateBody string `json:"template_body"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Announcement struct {
	AnnouncementID string
	TemplateID     *string
	MeetingID      *string
	GroupID        *string
	Date           time.Time
	Day            string
	Time           string
	Event          string
	Topic          string
	Notes          string
	GeneratedText  string
	Status         string
	CreatedBy      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AnnouncementDTO struct {
	AnnouncementID string `json:"announcement_id"`
	TemplateID     string `json:"template_id"`
	MeetingID      string `json:"meeting_id"`
	GroupID        string `json:"group_id"`
	Date           string `json:"date"`
	Day            string `json:"day"`
	Time           string `json:"time"`
	Event          string `json:"event"`
	Topic          string `json:"topic"`
	Notes          string `json:"notes"`
	GeneratedText  string `json:"generated_text"`
	Status         string `json:"status"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
