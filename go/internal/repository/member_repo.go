package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MemberRepo struct {
	pool *pgxpool.Pool
}

func NewMemberRepo(pool *pgxpool.Pool) *MemberRepo {
	return &MemberRepo{pool: pool}
}

const memberSelectCols = `
	member_id, nama_lengkap, nama_panggilan, jenis_kelamin,
	tempat_lahir, tanggal_lahir, foto_url, no_wa,
	alamat_rumah, desa, daerah, kelompok,
	is_muballigh, is_kerja, is_nikah, tinggi_badan, berat_badan,
	hobi, pekerjaan, status_pembinaan, status_aktif,
	tanggal_masuk, tanggal_keluar,
	jenjang_pendidikan, sekolah, jurusan,
	tahun_mulai_pendidikan, tahun_selesai_pendidikan,
	created_at, updated_at`

func (r *MemberRepo) FindAll(ctx context.Context) ([]model.Member, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+memberSelectCols+` FROM members WHERE status_aktif = true ORDER BY nama_lengkap`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMembers(rows)
}

func (r *MemberRepo) FindByID(ctx context.Context, id string) (*model.Member, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+memberSelectCols+` FROM members WHERE member_id = $1`, id)
	return scanMember(row)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMember(s rowScanner) (*model.Member, error) {
	var m model.Member
	err := s.Scan(
		&m.MemberID, &m.NamaLengkap, &m.NamaPanggilan, &m.JenisKelamin,
		&m.TempatLahir, &m.TanggalLahir, &m.FotoURL, &m.NoWA,
		&m.AlamatRumah, &m.Desa, &m.Daerah, &m.Kelompok,
		&m.IsMuballigh, &m.IsKerja, &m.IsNikah, &m.TinggiBadan, &m.BeratBadan,
		&m.Hobi, &m.Pekerjaan, &m.StatusPembinaan, &m.StatusAktif,
		&m.TanggalMasuk, &m.TanggalKeluar,
		&m.JenjangPendidikan, &m.Sekolah, &m.Jurusan,
		&m.TahunMulaiPendidikan, &m.TahunSelesaiPendidikan,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type rowsScanner interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

func scanMembers(rows rowsScanner) ([]model.Member, error) {
	var out []model.Member
	for rows.Next() {
		var m model.Member
		err := rows.Scan(
			&m.MemberID, &m.NamaLengkap, &m.NamaPanggilan, &m.JenisKelamin,
			&m.TempatLahir, &m.TanggalLahir, &m.FotoURL, &m.NoWA,
			&m.AlamatRumah, &m.Desa, &m.Daerah, &m.Kelompok,
			&m.IsMuballigh, &m.IsKerja, &m.IsNikah, &m.TinggiBadan, &m.BeratBadan,
			&m.Hobi, &m.Pekerjaan, &m.StatusPembinaan, &m.StatusAktif,
			&m.TanggalMasuk, &m.TanggalKeluar,
			&m.JenjangPendidikan, &m.Sekolah, &m.Jurusan,
			&m.TahunMulaiPendidikan, &m.TahunSelesaiPendidikan,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi rows: %w", err)
	}
	return out, nil
}
