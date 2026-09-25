package model

import "time"

type PendingMember struct {
	SubmissionID           string
	GroupID                *string
	NamaLengkap            string
	NamaPanggilan          string
	JenisKelamin           *string
	TempatLahir            string
	TanggalLahir           *time.Time
	NoWA                   string
	AlamatRumah            string
	Desa                   string
	Daerah                 string
	Pekerjaan              string
	Hobi                   string
	IsNikah                bool
	JenjangPendidikan      string
	Sekolah                string
	Jurusan                string
	TahunMulaiPendidikan   string
	TahunSelesaiPendidikan string
	FotoURL                string
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
	NamaLengkap            string `json:"nama_lengkap"`
	NamaPanggilan          string `json:"nama_panggilan"`
	JenisKelamin           string `json:"jenis_kelamin"`
	TempatLahir            string `json:"tempat_lahir"`
	TanggalLahir           string `json:"tanggal_lahir"`
	NoWA                   string `json:"no_wa"`
	AlamatRumah            string `json:"alamat_rumah"`
	Desa                   string `json:"desa"`
	Daerah                 string `json:"daerah"`
	Pekerjaan              string `json:"pekerjaan"`
	Hobi                   string `json:"hobi"`
	IsNikah                bool   `json:"is_nikah"`
	JenjangPendidikan      string `json:"jenjang_pendidikan"`
	Sekolah                string `json:"sekolah"`
	Jurusan                string `json:"jurusan"`
	TahunMulaiPendidikan   string `json:"tahun_mulai_pendidikan"`
	TahunSelesaiPendidikan string `json:"tahun_selesai_pendidikan"`
	FotoURL                string `json:"foto_url"`
	Username               string `json:"username"`
	Status                 string `json:"status"`
	SubmittedAt            string `json:"submitted_at"`
	SubmittedIP            string `json:"submitted_ip"`
	ReviewedBy             string `json:"reviewed_by"`
	ReviewedAt             string `json:"reviewed_at"`
	RejectionReason        string `json:"rejection_reason"`
	CreatedMemberID        string `json:"created_member_id"`
}
