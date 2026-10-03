package model

import "time"

type FridaySchedule struct {
	FridayID      string
	GroupID       *string
	Date       time.Time
	SermonLeader    string
	Muadzin       string
	Advisor     string
	ParkingAttendant string
	FootwearAttendant  string
	Notes       string
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type FridayScheduleDTO struct {
	FridayID      string `json:"friday_id"`
	GroupID       string `json:"group_id"`
	Date       string `json:"date"`
	Day          string `json:"day"`
	SermonLeader    string `json:"sermon_leader"`
	Muadzin       string `json:"muadzin"`
	Advisor     string `json:"advisor"`
	ParkingAttendant string `json:"parking_attendant"`
	FootwearAttendant  string `json:"footwear_attendant"`
	Notes       string `json:"notes"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
