package repository

import (
	"context"

	"pengajian-backend/internal/model"
)

type NewMemberInput struct {
	MemberID               string
	GroupID                *string
	NamaLengkap            string
	NamaPanggilan          string
	JenisKelamin           *string
	TempatLahir            string
	TanggalLahir           interface{}
	FotoURL                string
	NoWA                   string
	AlamatRumah            string
	Desa                   string
	Daerah                 string
	Kelompok               string
	IsMuballigh            bool
	IsKerja                bool
	IsNikah                bool
	Hobi                   string
	Pekerjaan              string
	StatusPembinaan        string
	JenjangPendidikan      string
	Sekolah                string
	Jurusan                string
	TahunMulaiPendidikan   string
	TahunSelesaiPendidikan string
}

func (r *MemberRepo) Insert(ctx context.Context, in NewMemberInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO members (
			member_id, group_id, nama_lengkap, nama_panggilan, jenis_kelamin,
			tempat_lahir, tanggal_lahir, foto_url, no_wa,
			alamat_rumah, desa, daerah, kelompok,
			is_muballigh, is_kerja, is_nikah,
			hobi, pekerjaan, status_pembinaan, status_aktif,
			tanggal_masuk, jenjang_pendidikan, sekolah, jurusan,
			tahun_mulai_pendidikan, tahun_selesai_pendidikan,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,true,
			now(),$19,$20,$21,$22,$23,now(),now()
		)
	`, in.MemberID, in.GroupID, in.NamaLengkap, in.NamaPanggilan, in.JenisKelamin,
		in.TempatLahir, in.TanggalLahir, in.FotoURL, in.NoWA,
		in.AlamatRumah, in.Desa, in.Daerah, in.Kelompok,
		in.IsMuballigh, in.IsKerja, in.IsNikah,
		in.Hobi, in.Pekerjaan, in.StatusPembinaan,
		in.JenjangPendidikan, in.Sekolah, in.Jurusan,
		in.TahunMulaiPendidikan, in.TahunSelesaiPendidikan)
	return err
}

var _ = model.Member{}
