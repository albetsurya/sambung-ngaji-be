package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type FridayRepo struct {
	pool *pgxpool.Pool
}

func NewFridayRepo(pool *pgxpool.Pool) *FridayRepo {
	return &FridayRepo{pool: pool}
}

const fridaySelectCols = `friday_id, group_id, tanggal, khatib_imam, muadzin, penasihat,
	petugas_parkir, penata_sandal, catatan, created_by, created_at, updated_at,
	reminder_sent_at`

// FindByRange: jadwal jumat dalam rentang tanggal ( inklusif ), urut naik.
func (r *FridayRepo) FindByRange(ctx context.Context, groupID, from, to string) ([]model.FridaySchedule, error) {
	q := `SELECT ` + fridaySelectCols + ` FROM friday_schedules
		 WHERE ($1 = '' OR tanggal >= $1::date) AND ($2 = '' OR tanggal <= $2::date)`
	args := []interface{}{from, to}
	if groupID != "" {
		q += ` AND (group_id = $3 OR group_id IS NULL)`
		args = append(args, groupID)
	}
	q += ` ORDER BY tanggal ASC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFridays(rows)
}

// FindByDate: satu jadwal berdasarkan tanggal (YYYY-MM-DD).
func (r *FridayRepo) FindByDate(ctx context.Context, groupID, tanggal string) (*model.FridaySchedule, error) {
	q := `SELECT ` + fridaySelectCols + ` FROM friday_schedules WHERE tanggal=$1::date`
	args := []interface{}{tanggal}
	if groupID != "" {
		q += ` AND (group_id = $2 OR group_id IS NULL)`
		args = append(args, groupID)
	}
	row := r.pool.QueryRow(ctx, q, args...)
	f, err := scanFriday(row)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Upsert: insert atau update berdasarkan tanggal (1 jadwal per tanggal).
func (r *FridayRepo) Upsert(ctx context.Context, f *model.FridaySchedule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO friday_schedules
		   (friday_id, group_id, tanggal, khatib_imam, muadzin, penasihat,
		    petugas_parkir, penata_sandal, catatan, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3::date, $4, $5, $6, $7, $8, $9, $10, now(), now())
		 ON CONFLICT (tanggal)
		 DO UPDATE SET group_id       = EXCLUDED.group_id,
		               khatib_imam    = EXCLUDED.khatib_imam,
		               muadzin        = EXCLUDED.muadzin,
		               penasihat      = EXCLUDED.penasihat,
		               petugas_parkir = EXCLUDED.petugas_parkir,
		               penata_sandal  = EXCLUDED.penata_sandal,
		               catatan        = EXCLUDED.catatan,
		               created_by     = EXCLUDED.created_by,
		               updated_at     = now()`,
		f.FridayID, f.GroupID, f.Tanggal, f.KhatibImam, f.Muadzin, f.Penasihat,
		f.PetugasParkir, f.PenataSandal, f.Catatan, f.CreatedBy)
	return err
}

// Delete: hapus berdasarkan tanggal (YYYY-MM-DD).
func (r *FridayRepo) Delete(ctx context.Context, groupID, tanggal string) error {
	q := `DELETE FROM friday_schedules WHERE tanggal=$1::date`
	args := []interface{}{tanggal}
	if groupID != "" {
		q += ` AND (group_id = $2 OR group_id IS NULL)`
		args = append(args, groupID)
	}
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

// FindUnsentByDate: satu jadwal yang reminder WA-nya belum terkirim.
// Return pgx.ErrNoRows kalau tidak ada (sudah terkirim / belum ada jadwal).
func (r *FridayRepo) FindUnsentByDate(ctx context.Context, tanggal string) (*model.FridaySchedule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules
		 WHERE tanggal=$1::date AND reminder_sent_at IS NULL`, tanggal)
	f, err := scanFriday(row)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// MarkReminderSent: tandai reminder WA sudah terkirim untuk satu tanggal.
func (r *FridayRepo) MarkReminderSent(ctx context.Context, tanggal string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE friday_schedules SET reminder_sent_at=now(), updated_at=now()
		 WHERE tanggal=$1::date`, tanggal)
	return err
}

func scanFriday(row pgx.Row) (model.FridaySchedule, error) {
	var f model.FridaySchedule
	err := row.Scan(&f.FridayID, &f.GroupID, &f.Tanggal, &f.KhatibImam, &f.Muadzin,
		&f.Penasihat, &f.PetugasParkir, &f.PenataSandal, &f.Catatan,
		&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt, &f.ReminderSentAt)
	return f, err
}

func scanFridays(rows rowsScanner) ([]model.FridaySchedule, error) {
	var out []model.FridaySchedule
	for rows.Next() {
		var f model.FridaySchedule
		if err := rows.Scan(&f.FridayID, &f.GroupID, &f.Tanggal, &f.KhatibImam, &f.Muadzin,
			&f.Penasihat, &f.PetugasParkir, &f.PenataSandal, &f.Catatan,
			&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt, &f.ReminderSentAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// MaxReminderSent — waktu kirim terakhir (NULL bila belum pernah).
func (r *FridayRepo) MaxReminderSent(ctx context.Context) (*string, error) {
	var v *string
	err := r.pool.QueryRow(ctx,
		`SELECT to_char(MAX(reminder_sent_at) AT TIME ZONE 'Asia/Jakarta',
		 'YYYY-MM-DD"T"HH24:MI:SSOF') FROM friday_schedules`).Scan(&v)
	return v, err
}

// FindUpcoming — N jadwal ke depan mulai hari ini, urut naik.
func (r *FridayRepo) FindUpcoming(ctx context.Context, limit int) ([]model.FridaySchedule, error) {
	if limit <= 0 {
		limit = 4
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules
		 WHERE tanggal >= CURRENT_DATE ORDER BY tanggal ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFridays(rows)
}
