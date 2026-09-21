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

	// Use background context for DB connection (no timeout on startup)
	db, err := database.Connect(context.Background(), cfg.DatabaseURL, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal koneksi database")
	}
	defer db.Close()
	log.Info().Msg("database connected")

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHrs)
	authSvc := auth.NewService(db, jwtMgr)

	providers := map[string]ai.Provider{
		"omniroute": ai.NewOmniRoute(
			cfg.OmniRouteEndpoint,
			cfg.OmniRouteAPIKey,
			cfg.OmniRouteModel,
			time.Duration(cfg.OmniRouteTimeout)*time.Second,
		),
	}
	if cfg.GroqAPIKey != "" {
		providers["groq"] = ai.NewGroq(cfg.GroqAPIKey, cfg.GroqModel, 60*time.Second)
	}
	if cfg.GeminiAPIKey != "" {
		providers["gemini"] = ai.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel, 60*time.Second)
	}
	providerOrder := []string{"omniroute", "gemini", "groq"}

	app := fiber.New(fiber.Config{
		AppName:      "Pengajian Backend",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"message": err.Error(),
			})
		},
	})

	// CORS — izinkan semua origin frontend (token dikirim di body, bukan cookie)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization,ngrok-skip-browser-warning",
		MaxAge:       3600,
	}))

	// Request ID middleware for correlation logging
	app.Use(api.RequestIDMiddleware())

	// Prometheus metrics middleware
	app.Use(api.MetricsMiddleware())

	// Rate limiting middleware
	app.Use(api.RateLimiterMiddleware(api.DefaultRateLimiterConfig()))

	// Logging middleware with correlation ID
	app.Use(api.LoggingMiddleware(&log.Logger))

	handler.RegisterHealth(app, db)
	storageSvc := service.NewStorageService(
		cfg.SupabaseURL,
		cfg.SupabaseServiceKey,
		cfg.SupabaseBucket,
	)
	services := api.NewServices(db, authSvc, providers, providerOrder, storageSvc)
	api.RegisterAPI(app, services)

	// WhatsApp webhook (verifikasi + terima notifikasi dari Meta)
	app.Get("/wa/webhook", api.HandleWAWebhookVerify)
	app.Post("/wa/webhook", api.HandleWAWebhookReceive)

	// Cron eksternal (dipanggil cron-job.org tiap 1 jam)
	api.SetCronServices(services)
	app.Post("/cron/reminder", api.HandleCronReminder)

	log.Info().Strs("actions", api.ListRegisteredActions()).Msg("actions terdaftar")

	// Create a context that will be cancelled on shutdown signal
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	// Cron reminder WA — cek tiap jam, kirim H-8 jam sebelum acara
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		if err := services.Reminder.RunOnce(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("reminder startup error")
		}

		for {
			select {
			case <-shutdownCtx.Done():
				return
			case <-ticker.C:
				if err := services.Reminder.RunOnce(shutdownCtx); err != nil {
					log.Error().Err(err).Msg("reminder tick error")
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
	shutdownCancel() // Cancel context for any background work

	// Graceful shutdown with timeout
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
