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
	// DB Pool
	DBMaxConns          int
	DBMinConns          int
	DBMaxConnLifetime   string
	DBMaxConnIdleTime   string
	DBHealthCheckPeriod string
	// AI (Gemini utama, Groq + Nvidia fallback — ketiganya gratis)
	// Tiap provider coba daftar model berurutan sampai ada yang sukses.
	AIProvider   string
	GeminiAPIKey string
	GeminiModels []string
	GroqAPIKey   string
	GroqModels   []string
	NvidiaAPIKey string
	NvidiaModels []string

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
		// DB Pool defaults
		DBMaxConns:          getEnvInt("DB_MAX_CONNS", 10),
		DBMinConns:          getEnvInt("DB_MIN_CONNS", 1),
		DBMaxConnLifetime:   getEnv("DB_MAX_CONN_LIFETIME", "1h"),
		DBMaxConnIdleTime:   getEnv("DB_MAX_CONN_IDLE_TIME", "30m"),
		DBHealthCheckPeriod: getEnv("DB_HEALTH_CHECK_PERIOD", "30s"),
	}

	cfg.AIProvider = getEnv("AI_PROVIDER", "gemini")
	cfg.GeminiAPIKey = envTrim("GEMINI_API_KEY")
	cfg.GeminiModels = getEnvList("GEMINI_MODELS", "GEMINI_MODEL", "gemini-3.8-flash,gemini-3.7-flash,gemini-3.6-flash,gemini-3.5-flash,gemini-2.5-pro")
	cfg.GroqAPIKey = envTrim("GROQ_API_KEY")
	cfg.GroqModels = getEnvList("GROQ_MODELS", "GROQ_MODEL", "openai/gpt-oss-120b,openai/gpt-oss-20b,qwen/qwen3.8-27b")
	cfg.NvidiaAPIKey = envTrim("NVIDIA_API_KEY")
	cfg.NvidiaModels = getEnvList("NVIDIA_MODELS", "NVIDIA_MODEL", "openai/gpt-oss-20b")

	cfg.SupabaseURL = envTrim("SUPABASE_URL")
	cfg.SupabaseServiceKey = envTrim("SUPABASE_SERVICE_ROLE_KEY")
	cfg.SupabaseBucket = envTrim("SUPABASE_BUCKET")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi")
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

// getEnvList: baca daftar model koma-separated.
// Prioritas: LIST_KEY (baru) > SINGLE_KEY (lama, 1 model) > default.
func getEnvList(listKey, singleKey, def string) []string {
	if v := strings.TrimSpace(os.Getenv(listKey)); v != "" {
		return splitList(v)
	}
	if v := strings.TrimSpace(os.Getenv(singleKey)); v != "" {
		return []string{v}
	}
	return splitList(def)
}

func splitList(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
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
