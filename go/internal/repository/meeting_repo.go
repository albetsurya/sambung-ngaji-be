package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MeetingRepo struct {
	pool *pgxpool.Pool
}

func NewMeetingRepo(pool *pgxpool.Pool) *MeetingRepo {
	return &MeetingRepo{pool: pool}
}

const meetingSelectCols = `
	meeting_id, date, day, time, start_time, group_id, event, topic,
	status, notes, target_categories, gender_target, created_by, created_at, updated_at`

func (r *MeetingRepo) FindAll(ctx context.Context, f model.MeetingListFilter) ([]model.Meeting, error) {
	q := `SELECT ` + meetingSelectCols + ` FROM meetings WHERE 1=1`
	args := []interface{}{}
	n := 1

	if f.From != "" {
		q += ` AND date >= $` + itoa(n)
		args = append(args, f.From)
		n++
	}
	if f.To != "" {
		q += ` AND date <= $` + itoa(n)
		args = append(args, f.To)
		n++
	}
	if f.GroupID != "" {
		q += ` AND group_id = $` + itoa(n)
		args = append(args, f.GroupID)
	}
	q += ` ORDER BY date DESC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMeetings(rows)
}

func (r *MeetingRepo) FindByID(ctx context.Context, id string) (*model.Meeting, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+meetingSelectCols+` FROM meetings WHERE meeting_id = $1`, id)
	return scanMeeting(row)
}

func scanMeeting(s rowScanner) (*model.Meeting, error) {
	var m model.Meeting
	var kat string
	err := s.Scan(
		&m.MeetingID, &m.Date, &m.Day, &m.Time, &m.StartTime, &m.GroupID,
		&m.Event, &m.Topic, &m.Status, &m.Notes, &kat, &m.GenderTarget,
		&m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.TargetCategories = parseJSONArray(kat)
	return &m, nil
}

func scanMeetings(rows rowsScanner) ([]model.Meeting, error) {
	var out []model.Meeting
	for rows.Next() {
		var m model.Meeting
		var kat string
		err := rows.Scan(
			&m.MeetingID, &m.Date, &m.Day, &m.Time, &m.StartTime, &m.GroupID,
			&m.Event, &m.Topic, &m.Status, &m.Notes, &kat, &m.GenderTarget,
			&m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.TargetCategories = parseJSONArray(kat)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MeetingRepo) Create(ctx context.Context, m *model.Meeting) error {
	kat := marshalJSONArray(m.TargetCategories)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO meetings
		(meeting_id, date, day, time, start_time, group_id, event, topic,
		 status, notes, target_categories, gender_target, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,now(),now())
	`, m.MeetingID, m.Date, m.Day, m.Time, m.StartTime, m.GroupID,
		m.Event, m.Topic, m.Status, m.Notes, kat, m.GenderTarget, m.CreatedBy)
	return err
}

type MeetingPatch struct {
	Date             *string
	Day              *string
	Time             *string
	StartTime        *string
	GroupID          *string
	Event            *string
	Topic            *string
	Status           *string
	Notes            *string
	TargetCategories *[]string
	GenderTarget     *string
}

func (r *MeetingRepo) Update(ctx context.Context, id string, p MeetingPatch) error {
	q := `UPDATE meetings SET updated_at = now()`
	args := []interface{}{}
	n := 1

	addStr := func(col string, v *string) {
		if v != nil {
			q += `, ` + col + ` = $` + itoa(n)
			args = append(args, *v)
			n++
		}
	}
	addStr("date", p.Date)
	addStr("day", p.Day)
	addStr("time", p.Time)
	addStr("start_time", p.StartTime)
	addStr("group_id", p.GroupID)
	addStr("event", p.Event)
	addStr("topic", p.Topic)
	addStr("status", p.Status)
	addStr("notes", p.Notes)
	if p.GenderTarget != nil {
		/* "" / null = Semua -> tulis NULL agar lolos CHECK
		   (hanya 'L'/'P'/NULL yang valid). */
		if *p.GenderTarget == "" {
			q += `, gender_target = NULL`
		} else {
			q += `, gender_target = $` + itoa(n)
			args = append(args, *p.GenderTarget)
			n++
		}
	}

	if p.TargetCategories != nil {
		q += `, target_categories = $` + itoa(n) + `::jsonb`
		args = append(args, marshalJSONArray(*p.TargetCategories))
		n++
	}

	q += ` WHERE meeting_id = $` + itoa(n)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *MeetingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM meetings WHERE meeting_id = $1`, id)
	return err
}

func (r *MeetingRepo) DeleteMany(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`DELETE FROM meetings WHERE meeting_id = ANY($1)`, ids)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type ReminderMeetingRow struct {
	MeetingID string
	Event     string
	Date      string
	Time      string
}

func (r *MeetingRepo) FindPendingReminder(ctx context.Context, from, to time.Time) ([]ReminderMeetingRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT meeting_id, COALESCE(event, ''),
		       TO_CHAR(date, 'YYYY-MM-DD'),
		       COALESCE(time, '')
		FROM meetings
		WHERE reminder_sent_at IS NULL
		  AND status != 'LIBUR'
		  AND send_reminder = true
		  AND (
		    (date::date + CASE
		       WHEN time ~ '^[0-9]{1,2}:[0-9]{2}' THEN time::time
		       ELSE TIME '19:00'
		     END)::timestamptz
		  ) BETWEEN $1 AND $2
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ReminderMeetingRow
	for rows.Next() {
		var row ReminderMeetingRow
		if err := rows.Scan(&row.MeetingID, &row.Event, &row.Date, &row.Time); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *MeetingRepo) MarkReminderSent(ctx context.Context, meetingID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE meetings SET reminder_sent_at = now() WHERE meeting_id = $1`,
		meetingID,
	)
	return err
}
