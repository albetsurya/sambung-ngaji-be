package repository

import (
	"context"

	"pengajian-backend/internal/model"
)

func (r *MonitoringRepo) Insert(ctx context.Context, m *model.Monitoring) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO monitoring
		(monitoring_id, member_id, date, type, status, notes,
		 follow_up, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now(),now())
	`, m.MonitoringID, m.MemberID, m.Date, m.Type, m.Status,
		m.Notes, m.FollowUp, m.CreatedBy)
	return err
}

func (r *MonitoringRepo) FindByID(ctx context.Context, id string) (*model.Monitoring, error) {
	var m model.Monitoring
	err := r.pool.QueryRow(ctx, `
		SELECT monitoring_id, member_id, date, type, status, notes,
		       follow_up, created_by, created_at, updated_at
		FROM monitoring WHERE monitoring_id = $1
	`, id).Scan(
		&m.MonitoringID, &m.MemberID, &m.Date, &m.Type, &m.Status,
		&m.Notes, &m.FollowUp, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MonitoringRepo) Update(ctx context.Context, id string, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	q := `UPDATE monitoring SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE monitoring_id = $` + itoa(n)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
