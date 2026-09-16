package model

import "time"

type AnnouncementTemplate struct {
	TemplateID   string
	NamaTemplate string
	Kode         string
	IsiTemplate  string
	StatusAktif  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AnnouncementTemplateDTO struct {
	TemplateID   string `json:"template_id"`
	NamaTemplate string `json:"nama_template"`
	Kode         string `json:"kode"`
	IsiTemplate  string `json:"isi_template"`
	StatusAktif  bool   `json:"status_aktif"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Announcement struct {
	AnnouncementID string
	TemplateID     *string
	MeetingID      *string
	GroupID        *string
	Tanggal        time.Time
	Hari           string
	Jam            string
	Acara          string
	Materi         string
	Catatan        string
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
	Tanggal        string `json:"tanggal"`
	Hari           string `json:"hari"`
	Jam            string `json:"jam"`
	Acara          string `json:"acara"`
	Materi         string `json:"materi"`
	Catatan        string `json:"catatan"`
	GeneratedText  string `json:"generated_text"`
	Status         string `json:"status"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
