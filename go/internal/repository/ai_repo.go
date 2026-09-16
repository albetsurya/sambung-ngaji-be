package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type AIRepo struct {
	pool *pgxpool.Pool
}

func NewAIRepo(pool *pgxpool.Pool) *AIRepo {
	return &AIRepo{pool: pool}
}

// ===== Quota =====

func (r *AIRepo) GetQuota(ctx context.Context, userID string, date time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(count, 0) FROM ai_quota WHERE user_id = $1 AND date = $2
	`, userID, date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (r *AIRepo) IncrementQuota(ctx context.Context, userID string, date time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_quota (user_id, date, count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, date) DO UPDATE SET count = ai_quota.count + 1
	`, userID, date.Format("2006-01-02"))
	return err
}

// ===== Usage =====

func (r *AIRepo) InsertUsage(ctx context.Context, u *model.AIUsageLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_usage
		(usage_id, user_id, user_nama, role, provider,
		 input_tokens, output_tokens, total_tokens, timestamp)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now())
	`, u.UsageID, u.UserID, u.UserNama, u.Role, u.Provider,
		u.InputTokens, u.OutputTokens, u.TotalTokens)
	return err
}

func (r *AIRepo) GetUsageStats(ctx context.Context) ([]model.AIUsageLog, error) {
	monthStart := time.Now().AddDate(0, 0, -30)
	rows, err := r.pool.Query(ctx, `
		SELECT usage_id, user_id, user_nama, role, provider,
		       input_tokens, output_tokens, total_tokens, timestamp
		FROM ai_usage
		WHERE timestamp >= $1
		ORDER BY timestamp DESC
	`, monthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.AIUsageLog
	for rows.Next() {
		var u model.AIUsageLog
		if err := rows.Scan(
			&u.UsageID, &u.UserID, &u.UserNama, &u.Role, &u.Provider,
			&u.InputTokens, &u.OutputTokens, &u.TotalTokens, &u.Timestamp,
		); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
