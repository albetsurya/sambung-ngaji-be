package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/ai"
	"pengajian-backend/internal/api"
	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/config"
	"pengajian-backend/internal/database"
	"pengajian-backend/internal/handler"
)

func main() {
	setupLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("gagal load config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal koneksi database")
	}
	defer db.Close()
	log.Info().Msg("database connected")

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHrs)
	authSvc := auth.NewService(db, jwtMgr)

	omniRoute := ai.NewOmniRoute(ai.OmniRouteConfig{
		Endpoint: cfg.OmniRouteEndpoint,
		APIKey:   cfg.OmniRouteAPIKey,
		Model:    cfg.OmniRouteModel,
		Timeout:  time.Duration(cfg.OmniRouteTimeout) * time.Second,
	})

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

	handler.RegisterHealth(app, db)
	services := api.NewServices(db, authSvc, omniRoute)
	api.RegisterAPI(app, services)

	log.Info().Strs("actions", api.ListRegisteredActions()).Msg("actions terdaftar")

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

	log.Info().Msg("shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Error().Err(err).Msg("shutdown error")
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
