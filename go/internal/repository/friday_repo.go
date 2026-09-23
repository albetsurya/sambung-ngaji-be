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

const fridaySelectCols = `friday_id, tanggal, khatib_imam, muadzin, penasihat,
	petugas_parkir, penata_sandal, catatan, created_by, created_at, updated_at`

// FindByRange: jadwal jumat dalam rentang tanggal ( inklusif ), urut naik.
func (r *FridayRepo) FindByRange(ctx context.Context, from, to string) ([]model.FridaySchedule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules
		 WHERE ($1 = '' OR tanggal >= $1::date) AND ($2 = '' OR tanggal <= $2::date)
		 ORDER BY tanggal ASC`,
		from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFridays(rows)
}

// FindByDate: satu jadwal berdasarkan tanggal (YYYY-MM-DD).
func (r *FridayRepo) FindByDate(ctx context.Context, tanggal string) (*model.FridaySchedule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules WHERE tanggal=$1::date`, tanggal)
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
		   (friday_id, tanggal, khatib_imam, muadzin, penasihat,
		    petugas_parkir, penata_sandal, catatan, created_by, created_at, updated_at)
		 VALUES ($1, $2::date, $3, $4, $5, $6, $7, $8, $9, now(), now())
		 ON CONFLICT (tanggal)
		 DO UPDATE SET khatib_imam    = EXCLUDED.khatib_imam,
		               muadzin        = EXCLUDED.muadzin,
		               penasihat      = EXCLUDED.penasihat,
		               petugas_parkir = EXCLUDED.petugas_parkir,
		               penata_sandal  = EXCLUDED.penata_sandal,
		               catatan        = EXCLUDED.catatan,
		               created_by     = EXCLUDED.created_by,
		               updated_at     = now()`,
		f.FridayID, f.Tanggal, f.KhatibImam, f.Muadzin, f.Penasihat,
		f.PetugasParkir, f.PenataSandal, f.Catatan, f.CreatedBy)
	return err
}

// Delete: hapus berdasarkan tanggal (YYYY-MM-DD).
func (r *FridayRepo) Delete(ctx context.Context, tanggal string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM friday_schedules WHERE tanggal=$1::date`, tanggal)
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
	err := row.Scan(&f.FridayID, &f.Tanggal, &f.KhatibImam, &f.Muadzin,
		&f.Penasihat, &f.PetugasParkir, &f.PenataSandal, &f.Catatan,
		&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func scanFridays(rows rowsScanner) ([]model.FridaySchedule, error) {
	var out []model.FridaySchedule
	for rows.Next() {
		var f model.FridaySchedule
		if err := rows.Scan(&f.FridayID, &f.Tanggal, &f.KhatibImam, &f.Muadzin,
			&f.Penasihat, &f.PetugasParkir, &f.PenataSandal, &f.Catatan,
			&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
