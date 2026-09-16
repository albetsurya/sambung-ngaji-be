package model

import "time"

type Monitoring struct {
	MonitoringID string
	MemberID     string
	Tanggal      time.Time
	Jenis        string
	Status       string
	Catatan      string
	TindakLanjut string
	CreatedBy    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MonitoringDTO struct {
	MonitoringID string `json:"monitoring_id"`
	MemberID     string `json:"member_id"`
	Tanggal      string `json:"tanggal"`
	Jenis        string `json:"jenis"`
	Status       string `json:"status"`
	Catatan      string `json:"catatan"`
	TindakLanjut string `json:"tindak_lanjut"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
