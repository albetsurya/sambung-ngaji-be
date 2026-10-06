package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/service"
)

func RegisterTilawatiEditorAPI(app *fiber.App, services *Services, authMiddleware fiber.Handler) {
	editor := app.Group("/api/v1/tilawati-editor", authMiddleware)

	editor.Get("/list", func(c *fiber.Ctx) error {
		jilidStr := c.Query("jilid")
		pageStr := c.Query("hal")

		jilid, err := strconv.Atoi(jilidStr)
		if err != nil || jilid <= 0 {
			return SendBadRequest(c, "Invalid jilid number")
		}
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return SendBadRequest(c, "Invalid page number")
		}

		// Pastikan hanya admin yang bisa mengakses ini
		if !auth.IsAdmin(c) {
			return SendUnauthorized(c, "Hanya admin yang bisa mengakses fitur ini")
		}

		published, err := services.TilawatiEditor.GetPublishedTimeline(c.Context(), jilid, page)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get published timeline")
			return SendInternalServerError(c, "Gagal mengambil timeline yang dipublikasikan")
		}

		clips, err := services.TilawatiEditor.ListPageAudio(jilid, page)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list page audio for editor")
			return SendInternalServerError(c, "Gagal mengambil daftar audio halaman")
		}

		response := fiber.Map{
			"ok":        true,
			"clips":     clips,
			"revision":  "",
			"published": nil,
		}

		if published != nil {
			response["revision"] = published.Revision
			// Unmarshal JSONB to the expected frontend structure (PublishedTimeline)
			var publishedFrontend service.PublishedTimeline
			if err := json.Unmarshal(published.PublishedTimelineJSON.Byte, &publishedFrontend); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal published timeline JSONB")
				return SendInternalServerError(c, "Gagal memproses data timeline")
			}
			response["published"] = fiber.Map{
				"timeline":  publishedFrontend,
				"questions": nil, // Frontend akan membuat ini dari timeline
			}
		}
		return SendSuccess(c, response)
	})

	editor.Post("/publish", func(c *fiber.Ctx) error {
		// Pastikan hanya admin yang bisa mengakses ini
		if !auth.IsAdmin(c) {
			return SendUnauthorized(c, "Hanya admin yang bisa mempublikasikan timeline")
		}

		var req service.PublishTimelineRequest
		if err := c.BodyParser(&req); err != nil {
			return SendBadRequest(c, "Invalid request body")
		}

		if req.Jilid <= 0 || req.Page <= 0 {
			return SendBadRequest(c, "Jilid dan halaman tidak valid")
		}

		publishedEdit, updatedAssets, err := services.TilawatiEditor.PublishTimeline(c.Context(), req)
		if err != nil {
			if strings.Contains(err.Error(), "revision mismatch") {
				return SendConflict(c, err.Error()) // Kode 409 Conflict
			}
			log.Error().Err(err).Msg("Failed to publish timeline")
			return SendInternalServerError(c, "Gagal mempublikasikan timeline")
		}

		// Unmarshal JSONB ke struktur PublishedTimeline yang sama dengan yang diterima
		var publishedFrontend service.PublishedTimeline
		if err := json.Unmarshal(publishedEdit.PublishedTimelineJSON.Byte, &publishedFrontend); err != nil {
			log.Error().Err(err).Err(err).Msg("Failed to unmarshal published timeline JSONB after publish")
			return SendInternalServerError(c, "Gagal memproses data timeline yang dipublikasikan")
		}

		return SendSuccess(c, fiber.Map{
			"ok": true,
			"published": fiber.Map{
				"timeline":  publishedFrontend,
				"revision":  publishedEdit.Revision,
				"questions": nil, // Frontend akan membuat ini dari timeline
			},
			"clips":   updatedAssets,
			"message": "Timeline berhasil dipublikasikan",
		})
	})

	// Tambahkan endpoint untuk action=status, jika diperlukan oleh frontend editor
	editor.Get("/status", func(c *fiber.Ctx) error {
		// Ini adalah dummy, karena ffmpeg sudah ada di Dockerfile
		// Nanti bisa tambahkan cek `ffprobe` atau `ffmpeg` versi di sini jika perlu
		return SendSuccess(c, fiber.Map{
			"ok":      true,
			"ffmpeg":  true,
			"storage": "supabase-storage",
		})
	})
}
