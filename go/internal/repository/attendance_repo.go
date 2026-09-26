package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type AttendanceRepo struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepo(pool *pgxpool.Pool) *AttendanceRepo {
	return &AttendanceRepo{pool: pool}
}

const attendanceSelectCols = `
	attendance_id, meeting_id, member_id, status, catatan,
	created_by, created_at, updated_at`

func (r *AttendanceRepo) FindByMeeting(ctx context.Context, meetingID string) ([]model.Attendance, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+attendanceSelectCols+` FROM attendance WHERE meeting_id = $1`, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendances(rows)
}

func (r *AttendanceRepo) FindByMember(ctx context.Context, memberID string) ([]model.Attendance, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+attendanceSelectCols+` FROM attendance WHERE member_id = $1 ORDER BY created_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendances(rows)
}

func (r *AttendanceRepo) FindByMeetingAndMember(ctx context.Context, meetingID, memberID string) (*model.Attendance, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+attendanceSelectCols+` FROM attendance WHERE meeting_id = $1 AND member_id = $2`,
		meetingID, memberID)
	return scanAttendance(row)
}

func scanAttendance(s rowScanner) (*model.Attendance, error) {
	var a model.Attendance
	err := s.Scan(
		&a.AttendanceID, &a.MeetingID, &a.MemberID, &a.Status, &a.Catatan,
		&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func scanAttendances(rows rowsScanner) ([]model.Attendance, error) {
	var out []model.Attendance
	for rows.Next() {
		var a model.Attendance
		err := rows.Scan(
			&a.AttendanceID, &a.MeetingID, &a.MemberID, &a.Status, &a.Catatan,
			&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AttendanceRepo) Insert(ctx context.Context, a *model.Attendance) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attendance
		(attendance_id, meeting_id, member_id, status, catatan, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,now(),now())
	`, a.AttendanceID, a.MeetingID, a.MemberID, a.Status, a.Catatan, a.CreatedBy)
	return err
}

func (r *AttendanceRepo) Update(ctx context.Context, id, status, catatan string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attendance SET status=$1, catatan=$2, updated_at=now()
		WHERE attendance_id=$3
	`, status, catatan, id)
	return err
}

func (r *AttendanceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM attendance WHERE attendance_id=$1`, id)
	return err
}

func (r *AttendanceRepo) DeleteByMeeting(ctx context.Context, meetingID string) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM attendance WHERE meeting_id=$1`, meetingID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *AttendanceRepo) DeleteByMember(ctx context.Context, memberID string) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM attendance WHERE member_id=$1`, memberID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type BulkItem struct {
	MemberID string
	Status   string
	Catatan  string
}

func (r *AttendanceRepo) BulkUpsert(ctx context.Context, meetingID string, items []BulkItem, createdBy string) (inserted, updated int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, item := range items {
		if item.MemberID == "" || item.Status == "" {
			continue
		}
		var existingID string
		err := tx.QueryRow(ctx,
			`SELECT attendance_id FROM attendance WHERE meeting_id=$1 AND member_id=$2`,
			meetingID, item.MemberID,
		).Scan(&existingID)
		if err == nil {
			_, err = tx.Exec(ctx, `
				UPDATE attendance SET status=$1, catatan=$2, updated_at=now()
				WHERE attendance_id=$3
			`, item.Status, item.Catatan, existingID)
			if err != nil {
				return 0, 0, err
			}
			updated++
		} else {
			newID := generateAttendanceID()
			_, err = tx.Exec(ctx, `
				INSERT INTO attendance
				(attendance_id, meeting_id, member_id, status, catatan, created_by, created_at, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,now(),now())
			`, newID, meetingID, item.MemberID, item.Status, item.Catatan, createdBy)
			if err != nil {
				return 0, 0, err
			}
			inserted++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return inserted, updated, nil
}

func generateAttendanceID() string {
	return newIDWithPrefix("ATD")
}

func (r *AttendanceRepo) FindByMeetingIDs(ctx context.Context, meetingIDs []string) ([]model.Attendance, error) {
	if len(meetingIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+attendanceSelectCols+` FROM attendance WHERE meeting_id = ANY($1)`,
		meetingIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendances(rows)
}

func (r *AttendanceRepo) FindByMemberIDs(ctx context.Context, memberIDs []string) ([]model.Attendance, error) {
	if len(memberIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+attendanceSelectCols+` FROM attendance
		 WHERE member_id = ANY($1)
		 ORDER BY created_at DESC`,
		memberIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendances(rows)
}
