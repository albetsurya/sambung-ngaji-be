package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type MonitoringRepo struct {
	pool *pgxpool.Pool
}

func NewMonitoringRepo(pool *pgxpool.Pool) *MonitoringRepo {
	return &MonitoringRepo{pool: pool}
}

const monitoringSelectCols = `
	monitoring_id, member_id, date, type, status, notes,
	follow_up, created_by, created_at, updated_at`

func (r *MonitoringRepo) FindByMember(ctx context.Context, memberID string) ([]model.Monitoring, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+monitoringSelectCols+` FROM monitoring WHERE member_id=$1 ORDER BY date DESC`,
		memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitorings(rows)
}

func scanMonitorings(rows rowsScanner) ([]model.Monitoring, error) {
	var out []model.Monitoring
	for rows.Next() {
		var m model.Monitoring
		err := rows.Scan(
			&m.MonitoringID, &m.MemberID, &m.Date, &m.Type, &m.Status, &m.Notes,
			&m.FollowUp, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
