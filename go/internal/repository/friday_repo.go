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

const fridaySelectCols = `friday_id, group_id, date, sermon_leader, muadzin, advisor,
	parking_attendant, footwear_attendant, notes, created_by, created_at, updated_at`

func (r *FridayRepo) FindByRange(ctx context.Context, groupID, from, to string) ([]model.FridaySchedule, error) {
	q := `SELECT ` + fridaySelectCols + ` FROM friday_schedules
		 WHERE ($1 = '' OR date >= $1::date) AND ($2 = '' OR date <= $2::date)`
	args := []interface{}{from, to}
	if groupID != "" {
		q += ` AND (group_id = $3 OR group_id IS NULL)`
		args = append(args, groupID)
	}
	q += ` ORDER BY date ASC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFridays(rows)
}

func (r *FridayRepo) FindByDate(ctx context.Context, groupID, date string) (*model.FridaySchedule, error) {
	q := `SELECT ` + fridaySelectCols + ` FROM friday_schedules WHERE date=$1::date`
	args := []interface{}{date}
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

func (r *FridayRepo) Upsert(ctx context.Context, f *model.FridaySchedule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO friday_schedules
		   (friday_id, group_id, date, sermon_leader, muadzin, advisor,
		    parking_attendant, footwear_attendant, notes, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3::date, $4, $5, $6, $7, $8, $9, $10, now(), now())
		 ON CONFLICT (date)
		 DO UPDATE SET group_id       = EXCLUDED.group_id,
		               sermon_leader    = EXCLUDED.sermon_leader,
		               muadzin        = EXCLUDED.muadzin,
		               advisor      = EXCLUDED.advisor,
		               parking_attendant = EXCLUDED.parking_attendant,
		               footwear_attendant  = EXCLUDED.footwear_attendant,
		               notes        = EXCLUDED.notes,
		               created_by     = EXCLUDED.created_by,
		               updated_at     = now()`,
		f.FridayID, f.GroupID, f.Date, f.SermonLeader, f.Muadzin, f.Advisor,
		f.ParkingAttendant, f.FootwearAttendant, f.Notes, f.CreatedBy)
	return err
}

func (r *FridayRepo) Delete(ctx context.Context, groupID, date string) error {
	q := `DELETE FROM friday_schedules WHERE date=$1::date`
	args := []interface{}{date}
	if groupID != "" {
		q += ` AND (group_id = $2 OR group_id IS NULL)`
		args = append(args, groupID)
	}
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *FridayRepo) FindUnsentByDate(ctx context.Context, date string) (*model.FridaySchedule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules
		 WHERE date=$1::date AND reminder_sent_at IS NULL`, date)
	f, err := scanFriday(row)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FridayRepo) MarkReminderSent(ctx context.Context, date string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE friday_schedules SET reminder_sent_at=now(), updated_at=now()
		 WHERE date=$1::date`, date)
	return err
}

func scanFriday(row pgx.Row) (model.FridaySchedule, error) {
	var f model.FridaySchedule
	err := row.Scan(&f.FridayID, &f.GroupID, &f.Date, &f.SermonLeader, &f.Muadzin,
		&f.Advisor, &f.ParkingAttendant, &f.FootwearAttendant, &f.Notes,
		&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func scanFridays(rows rowsScanner) ([]model.FridaySchedule, error) {
	var out []model.FridaySchedule
	for rows.Next() {
		var f model.FridaySchedule
		if err := rows.Scan(&f.FridayID, &f.GroupID, &f.Date, &f.SermonLeader, &f.Muadzin,
			&f.Advisor, &f.ParkingAttendant, &f.FootwearAttendant, &f.Notes,
			&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FridayRepo) MaxReminderSent(ctx context.Context) (*string, error) {
	var v *string
	err := r.pool.QueryRow(ctx,
		`SELECT to_char(MAX(reminder_sent_at) AT TIME ZONE 'Asia/Jakarta',
		 'YYYY-MM-DD"T"HH24:MI:SSOF') FROM friday_schedules`).Scan(&v)
	return v, err
}

func (r *FridayRepo) FindUpcoming(ctx context.Context, limit int) ([]model.FridaySchedule, error) {
	if limit <= 0 {
		limit = 4
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+fridaySelectCols+` FROM friday_schedules
		 WHERE date >= CURRENT_DATE ORDER BY date ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFridays(rows)
}
