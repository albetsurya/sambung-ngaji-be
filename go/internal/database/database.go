package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"pengajian-backend/internal/config"
)

func Connect(ctx context.Context, url string, cfg *config.Config) (*pgxpool.Pool, error) {
	// Wajibkan TLS ke database remote (Supabase direct connection mendukung SSL).
	// Tanpa ini driver default ke prefer dan bisa downgrade ke plaintext.
	// Localhost dikecualikan agar dev lokal (tanpa SSL) tetap jalan.
	if !strings.Contains(url, "sslmode=") &&
		!strings.Contains(url, "@localhost") &&
		!strings.Contains(url, "@127.0.0.1") {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		url += sep + "sslmode=require"
	}

	pgCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Parse duration strings
	maxConnLifetime, _ := time.ParseDuration(cfg.DBMaxConnLifetime)
	maxConnIdleTime, _ := time.ParseDuration(cfg.DBMaxConnIdleTime)
	healthCheckPeriod, _ := time.ParseDuration(cfg.DBHealthCheckPeriod)

	pgCfg.MaxConns = int32(cfg.DBMaxConns)
	pgCfg.MinConns = int32(cfg.DBMinConns)
	pgCfg.MaxConnLifetime = maxConnLifetime
	pgCfg.MaxConnIdleTime = maxConnIdleTime
	pgCfg.HealthCheckPeriod = healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
