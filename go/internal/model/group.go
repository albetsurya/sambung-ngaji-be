package model

import "time"

type Group struct {
	GroupID       string
	GroupCode     string
	GroupName     string
	Pembina       string
	Penandatangan string
	Jadwal        string
	StatusAktif   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GroupDTO struct {
	GroupID       string `json:"group_id"`
	GroupCode     string `json:"group_code"`
	GroupName     string `json:"group_name"`
	Pembina       string `json:"pembina"`
	Penandatangan string `json:"penandatangan"`
	Jadwal        string `json:"jadwal"`
	StatusAktif   bool   `json:"status_aktif"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
