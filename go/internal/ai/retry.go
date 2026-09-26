package ai

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Jitter        bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     500 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryableError(err) {
			return err
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.BackoffFactor, float64(attempt)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		if cfg.Jitter {
			jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
			delay += jitter
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"temporary failure",
		"503",
		"502",
		"504",
		"429",
	}

	for _, p := range retryablePatterns {
		if contains(errStr, p) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
