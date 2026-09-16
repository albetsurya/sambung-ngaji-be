package config

import (
	"fmt"
	"os"
	"strconv"

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
}

func Load() (*Config, error) {
	// .env opsional — di Fly.io pakai secrets
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("APP_PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTExpiryHrs: getEnvInt("JWT_EXPIRY_HOURS", 12),
	}

	cfg.AIProvider = getEnv("AI_PROVIDER", "omniroute")
	cfg.OmniRouteEndpoint = os.Getenv("OMNIROUTE_ENDPOINT")
	cfg.OmniRouteAPIKey = os.Getenv("OMNIROUTE_API_KEY")
	cfg.OmniRouteModel = getEnv("OMNIROUTE_MODEL", "auto/best-vision")
	cfg.OmniRouteTimeout = getEnvInt("OMNIROUTE_TIMEOUT_SEC", 60)

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}

	if cfg.AppEnv == "production" && cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi di production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
