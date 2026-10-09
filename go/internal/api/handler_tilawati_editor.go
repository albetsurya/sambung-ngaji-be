package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/service"
)

func RegisterTilawatiEditorAPI(app *fiber.App, services *Services, authMiddleware fiber.Handler) {
	// Hanya hasil publish yang publik; draft dan operasi tulis tetap khusus admin.
	app.Get("/api/v1/tilawati-editor/published", func(c *fiber.Ctx) error {
		jilid, _ := strconv.Atoi(c.Query("jilid"))
		page, _ := strconv.Atoi(c.Query("hal"))
		if jilid < 1 || jilid > 6 || page < 1 || page > 44 {
			return Fail(c, "Jilid dan halaman tidak valid")
		}
		edit, err := services.TilawatiEditor.GetPublishedTimeline(c.Context(), jilid, page)
		if err != nil {
			return Fail(c, "Gagal mengambil timeline")
		}
		if edit == nil {
			return Ok(c, fiber.Map{"published": nil})
		}
		var timeline service.PublishedTimeline
		if err := json.Unmarshal(edit.PublishedTimelineJSON, &timeline); err != nil {
			return Fail(c, "Timeline tidak valid")
		}
		return Ok(c, fiber.Map{"published": fiber.Map{"timeline": timeline, "revision": edit.Revision}})
	})
	editor := app.Group("/api/v1/tilawati-editor", authMiddleware)
	editor.Get("/draft", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil || (u.Role != "SUPER_ADMIN" && u.Role != "ADMIN") {
			return c.SendStatus(fiber.StatusForbidden)
		}
		jilid, _ := strconv.Atoi(c.Query("jilid"))
		page, _ := strconv.Atoi(c.Query("hal"))
		if jilid < 1 || jilid > 6 || page < 1 || page > 44 {
			return Fail(c, "Jilid dan halaman tidak valid")
		}
		draft, err := services.TilawatiEditor.GetDraft(c.Context(), u.UserID, jilid, page)
		if err != nil {
			return Fail(c, "Gagal mengambil draft")
		}
		return Ok(c, fiber.Map{"draft": draft})
	})
	editor.Post("/draft-save", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil || (u.Role != "SUPER_ADMIN" && u.Role != "ADMIN") {
			return c.SendStatus(fiber.StatusForbidden)
		}
		var input struct {
			Jilid int `json:"jilid"`
			Page  int `json:"page"`
			service.TimelineDraft
		}
		if err := c.BodyParser(&input); err != nil {
			return Fail(c, "Draft tidak valid")
		}
		if err := services.TilawatiEditor.SaveDraft(c.Context(), u.UserID, input.Jilid, input.Page, input.TimelineDraft); err != nil {
			if strings.Contains(err.Error(), "conflict") {
				return c.Status(fiber.StatusConflict).JSON(Envelope{Success: false, Message: err.Error()})
			}
			return Fail(c, err.Error())
		}
		return Ok(c, fiber.Map{"draft": input.TimelineDraft})
	})

	editor.Get("/list", func(c *fiber.Ctx) error {
		jilidStr := c.Query("jilid")
		pageStr := c.Query("hal")

		jilid, err := strconv.Atoi(jilidStr)
		if err != nil || jilid <= 0 {
			return Fail(c, "Invalid jilid number")
		}
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return Fail(c, "Invalid page number")
		}

		u := UserOf(c)
		if u == nil || (u.Role != "SUPER_ADMIN" && u.Role != "ADMIN") {
			return Fail(c, "Hanya admin yang bisa mengakses fitur ini")
		}

		published, err := services.TilawatiEditor.GetPublishedTimeline(c.Context(), jilid, page)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get published timeline")
			return Fail(c, "Gagal mengambil timeline yang dipublikasikan")
		}

		clips, err := services.TilawatiEditor.ListPageAudio(jilid, page)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list page audio for editor")
			return Fail(c, "Gagal mengambil daftar audio halaman")
		}

		response := fiber.Map{
			"ok":        true,
			"clips":     clips,
			"revision":  "",
			"published": nil,
		}

		if published != nil {
			response["revision"] = published.Revision
			var publishedFrontend service.PublishedTimeline
			if err := json.Unmarshal(published.PublishedTimelineJSON, &publishedFrontend); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal published timeline JSON")
				return Fail(c, "Gagal memproses data timeline")
			}
			response["published"] = fiber.Map{
				"timeline":  publishedFrontend,
				"revision":  published.Revision,
				"questions": nil,
			}
		}
		return Ok(c, response)
	})

	editor.Post("/publish", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil || (u.Role != "SUPER_ADMIN" && u.Role != "ADMIN") {
			return Fail(c, "Hanya admin yang bisa mempublikasikan timeline")
		}

		var req service.PublishTimelineRequest
		if err := c.BodyParser(&req); err != nil {
			return Fail(c, "Invalid request body")
		}

		if req.Jilid <= 0 || req.Page <= 0 {
			return Fail(c, "Jilid dan halaman tidak valid")
		}
		req.PublishedBy = u.UserID

		publishedEdit, updatedAssets, err := services.TilawatiEditor.PublishTimeline(c.Context(), req)
		if err != nil {
			if strings.Contains(err.Error(), "conflict") {
				return c.Status(fiber.StatusConflict).JSON(Envelope{Success: false, Message: err.Error()})
			}
			log.Error().Err(err).Msg("Failed to publish timeline")
			return Fail(c, "Gagal mempublikasikan timeline")
		}

		var publishedFrontend service.PublishedTimeline
		if err := json.Unmarshal(publishedEdit.PublishedTimelineJSON, &publishedFrontend); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal published timeline JSON after publish")
			return Fail(c, "Gagal memproses data timeline yang dipublikasikan")
		}

		return Ok(c, fiber.Map{
			"ok": true,
			"published": fiber.Map{
				"timeline":  publishedFrontend,
				"revision":  publishedEdit.Revision,
				"questions": nil,
			},
			"clips":   updatedAssets,
			"message": "Timeline berhasil dipublikasikan",
		})
	})

	editor.Get("/history", func(c *fiber.Ctx) error {
		u := UserOf(c)
		if u == nil || (u.Role != "SUPER_ADMIN" && u.Role != "ADMIN") {
			return Fail(c, "Hanya admin yang bisa mengakses fitur ini")
		}
		jilid, err := strconv.Atoi(c.Query("jilid"))
		if err != nil || jilid <= 0 {
			return Fail(c, "Invalid jilid number")
		}
		page, err := strconv.Atoi(c.Query("hal"))
		if err != nil || page <= 0 {
			return Fail(c, "Invalid page number")
		}
		history, err := services.TilawatiEditor.ListHistory(c.Context(), jilid, page, 20)
		if err != nil {
			return Fail(c, "Gagal mengambil riwayat timeline")
		}
		return Ok(c, fiber.Map{"history": history})
	})

	editor.Get("/status", func(c *fiber.Ctx) error {
		ffmpegAvailable, ffprobeAvailable := services.TilawatiEditor.AudioToolsAvailable()
		return Ok(c, fiber.Map{
			"ok":      ffmpegAvailable && ffprobeAvailable,
			"ffmpeg":  ffmpegAvailable,
			"ffprobe": ffprobeAvailable,
			"storage": "supabase-storage",
		})
	})
}
