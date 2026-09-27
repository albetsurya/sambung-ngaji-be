package model

import "time"

type FridaySchedule struct {
	FridayID      string
	GroupID       *string
	Tanggal       time.Time
	KhatibImam    string
	Muadzin       string
	Penasihat     string
	PetugasParkir string
	PenataSandal  string
	Catatan       string
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type FridayScheduleDTO struct {
	FridayID      string `json:"friday_id"`
	GroupID       string `json:"group_id"`
	Tanggal       string `json:"tanggal"`
	Hari          string `json:"hari"`
	KhatibImam    string `json:"khatib_imam"`
	Muadzin       string `json:"muadzin"`
	Penasihat     string `json:"penasihat"`
	PetugasParkir string `json:"petugas_parkir"`
	PenataSandal  string `json:"penata_sandal"`
	Catatan       string `json:"catatan"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
