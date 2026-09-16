package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiryHrs int
	// AI
	AIProvider        string
	OmniRouteEndpoint string
	OmniRouteAPIKey   string
	OmniRouteModel    string
	OmniRouteTimeout  int
	GroqAPIKey        string
	GroqModel         string
	GeminiAPIKey      string
	GeminiModel       string

	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseBucket     string
}

func Load() (*Config, error) {
	// .env opsional — di production pakai secrets platform
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("PORT", getEnv("APP_PORT", "8080")),
		DatabaseURL:  envTrim("DATABASE_URL"),
		JWTSecret:    envTrim("JWT_SECRET"),
		JWTExpiryHrs: getEnvInt("JWT_EXPIRY_HOURS", 12),
	}

	cfg.AIProvider = getEnv("AI_PROVIDER", "omniroute")
	cfg.OmniRouteEndpoint = envTrim("OMNIROUTE_ENDPOINT")
	cfg.OmniRouteAPIKey = envTrim("OMNIROUTE_API_KEY")
	cfg.OmniRouteModel = getEnv("OMNIROUTE_MODEL", "auto/best-vision")
	cfg.OmniRouteTimeout = getEnvInt("OMNIROUTE_TIMEOUT_SEC", 60)
	cfg.GroqAPIKey = envTrim("GROQ_API_KEY")
	cfg.GroqModel = getEnv("GROQ_MODEL", "openai/gpt-oss-120b")
	cfg.GeminiAPIKey = envTrim("GEMINI_API_KEY")
	cfg.GeminiModel = getEnv("GEMINI_MODEL", "gemini-3.6-flash")

	cfg.SupabaseURL = envTrim("SUPABASE_URL")
	cfg.SupabaseServiceKey = envTrim("SUPABASE_SERVICE_ROLE_KEY")
	cfg.SupabaseBucket = envTrim("SUPABASE_BUCKET")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}

	if cfg.AppEnv == "production" && cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi di production")
	}

	return cfg, nil
}

// envTrim: baca env var + trim whitespace (proteksi newline dari paste).
func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
