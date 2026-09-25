package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
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
	LocalsBody      = "body"
	LocalsUser      = "user"
	LocalsClaims    = "claims"
	LocalsRequestID = "request_id"
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
	// Fitur publik tanpa login (jadwal umum, info petugas).
	"getMeetings":        true,
	"getFridaySchedules": true,
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
// Token diambil dari header Authorization, fallback ke HttpOnly cookie.
// Kalau token berasal dari cookie (bukan header), origin request wajib
// ada di allowlist CORS — proteksi CSRF untuk cookie auth.
func AuthMiddleware(svc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := BodyOf(c)
		action, _ := body["action"].(string)
		if action == "" || PublicActions[action] {
			return c.Next()
		}

		token := c.Get("Authorization")
		fromCookie := false
		if token == "" {
			token = c.Cookies(sessionCookieName)
			fromCookie = token != ""
		}
		if token == "" {
			return Fail(c, "Unauthorized: token tidak ada")
		}
		token = strings.TrimPrefix(token, "Bearer ")

		if fromCookie && !originAllowed(c.Get("Origin")) {
			return Fail(c, "Unauthorized: origin tidak diizinkan")
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

		// Visibilitas Data berbasis Group ID:
		// 1. SUPER_ADMIN -> Akses Global. Tidak di-override; jika SUPER_ADMIN kirim group_id di body maka ter-filter, jika tidak maka lihat semua data.
		// 2. Role selain SUPER_ADMIN -> Hanya bisa melihat data dalam 1 kelompok yang ditugaskan. group_id di body dipaksa menjadi u.GroupID.
		if u.Role != "SUPER_ADMIN" {
			body := BodyOf(c)
			if u.GroupID != nil && *u.GroupID != "" {
				body["group_id"] = *u.GroupID
			} else {
				// User non-SUPER_ADMIN yang belum punya group_id tidak boleh melihat data kelompok lain
				body["group_id"] = "__UNASSIGNED_GROUP__"
			}
		}

		return c.Next()
	}
}

// originAllowed: cek Origin terhadap allowlist CORS yang sama dengan main.go.
// Origin kosong diizinkan (request same-origin/non-browser tidak kirim Origin).
func originAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	allowed := os.Getenv("CORS_ORIGINS")
	if allowed == "" {
		allowed = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174"
	}
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
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

// RateLimiterStore defines the interface for rate limiter storage backends.
// Implement this to swap in-memory store with Redis, etc.
type RateLimiterStore interface {
	CheckAndInc(key string, window time.Duration, maxRequests int) (bool, error)
}

// InMemoryStore is a RateLimiterStore implementation using an in-memory map.
// Suitable for single-instance deployments. Use Redis for multi-instance.
type InMemoryStore struct {
	mu      sync.Mutex
	clients map[string]*clientData
}

type clientData struct {
	count   int
	resetAt time.Time
}

// NewInMemoryStore creates a new InMemoryStore
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		clients: make(map[string]*clientData),
	}
}

func (s *InMemoryStore) CheckAndInc(key string, window time.Duration, maxRequests int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	data, exists := s.clients[key]
	if !exists || now.After(data.resetAt) {
		s.clients[key] = &clientData{
			count:   1,
			resetAt: now.Add(window),
		}
		return true, nil
	}

	data.count++
	if data.count > maxRequests {
		return false, nil
	}
	return true, nil
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

// LoginRateLimiterMiddleware — throttle ketat khusus action login.
// Kunci: IP + username (lowercase, max 64 char) agar brute-force per akun
// tetap kena throttle walau attacker rotasi IP. Fail-closed: saat store
// error, login ditolak sementara (aman) alih-alih diloloskan.
func LoginRateLimiterMiddleware() fiber.Handler {
	store := NewInMemoryStore()
	const maxAttempts = 10
	const window = time.Minute

	return func(c *fiber.Ctx) error {
		if BodyString(c, "action") != "login" {
			return c.Next()
		}
		username := BodyString(c, "username")
		if len(username) > 64 {
			username = username[:64]
		}
		key := "login:" + c.IP() + ":" + strings.ToLower(strings.TrimSpace(username))

		allowed, err := store.CheckAndInc(key, window, maxAttempts)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"message": "Layanan sibuk, coba lagi sebentar",
			})
		}
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"message": "Terlalu banyak percobaan login. Coba lagi semenit lagi.",
			})
		}
		return c.Next()
	}
}
func RateLimiterMiddleware(config RateLimiterConfig, store RateLimiterStore) fiber.Handler {
	if store == nil {
		store = NewInMemoryStore()
	}

	return func(c *fiber.Ctx) error {
		key := config.KeyFunc(c)

		allowed, err := store.CheckAndInc(key, config.Window, config.MaxRequests)
		if err != nil {
			// Jika store error, allow request (fail open)
			return c.Next()
		}

		if !allowed {
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
