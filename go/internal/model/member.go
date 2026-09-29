package model

import "time"

type Member struct {
	MemberID               string
	GroupID                *string
	NamaLengkap            string
	NamaPanggilan          string
	JenisKelamin           *string
	TempatLahir            string
	TanggalLahir           *time.Time
	FotoURL                string
	NoWA                   string
	AlamatRumah            string
	Desa                   string
	Daerah                 string
	Kelompok               string
	IsMuballigh            bool
	IsKerja                bool
	IsNikah                bool
	TinggiBadan            string
	BeratBadan             string
	Hobi                   string
	Pekerjaan              string
	StatusPembinaan        string
	StatusAktif            bool
	TanggalMasuk           *time.Time
	TanggalKeluar          *time.Time
	JenjangPendidikan      string
	Sekolah                string
	Jurusan                string
	TahunMulaiPendidikan   string
	TahunSelesaiPendidikan string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type MemberListDTO struct {
	MemberID      string `json:"member_id"`
	GroupID       string `json:"group_id"`
	NamaLengkap   string `json:"nama_lengkap"`
	NamaPanggilan string `json:"nama_panggilan"`
	JenisKelamin  string `json:"jenis_kelamin"`
	Kelompok      string `json:"kelompok"`
	GroupName     string `json:"group_name"`
	Kategori      string `json:"kategori"`
	FotoURL       string `json:"foto_url"`
	HasUser       bool   `json:"has_user"`
}

type AttendanceMemberDTO struct {
	MemberID     string `json:"member_id"`
	GroupID      string `json:"group_id"`
	NamaLengkap  string `json:"nama_lengkap"`
	Kelompok     string `json:"kelompok"`
	Kategori     string `json:"kategori"`
	JenisKelamin string `json:"jenis_kelamin"`
}

type MemberDetailDTO struct {
	MemberID               string `json:"member_id"`
	GroupID                string `json:"group_id"`
	NamaLengkap            string `json:"nama_lengkap"`
	NamaPanggilan          string `json:"nama_panggilan"`
	JenisKelamin           string `json:"jenis_kelamin"`
	TempatLahir            string `json:"tempat_lahir"`
	TanggalLahir           string `json:"tanggal_lahir"`
	FotoURL                string `json:"foto_url"`
	NoWA                   string `json:"no_wa"`
	AlamatRumah            string `json:"alamat_rumah"`
	Desa                   string `json:"desa"`
	Daerah                 string `json:"daerah"`
	Kelompok               string `json:"kelompok"`
	GroupName              string `json:"group_name"`
	IsMuballigh            bool   `json:"is_muballigh"`
	IsKerja                bool   `json:"is_kerja"`
	IsNikah                bool   `json:"is_nikah"`
	TinggiBadan            string `json:"tinggi_badan"`
	BeratBadan             string `json:"berat_badan"`
	Hobi                   string `json:"hobi"`
	Pekerjaan              string `json:"pekerjaan"`
	StatusPembinaan        string `json:"status_pembinaan"`
	StatusAktif            bool   `json:"status_aktif"`
	TanggalMasuk           string `json:"tanggal_masuk"`
	TanggalKeluar          string `json:"tanggal_keluar"`
	JenjangPendidikan      string `json:"jenjang_pendidikan"`
	Sekolah                string `json:"sekolah"`
	Jurusan                string `json:"jurusan"`
	TahunMulaiPendidikan   string `json:"tahun_mulai_pendidikan"`
	TahunSelesaiPendidikan string `json:"tahun_selesai_pendidikan"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	Kategori               string `json:"kategori"`
	Usia                   int    `json:"usia"`
	Pendidikan             []any  `json:"pendidikan"`
	HasUser                bool   `json:"has_user"`
}

type MemberListFilter struct {
	Search          string
	Kelompok        string
	JenisKelamin    string
	Desa            string
	Kategori        string
	IncludeInactive bool
	Limit           int
	Offset          int
}
