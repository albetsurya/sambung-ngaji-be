package model

import "time"

type Member struct {
	MemberID           string
	GroupID            *string
	FullName           string
	Nickname           string
	Gender             *string
	BirthPlace         string
	BirthDate          *time.Time
	PhotoURL           string
	WhatsappNumber     string
	HomeAddress        string
	Village            string
	Region             string
	GroupLabel         string
	IsPreacher         bool
	IsEmployed         bool
	IsMarried          bool
	Height             string
	Weight             string
	Hobby              string
	Occupation         string
	MentoringStatus    string
	IsActive           bool
	JoinedDate         *time.Time
	LeftDate           *time.Time
	EducationLevel     string
	School             string
	Major              string
	EducationStartYear string
	EducationEndYear   string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type MemberListDTO struct {
	MemberID   string `json:"member_id"`
	GroupID    string `json:"group_id"`
	FullName   string `json:"full_name"`
	Nickname   string `json:"nickname"`
	Gender     string `json:"gender"`
	GroupLabel string `json:"group_label"`
	GroupName  string `json:"group_name"`
	Kategori   string `json:"kategori"`
	PhotoURL   string `json:"photo_url"`
	HasUser    bool   `json:"has_user"`
}

type AttendanceMemberDTO struct {
	MemberID   string `json:"member_id"`
	GroupID    string `json:"group_id"`
	FullName   string `json:"full_name"`
	GroupLabel string `json:"group_label"`
	Kategori   string `json:"kategori"`
	Gender     string `json:"gender"`
}

type MemberDetailDTO struct {
	MemberID           string `json:"member_id"`
	GroupID            string `json:"group_id"`
	FullName           string `json:"full_name"`
	Nickname           string `json:"nickname"`
	Gender             string `json:"gender"`
	BirthPlace         string `json:"birth_place"`
	BirthDate          string `json:"birth_date"`
	PhotoURL           string `json:"photo_url"`
	WhatsappNumber     string `json:"whatsapp_number"`
	HomeAddress        string `json:"home_address"`
	Village            string `json:"village"`
	Region             string `json:"region"`
	GroupLabel         string `json:"group_label"`
	GroupName          string `json:"group_name"`
	IsPreacher         bool   `json:"is_preacher"`
	IsEmployed         bool   `json:"is_employed"`
	IsMarried          bool   `json:"is_married"`
	Height             string `json:"height"`
	Weight             string `json:"weight"`
	Hobby              string `json:"hobby"`
	Occupation         string `json:"occupation"`
	MentoringStatus    string `json:"mentoring_status"`
	IsActive           bool   `json:"is_active"`
	JoinedDate         string `json:"joined_date"`
	LeftDate           string `json:"left_date"`
	EducationLevel     string `json:"education_level"`
	School             string `json:"school"`
	Major              string `json:"major"`
	EducationStartYear string `json:"education_start_year"`
	EducationEndYear   string `json:"education_end_year"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	Kategori           string `json:"kategori"`
	Usia               int    `json:"usia"`
	Pendidikan         []any  `json:"pendidikan"`
	HasUser            bool   `json:"has_user"`
}

type MemberListFilter struct {
	Search          string
	GroupLabel      string
	Gender          string
	Village         string
	Kategori        string
	IncludeInactive bool
	Limit           int
	Offset          int
}
