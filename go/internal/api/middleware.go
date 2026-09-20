package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/model"
)

const (
	LocalsBody       = "body"
	LocalsUser       = "user"
	LocalsClaims     = "claims"
	LocalsRequestID  = "request_id"
)

// Prometheus metrics
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of in-flight HTTP requests",
		},
	)
)

// PublicActions — action yang tidak butuh token.
var PublicActions = map[string]bool{
	"login":                     true,
	"submitPublicRegistration":  true,
	"checkUsernameAvailability": true,
	"health":                    true,
	"ready":                     true,
}

// MetricsMiddleware records Prometheus metrics for HTTP requests
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		err := c.Next()

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())

		return err
	}
}

// RequestIDMiddleware adds a correlation ID to each request for tracing
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()[:8]
		}
		c.Locals(LocalsRequestID, requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

// RequestIDOf returns the correlation ID for the current request
func RequestIDOf(c *fiber.Ctx) string {
	if v := c.Locals(LocalsRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// BodyParserMiddleware baca raw body JSON sekali, simpan di Locals.
func BodyParserMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Body()
		if len(raw) == 0 {
			return c.Next()
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return c.Next()
		}
		c.Locals(LocalsBody, m)
		return c.Next()
	}
}

func BodyOf(c *fiber.Ctx) map[string]interface{} {
	if v := c.Locals(LocalsBody); v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return map[string]interface{}{}
}

func BodyString(c *fiber.Ctx, key string) string {
	if v, ok := BodyOf(c)[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func BodyBool(c *fiber.Ctx, key string) bool {
	if v, ok := BodyOf(c)[key]; ok {
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return t == "true" || t == "1" || t == "TRUE"
		}
	}
	return false
}

func BodyFloat(c *fiber.Ctx, key string) float64 {
	if v, ok := BodyOf(c)[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		case string:
			var f float64
			_, _ = fmtSscan(t, &f)
			return f
		}
	}
	return 0
}

// ValidateBody validates required fields in the request body
func ValidateBody(requiredFields ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := BodyOf(c)
		if len(body) == 0 {
			return Fail(c, "Request body is required")
		}
		
		var missing []string
		for _, field := range requiredFields {
			if _, ok := body[field]; !ok {
				missing = append(missing, field)
			}
		}
		
		if len(missing) > 0 {
			return Fail(c, fmt.Sprintf("Missing required fields: %s", strings.Join(missing, ", ")))
		}
		
		return c.Next()
	}
}

// AuthMiddleware + Permission check.
// Untuk public action: lewat.
// Untuk action lain: JWT valid + role boleh akses.
func AuthMiddleware(svc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := BodyOf(c)
		action, _ := body["action"].(string)
		if action == "" || PublicActions[action] {
			return c.Next()
		}

		token, _ := body["token"].(string)
		if token == "" {
			return Fail(c, "Unauthorized: token tidak ada")
		}

		u, claims, err := svc.ValidateSession(c.Context(), token)
		if err != nil {
			return Fail(c, "Unauthorized: sesi tidak valid atau kadaluarsa")
		}

		// Super admin only
		if auth.IsSuperAdminOnly(action) && u.Role != "SUPER_ADMIN" {
			return Fail(c, "Forbidden: hanya SUPER_ADMIN")
		}

		// Permission check
		if !auth.CanAccess(u.Role, action) {
			return Fail(c, "Forbidden: role "+u.Role+" tidak memiliki akses ke "+action)
		}

		c.Locals(LocalsUser, u)
		c.Locals(LocalsClaims, claims)
		return c.Next()
	}
}

func UserOf(c *fiber.Ctx) *model.User {
	if v := c.Locals(LocalsUser); v != nil {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}

func ClaimsOf(c *fiber.Ctx) *model.SessionClaims {
	if v := c.Locals(LocalsClaims); v != nil {
		if cl, ok := v.(*model.SessionClaims); ok {
			return cl
		}
	}
	return nil
}

// fmtSscan helper kecil (hindari import fmt di file ini).
func fmtSscan(s string, f *float64) (int, error) {
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
			continue
		}
		if c == '.' {
			break
		}
	}
	*f = float64(n)
	return 1, nil
}

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	MaxRequests int
	Window      time.Duration
	KeyFunc     func(*fiber.Ctx) string
}

// DefaultRateLimiterConfig returns a default rate limiter config
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxRequests: 100,
		Window:      time.Minute,
		KeyFunc: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}
}

// RateLimiterMiddleware returns a rate limiting middleware
func RateLimiterMiddleware(config RateLimiterConfig) fiber.Handler {
	// Simple in-memory store (for production, use Redis)
	type clientData struct {
		count   int
		resetAt time.Time
	}
	
	clients := make(map[string]*clientData)
	
	return func(c *fiber.Ctx) error {
		key := config.KeyFunc(c)
		now := time.Now()
		
		data, exists := clients[key]
		if !exists || now.After(data.resetAt) {
			clients[key] = &clientData{
				count:   1,
				resetAt: now.Add(config.Window),
			}
			return c.Next()
		}
		
		data.count++
		if data.count > config.MaxRequests {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"message": "Rate limit exceeded. Please try again later.",
			})
		}
		
		return c.Next()
	}
}

// LoggingMiddleware logs HTTP requests with correlation ID
func LoggingMiddleware(logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := RequestIDOf(c)
		
		// Log request
		logger.Info().
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Msg("request started")
		
		err := c.Next()
		
		// Log response
		logger.Info().
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Msg("request completed")
		
		return err
	}
}
