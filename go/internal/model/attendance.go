package model

import "time"

type Attendance struct {
	AttendanceID string
	MeetingID    string
	MemberID     string
	Status       string
	Catatan      string
	CreatedBy    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AttendanceDTO struct {
	AttendanceID string `json:"attendance_id"`
	MeetingID    string `json:"meeting_id"`
	MemberID     string `json:"member_id"`
	Status       string `json:"status"`
	Catatan      string `json:"catatan"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type AttendancePageDTO struct {
	Meeting    MeetingDTO            `json:"meeting"`
	Members    []AttendanceMemberDTO `json:"members"`
	Attendance []AttendanceDTO       `json:"attendance"`
}
