package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

const auditSelectCols = `
	log_id, user_id, user_name, action, target_type, target_id, timestamp`

func (r *AuditRepo) FindAll(ctx context.Context, userID, targetType string, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	q := `SELECT ` + auditSelectCols + ` FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	n := 1
	if userID != "" {
		q += ` AND user_id = $` + itoa(n)
		args = append(args, userID)
		n++
	}
	if targetType != "" {
		q += ` AND target_type = $` + itoa(n)
		args = append(args, targetType)
		n++
	}
	q += ` ORDER BY timestamp DESC LIMIT $` + itoa(n)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.AuditLog
	for rows.Next() {
		var a model.AuditLog
		if err := rows.Scan(
			&a.LogID, &a.UserID, &a.UserName, &a.Action,
			&a.TargetType, &a.TargetID, &a.Timestamp,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AuditRepo) Insert(ctx context.Context, logID string, userID *string, userNama, action, targetType, targetID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (log_id, user_id, user_name, action, target_type, target_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, now())
	`, logID, userID, userNama, action, targetType, targetID)
	return err
}
