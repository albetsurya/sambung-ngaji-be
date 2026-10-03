package model

import "time"

type PendingMember struct {
	SubmissionID           string
	GroupID                *string
	FullName            string
	Nickname          string
	Gender           *string
	BirthPlace            string
	BirthDate           *time.Time
	WhatsappNumber                   string
	HomeAddress            string
	Village                   string
	Region                 string
	Occupation              string
	Hobby                   string
	IsMarried                bool
	EducationLevel      string
	School                string
	Major                string
	EducationStartYear   string
	EducationEndYear string
	PhotoURL                string
	Username               string
	PasswordHash           string
	Status                 string
	SubmittedAt            time.Time
	SubmittedIP            string
	ReviewedBy             *string
	ReviewedAt             *time.Time
	RejectionReason        string
	CreatedMemberID        *string
}

type PendingMemberDTO struct {
	SubmissionID           string `json:"submission_id"`
	GroupID                string `json:"group_id"`
	FullName            string `json:"full_name"`
	Nickname          string `json:"nickname"`
	Gender           string `json:"gender"`
	BirthPlace            string `json:"birth_place"`
	BirthDate           string `json:"birth_date"`
	WhatsappNumber                   string `json:"whatsapp_number"`
	HomeAddress            string `json:"home_address"`
	Village                   string `json:"village"`
	Region                 string `json:"region"`
	Occupation              string `json:"occupation"`
	Hobby                   string `json:"hobby"`
	IsMarried                bool   `json:"is_married"`
	EducationLevel      string `json:"education_level"`
	School                string `json:"school"`
	Major                string `json:"major"`
	EducationStartYear   string `json:"education_start_year"`
	EducationEndYear string `json:"education_end_year"`
	PhotoURL                string `json:"photo_url"`
	Username               string `json:"username"`
	Status                 string `json:"status"`
	SubmittedAt            string `json:"submitted_at"`
	SubmittedIP            string `json:"submitted_ip"`
	ReviewedBy             string `json:"reviewed_by"`
	ReviewedAt             string `json:"reviewed_at"`
	RejectionReason        string `json:"rejection_reason"`
	CreatedMemberID        string `json:"created_member_id"`
}
