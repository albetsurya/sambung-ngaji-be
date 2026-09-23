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
	meeting_id, tanggal, hari, jam, jam_start, group_id, acara, materi,
	status, catatan, kategori_target, gender_target, send_reminder, created_by, created_at, updated_at`

func (r *MeetingRepo) FindAll(ctx context.Context, f model.MeetingListFilter) ([]model.Meeting, error) {
	q := `SELECT ` + meetingSelectCols + ` FROM meetings WHERE 1=1`
	args := []interface{}{}
	n := 1

	if f.From != "" {
		q += ` AND tanggal >= $` + itoa(n)
		args = append(args, f.From)
		n++
	}
	if f.To != "" {
		q += ` AND tanggal <= $` + itoa(n)
		args = append(args, f.To)
		n++
	}
	if f.GroupID != "" {
		q += ` AND group_id = $` + itoa(n)
		args = append(args, f.GroupID)
	}
	q += ` ORDER BY tanggal DESC`

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
	var sendReminder bool
	err := s.Scan(
		&m.MeetingID, &m.Tanggal, &m.Hari, &m.Jam, &m.JamStart, &m.GroupID,
		&m.Acara, &m.Materi, &m.Status, &m.Catatan, &kat, &m.GenderTarget,
		&sendReminder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.KategoriTarget = parseJSONArray(kat)
	m.SendReminder = &sendReminder
	return &m, nil
}

func scanMeetings(rows rowsScanner) ([]model.Meeting, error) {
	var out []model.Meeting
	for rows.Next() {
		var m model.Meeting
		var kat string
		var sendReminder bool
		err := rows.Scan(
			&m.MeetingID, &m.Tanggal, &m.Hari, &m.Jam, &m.JamStart, &m.GroupID,
			&m.Acara, &m.Materi, &m.Status, &m.Catatan, &kat, &m.GenderTarget,
			&sendReminder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.KategoriTarget = parseJSONArray(kat)
		m.SendReminder = &sendReminder
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MeetingRepo) Create(ctx context.Context, m *model.Meeting) error {
	kat := marshalJSONArray(m.KategoriTarget)
	sendReminder := true
	if m.SendReminder != nil {
		sendReminder = *m.SendReminder
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO meetings
		(meeting_id, tanggal, hari, jam, jam_start, group_id, acara, materi,
		 status, catatan, kategori_target, gender_target, send_reminder, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,$14,now(),now())
	`, m.MeetingID, m.Tanggal, m.Hari, m.Jam, m.JamStart, m.GroupID,
		m.Acara, m.Materi, m.Status, m.Catatan, kat, m.GenderTarget, sendReminder, m.CreatedBy)
	return err
}

type MeetingPatch struct {
	Tanggal        *string
	Hari           *string
	Jam            *string
	JamStart       *string
	GroupID        *string
	Acara          *string
	Materi         *string
	Status         *string
	Catatan        *string
	KategoriTarget *[]string
	GenderTarget   *string // ← BARU
	SendReminder   *bool
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
	addStr("tanggal", p.Tanggal)
	addStr("hari", p.Hari)
	addStr("jam", p.Jam)
	addStr("jam_start", p.JamStart)
	addStr("group_id", p.GroupID)
	addStr("acara", p.Acara)
	addStr("materi", p.Materi)
	addStr("status", p.Status)
	addStr("catatan", p.Catatan)
	if p.GenderTarget != nil {
		q += `, gender_target = $` + itoa(n)
		args = append(args, *p.GenderTarget)
		n++
	}

	if p.KategoriTarget != nil {
		q += `, kategori_target = $` + itoa(n) + `::jsonb`
		args = append(args, marshalJSONArray(*p.KategoriTarget))
		n++
	}

	if p.SendReminder != nil {
		q += `, send_reminder = $` + itoa(n)
		args = append(args, *p.SendReminder)
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

// DeleteMany: hapus multiple meetings dalam transaction.
// Return: jumlah meeting terhapus, error.
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

/* ===== Reminder WA (Fonnte) ===== */

type ReminderMeetingRow struct {
	MeetingID string
	Acara     string
	Tanggal   string
	Jam       string
}

// FindPendingReminder — meeting dalam window [from, to] yang reminder-nya belum dikirim.
// Jam di-parse: kalau format HH:MM pakai itu, kalau bukan (mis. "Isya di tempat")
// fallback ke 19:00.
func (r *MeetingRepo) FindPendingReminder(ctx context.Context, from, to time.Time) ([]ReminderMeetingRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT meeting_id, COALESCE(acara, ''),
		       TO_CHAR(tanggal, 'YYYY-MM-DD'),
		       COALESCE(jam, '')
		FROM meetings
		WHERE reminder_sent_at IS NULL
		  AND status != 'LIBUR'
		  AND send_reminder = true
		  AND (
		    (tanggal::date + CASE
		       WHEN jam ~ '^[0-9]{1,2}:[0-9]{2}' THEN jam::time
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
		if err := rows.Scan(&row.MeetingID, &row.Acara, &row.Tanggal, &row.Jam); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// MarkReminderSent — tandai meeting sudah di-reminder, biar tidak dobel.
func (r *MeetingRepo) MarkReminderSent(ctx context.Context, meetingID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE meetings SET reminder_sent_at = now() WHERE meeting_id = $1`,
		meetingID,
	)
	return err
}
