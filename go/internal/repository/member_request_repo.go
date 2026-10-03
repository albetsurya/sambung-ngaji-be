package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MemberRequestRepo struct {
	pool *pgxpool.Pool
}

func NewMemberRequestRepo(pool *pgxpool.Pool) *MemberRequestRepo {
	return &MemberRequestRepo{pool: pool}
}

const memberRequestSelectCols = `
	request_id, user_id, name, status, member_id, reason,
	created_at, reviewed_by, reviewed_at`

func (r *MemberRequestRepo) Insert(ctx context.Context, in *model.MemberRequest) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO member_requests (request_id, user_id, name, status, created_at)
		 VALUES ($1,$2,$3,'PENDING', now())`,
		in.RequestID, in.UserID, in.Name)
	return err
}

func (r *MemberRequestRepo) FindByUserID(ctx context.Context, userID string) (*model.MemberRequest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+memberRequestSelectCols+` FROM member_requests WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID)
	return scanMemberRequest(row)
}

func (r *MemberRequestRepo) FindByID(ctx context.Context, requestID string) (*model.MemberRequest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+memberRequestSelectCols+` FROM member_requests WHERE request_id = $1`,
		requestID)
	return scanMemberRequest(row)
}

func (r *MemberRequestRepo) FindByStatus(ctx context.Context, status, groupID string) ([]model.MemberRequest, error) {
	/* Filter group via users.group_id (request milik user ber-group_label).
	   Kolom di-qualify karena user_id ada di kedua tabel. */
	cols := memberRequestSelectCols
	from := `FROM member_requests`
	args := []interface{}{}
	if groupID != "" {
		cols = `member_requests.request_id, member_requests.user_id, member_requests.name, member_requests.status, member_requests.member_id, member_requests.reason, member_requests.created_at, member_requests.reviewed_by, member_requests.reviewed_at`
		from += ` JOIN users u ON u.user_id = member_requests.user_id AND u.group_id = $1`
		args = append(args, groupID)
	}
	var rows rowsScanner
	if status == "" {
		rr, err := r.pool.Query(ctx, `SELECT `+cols+` `+from+` ORDER BY member_requests.created_at DESC`, args...)
		if err != nil {
			return nil, err
		}
		defer rr.Close()
		rows = rr
	} else {
		rr, err := r.pool.Query(ctx,
			`SELECT `+cols+` `+from+` WHERE member_requests.status = $`+itoa(len(args)+1)+` ORDER BY member_requests.created_at DESC`,
			append(args, status)...)
		if err != nil {
			return nil, err
		}
		defer rr.Close()
		rows = rr
	}

	var out []model.MemberRequest
	for rows.Next() {
		var m model.MemberRequest
		if err := rows.Scan(&m.RequestID, &m.UserID, &m.Name, &m.Status, &m.MemberID, &m.Reason,
			&m.CreatedAt, &m.ReviewedBy, &m.ReviewedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRequestRepo) UpdateApproved(ctx context.Context, requestID, memberID, reviewerID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE member_requests SET status='APPROVED', member_id=$2, reviewed_by=$3, reviewed_at=now() WHERE request_id=$1`,
		requestID, memberID, reviewerID)
	return err
}

func (r *MemberRequestRepo) UpdateRejected(ctx context.Context, requestID, reviewerID, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE member_requests SET status='REJECTED', reviewed_by=$2, reviewed_at=now(), reason=$3 WHERE request_id=$1`,
		requestID, reviewerID, reason)
	return err
}

func scanMemberRequest(row rowScanner) (*model.MemberRequest, error) {
	var m model.MemberRequest
	if err := row.Scan(&m.RequestID, &m.UserID, &m.Name, &m.Status, &m.MemberID, &m.Reason,
		&m.CreatedAt, &m.ReviewedBy, &m.ReviewedAt); err != nil {
		return nil, err
	}
	return &m, nil
}
