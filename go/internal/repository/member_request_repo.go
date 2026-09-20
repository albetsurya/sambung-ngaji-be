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
	request_id, user_id, nama, status, member_id, reason,
	created_at, reviewed_by, reviewed_at`

func (r *MemberRequestRepo) Insert(ctx context.Context, in *model.MemberRequest) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO member_requests (request_id, user_id, nama, status, created_at)
		 VALUES ($1,$2,$3,'PENDING', now())`,
		in.RequestID, in.UserID, in.Nama)
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

func (r *MemberRequestRepo) FindByStatus(ctx context.Context, status string) ([]model.MemberRequest, error) {
	var rows rowsScanner
	if status == "" {
		rr, err := r.pool.Query(ctx, `SELECT `+memberRequestSelectCols+` FROM member_requests ORDER BY created_at DESC`)
		if err != nil {
			return nil, err
		}
		defer rr.Close()
		rows = rr
	} else {
		rr, err := r.pool.Query(ctx,
			`SELECT `+memberRequestSelectCols+` FROM member_requests WHERE status = $1 ORDER BY created_at DESC`,
			status)
		if err != nil {
			return nil, err
		}
		defer rr.Close()
		rows = rr
	}

	var out []model.MemberRequest
	for rows.Next() {
		var m model.MemberRequest
		if err := rows.Scan(&m.RequestID, &m.UserID, &m.Nama, &m.Status, &m.MemberID, &m.Reason,
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
	if err := row.Scan(&m.RequestID, &m.UserID, &m.Nama, &m.Status, &m.MemberID, &m.Reason,
		&m.CreatedAt, &m.ReviewedBy, &m.ReviewedAt); err != nil {
		return nil, err
	}
	return &m, nil
}
