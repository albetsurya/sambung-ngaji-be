package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberExportRow struct {
	MemberID          string
	NamaLengkap       string
	NamaPanggilan     string
	JenisKelamin      *string
	TempatLahir       string
	TanggalLahir      *time.Time
	Desa              string
	Daerah            string
	AlamatRumah       string
	NoWA              string
	Pekerjaan         string
	Hobi              string
	StatusPembinaan   string
	Kelompok          string
	JenjangPendidikan string
	IsNikah           bool
	StatusAktif       bool
}

func (r *MemberRepo) Deactivate(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE members
		SET status_aktif = false, tanggal_keluar = $1, updated_at = now()
		WHERE member_id = $2
	`, time.Now().Format("2006-01-02"), id)
	return err
}

func (r *MemberRepo) FindAllIncludingInactive(ctx context.Context) ([]MemberExportRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT member_id, nama_lengkap, nama_panggilan, jenis_kelamin,
		       tempat_lahir, tanggal_lahir, desa, daerah, alamat_rumah,
		       no_wa, pekerjaan, hobi, status_pembinaan, kelompok,
		       jenjang_pendidikan, is_nikah, status_aktif
		FROM members
		ORDER BY nama_lengkap
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MemberExportRow
	for rows.Next() {
		var m MemberExportRow
		if err := rows.Scan(
			&m.MemberID, &m.NamaLengkap, &m.NamaPanggilan, &m.JenisKelamin,
			&m.TempatLahir, &m.TanggalLahir, &m.Desa, &m.Daerah, &m.AlamatRumah,
			&m.NoWA, &m.Pekerjaan, &m.Hobi, &m.StatusPembinaan, &m.Kelompok,
			&m.JenjangPendidikan, &m.IsNikah, &m.StatusAktif,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRepo) Pool() *pgxpool.Pool { return r.pool }
func (r *MemberRepo) UpdateFotoURL(ctx context.Context, memberID, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE members SET foto_url = $1, updated_at = now() WHERE member_id = $2
	`, url, memberID)
	return err
}

func (r *MemberRepo) GetFotoURL(ctx context.Context, memberID string) (string, error) {
	var url string
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(foto_url, '') FROM members WHERE member_id = $1`, memberID,
	).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}
