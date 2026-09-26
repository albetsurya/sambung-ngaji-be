package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/ai"
	"pengajian-backend/internal/api"
	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/config"
	"pengajian-backend/internal/database"
	"pengajian-backend/internal/handler"
	"pengajian-backend/internal/service"
)

func main() {
	setupLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("gagal load config")
	}

	db, err := database.Connect(context.Background(), cfg.DatabaseURL, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal koneksi database")
	}
	defer db.Close()
	log.Info().Msg("database connected")

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHrs)
	authSvc := auth.NewService(db, jwtMgr)

	providers := map[string]ai.Provider{}
	if cfg.GeminiAPIKey != "" {
		providers["gemini"] = ai.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModels, 60*time.Second)
	}
	if cfg.GroqAPIKey != "" {
		providers["groq"] = ai.NewGroq(cfg.GroqAPIKey, cfg.GroqModels, 60*time.Second)
	}
	if cfg.NvidiaAPIKey != "" {
		providers["nvidia"] = ai.NewNvidia(cfg.NvidiaAPIKey, cfg.NvidiaModels, 60*time.Second)
	}
	providerOrder := []string{"gemini", "groq", "nvidia"}

	app := fiber.New(fiber.Config{
		AppName:      "Pengajian Backend",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    2 * 1024 * 1024,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			log.Error().Err(err).Str("path", c.Path()).Msg("request error")
			msg := "Gagal memproses permintaan"
			if code >= 400 && code < 500 {
				msg = err.Error()
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"message": msg,
			})
		},
	})

	app.Use(helmet.New(helmet.Config{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		HSTSMaxAge:         31536000,
	}))

	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174,https://sambung-ngaji.vercel.app,https://sambung-ngaji-staging.vercel.app,https://sambung-ngaji-stag.vercel.app,https://sambung-ngaji-dev.vercel.app"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization,ngrok-skip-browser-warning",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(api.RequestIDMiddleware())

	app.Use(api.MetricsMiddleware())

	app.Use(api.RateLimiterMiddleware(api.DefaultRateLimiterConfig(), nil))
	app.Use(api.LoginRateLimiterMiddleware())

	app.Use(api.LoggingMiddleware(&log.Logger))

	handler.RegisterHealth(app, db)
	storageSvc := service.NewStorageService(
		cfg.SupabaseURL,
		cfg.SupabaseServiceKey,
		cfg.SupabaseBucket,
	)
	services := api.NewServices(db, authSvc, providers, providerOrder, storageSvc)
	api.RegisterAPI(app, services)

	api.SetCronServices(services)
	app.Post("/cron/reminder", api.HandleCronReminder)

	log.Info().Strs("actions", api.ListRegisteredActions()).Msg("actions terdaftar")

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		if err := services.Reminder.RunOnce(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("reminder startup error")
		}
		if err := services.FridayReminder.RunOnce(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("friday reminder startup error")
		}

		for {
			select {
			case <-shutdownCtx.Done():
				return
			case <-ticker.C:
				if err := services.Reminder.RunOnce(shutdownCtx); err != nil {
					log.Error().Err(err).Msg("reminder tick error")
				}
				if err := services.FridayReminder.RunOnce(shutdownCtx); err != nil {
					log.Error().Err(err).Msg("friday reminder tick error")
				}
			}
		}
	}()

	go func() {
		addr := ":" + cfg.AppPort
		log.Info().Str("addr", addr).Msg("server listening")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received, draining connections...")
	shutdownCancel()

	shutdownTimeout := 15 * time.Second
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	} else {
		log.Info().Msg("graceful shutdown completed")
	}
	log.Info().Msg("bye")
}

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		return
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}
