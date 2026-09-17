package model

import "time"

type Meeting struct {
	MeetingID      string
	Tanggal        time.Time
	Hari           string
	Jam            string
	JamStart       string
	GroupID        *string
	Acara          string
	Materi         string
	Status         string
	Catatan        string
	KategoriTarget []string
	GenderTarget   *string // ← BARU
	CreatedBy      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MeetingDTO struct {
	MeetingID      string   `json:"meeting_id"`
	Tanggal        string   `json:"tanggal"`
	Hari           string   `json:"hari"`
	Jam            string   `json:"jam"`
	JamStart       string   `json:"jam_start"`
	GroupID        string   `json:"group_id"`
	Acara          string   `json:"acara"`
	Materi         string   `json:"materi"`
	Status         string   `json:"status"`
	Catatan        string   `json:"catatan"`
	KategoriTarget []string `json:"kategori_target"`
	GenderTarget   string   `json:"gender_target"` // ← BARU
	CreatedBy      string   `json:"created_by"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type MeetingListFilter struct {
	From    string
	To      string
	GroupID string
}
