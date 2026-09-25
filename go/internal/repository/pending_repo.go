package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type PendingRepo struct {
	pool *pgxpool.Pool
}

func NewPendingRepo(pool *pgxpool.Pool) *PendingRepo {
	return &PendingRepo{pool: pool}
}

const pendingSelectCols = `
	submission_id, group_id, nama_lengkap, nama_panggilan, jenis_kelamin,
	tempat_lahir, tanggal_lahir, no_wa, alamat_rumah, desa, daerah,
	pekerjaan, hobi, is_nikah, jenjang_pendidikan, sekolah, jurusan,
	tahun_mulai_pendidikan, tahun_selesai_pendidikan, foto_url,
	username, password_hash, status, submitted_at, submitted_ip,
	reviewed_by, reviewed_at, rejection_reason, created_member_id`

func (r *PendingRepo) FindAll(ctx context.Context, groupID, status string) ([]model.PendingMember, error) {
	q := `SELECT ` + pendingSelectCols + ` FROM pending_members WHERE 1=1`
	args := []interface{}{}
	n := 1
	if groupID != "" {
		q += ` AND group_id = $` + itoa(n)
		args = append(args, groupID)
		n++
	}
	if status != "" {
		q += ` AND status = $` + itoa(n)
		args = append(args, status)
		n++
	}
	q += ` ORDER BY submitted_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPendings(rows)
}

func (r *PendingRepo) FindByID(ctx context.Context, id string) (*model.PendingMember, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+pendingSelectCols+` FROM pending_members WHERE submission_id = $1`, id)
	return scanPending(row)
}

func (r *PendingRepo) CountByIPToday(ctx context.Context, ip, today string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pending_members
		WHERE submitted_ip = $1 AND submitted_at::date = $2::date
	`, ip, today).Scan(&n)
	return n, err
}

func (r *PendingRepo) CountByWAPending(ctx context.Context, noWA string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pending_members
		WHERE no_wa = $1 AND status = 'PENDING'
	`, noWA).Scan(&n)
	return n, err
}

func (r *PendingRepo) CountByUsernamePending(ctx context.Context, username string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pending_members
		WHERE username = $1 AND status = 'PENDING'
	`, username).Scan(&n)
	return n, err
}

func (r *PendingRepo) Insert(ctx context.Context, p *model.PendingMember) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO pending_members
		(submission_id, group_id, nama_lengkap, nama_panggilan, jenis_kelamin,
		 tempat_lahir, tanggal_lahir, no_wa, alamat_rumah, desa, daerah,
		 pekerjaan, hobi, is_nikah, jenjang_pendidikan, sekolah, jurusan,
		 tahun_mulai_pendidikan, tahun_selesai_pendidikan, foto_url,
		 username, password_hash, status, submitted_at, submitted_ip,
		 reviewed_by, reviewed_at, rejection_reason, created_member_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,now(),$24,$25,$26,$27,$28)
	`, p.SubmissionID, p.GroupID, p.NamaLengkap, p.NamaPanggilan, p.JenisKelamin,
		p.TempatLahir, p.TanggalLahir, p.NoWA, p.AlamatRumah, p.Desa, p.Daerah,
		p.Pekerjaan, p.Hobi, p.IsNikah, p.JenjangPendidikan, p.Sekolah, p.Jurusan,
		p.TahunMulaiPendidikan, p.TahunSelesaiPendidikan, p.FotoURL,
		p.Username, p.PasswordHash, p.Status, p.SubmittedIP,
		p.ReviewedBy, p.ReviewedAt, p.RejectionReason, p.CreatedMemberID)
	return err
}

func (r *PendingRepo) UpdateApproved(ctx context.Context, id, reviewedBy, memberID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pending_members
		SET status = 'APPROVED', reviewed_by = $1, reviewed_at = now(), created_member_id = $2
		WHERE submission_id = $3
	`, reviewedBy, memberID, id)
	return err
}

func (r *PendingRepo) UpdateRejected(ctx context.Context, id, reviewedBy, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pending_members
		SET status = 'REJECTED', reviewed_by = $1, reviewed_at = now(), rejection_reason = $2
		WHERE submission_id = $3
	`, reviewedBy, reason, id)
	return err
}

func scanPending(s rowScanner) (*model.PendingMember, error) {
	var p model.PendingMember
	err := s.Scan(
		&p.SubmissionID, &p.GroupID, &p.NamaLengkap, &p.NamaPanggilan, &p.JenisKelamin,
		&p.TempatLahir, &p.TanggalLahir, &p.NoWA, &p.AlamatRumah, &p.Desa, &p.Daerah,
		&p.Pekerjaan, &p.Hobi, &p.IsNikah, &p.JenjangPendidikan, &p.Sekolah, &p.Jurusan,
		&p.TahunMulaiPendidikan, &p.TahunSelesaiPendidikan, &p.FotoURL,
		&p.Username, &p.PasswordHash, &p.Status, &p.SubmittedAt, &p.SubmittedIP,
		&p.ReviewedBy, &p.ReviewedAt, &p.RejectionReason, &p.CreatedMemberID,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPendings(rows rowsScanner) ([]model.PendingMember, error) {
	var out []model.PendingMember
	for rows.Next() {
		var p model.PendingMember
		err := rows.Scan(
			&p.SubmissionID, &p.GroupID, &p.NamaLengkap, &p.NamaPanggilan, &p.JenisKelamin,
			&p.TempatLahir, &p.TanggalLahir, &p.NoWA, &p.AlamatRumah, &p.Desa, &p.Daerah,
			&p.Pekerjaan, &p.Hobi, &p.IsNikah, &p.JenjangPendidikan, &p.Sekolah, &p.Jurusan,
			&p.TahunMulaiPendidikan, &p.TahunSelesaiPendidikan, &p.FotoURL,
			&p.Username, &p.PasswordHash, &p.Status, &p.SubmittedAt, &p.SubmittedIP,
			&p.ReviewedBy, &p.ReviewedAt, &p.RejectionReason, &p.CreatedMemberID,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
