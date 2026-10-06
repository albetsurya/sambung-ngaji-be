package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

// TilawatiTimelineEdit merepresentasikan entri di tabel tilawati_timeline_edits
type TilawatiTimelineEdit struct {
	Jilid                int            `json:"jilid"`
	Page                 int            `json:"page"`
	PublishedTimelineJSON util.JSONB     `json:"publishedTimelineJson"` // Store as JSONB
	Revision             string         `json:"revision"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
}

// PublishedTimeline merefleksikan struktur data dari frontend
type PublishedTimeline struct {
	Clips []struct {
		ID        string `json:"id"`
		OutputURL string `json:"outputUrl"` // URL setelah di-publish
	} `json:"clips"`
	Texts []interface{} `json:"texts"` // Asumsikan ini adalah array of objects
}

type AudioAsset struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	AudioURL  string `json:"audioUrl"`
	Duration  float64 `json:"duration"`
	IsOpening bool   `json:"isOpening"` // true for header, false for isi
}

type TimelineClip struct {
	ID             string    `json:"id"`
	SourceAudioURL string    `json:"sourceAudioUrl"`
	Segments       []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"segments"`
	TrimStart float64 `json:"trimStart"`
	TrimEnd   float64 `json:"trimEnd"`
}

type PublishTimelineRequest struct {
	Jilid            int               `json:"jilid"`
	Page             int               `json:"page"`
	Timeline         PublishedTimeline `json:"timeline"`
	ExpectedRevision string            `json:"expectedRevision"`
}

type TilawatiEditorService struct {
	db          *pgxpool.Pool
	repo        *repository.Repository
	storageSvc  *StorageService
	audioRoot   string // Path ke public/audio/tilawati
	frontendDir string // Path ke frontend directory
}

func NewTilawatiEditorService(db *pgxpool.Pool, repo *repository.Repository, storageSvc *StorageService, frontendDir string) *TilawatiEditorService {
	return &TilawatiEditorService{
		db:          db,
		repo:        repo,
		storageSvc:  storageSvc,
		audioRoot:   filepath.Join(frontendDir, "public", "audio", "tilawati"),
		frontendDir: frontendDir,
	}
}

func (s *TilawatiEditorService) GetPublishedTimeline(ctx context.Context, jilid, page int) (*TilawatiTimelineEdit, error) {
	query := `SELECT jilid, page, published_timeline_json, revision, created_at, updated_at FROM tilawati_timeline_edits WHERE jilid = $1 AND page = $2`
	edit := TilawatiTimelineEdit{}
	err := s.db.QueryRow(ctx, query, jilid, page).Scan(&edit.Jilid, &edit.Page, &edit.PublishedTimelineJSON, &edit.Revision, &edit.CreatedAt, &edit.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil // Tidak ada data yang dipublikasikan
		}
		return nil, fmt.Errorf("failed to get published timeline from DB: %w", err)
	}
	return &edit, nil
}

func (s *TilawatiEditorService) PublishTimeline(ctx context.Context, req PublishTimelineRequest) (*TilawatiTimelineEdit, []AudioAsset, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// Check revision (optimistic locking)
	if req.ExpectedRevision != "" {
		currentRevisionQuery := `SELECT revision FROM tilawati_timeline_edits WHERE jilid = $1 AND page = $2`
		var currentRevision string
		err = s.db.QueryRow(ctx, currentRevisionQuery, req.Jilid, req.Page).Scan(&currentRevision)
		if err == nil && currentRevision != req.ExpectedRevision {
			return nil, nil, fmt.Errorf("revision mismatch: expected %s, got %s", req.ExpectedRevision, currentRevision)
		} else if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil, fmt.Errorf("failed to check current revision: %w", err)
		}
	}

	pageDirectory := filepath.Join(s.audioRoot, fmt.Sprintf("jilid_%d", req.Jilid), fmt.Sprintf("hal%d", req.Page))
	if err := os.MkdirAll(pageDirectory, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create page directory: %w", err)
	}

	newRevision := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	outputByClip := make(map[string]string)
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Timeline.Clips))

	// Hapus file lama di Supabase Storage untuk jilid dan halaman ini (opsional, untuk clean up)
	// Atau hapus hanya file yang tidak ada di `req.Timeline.Clips` yang baru
	// Untuk simplicity, kita akan upload yang baru, dan yang lama mungkin akan tertimpa/tetap ada jika namanya berbeda

	for _, rawClip := range req.Timeline.Clips {
		clip := rawClip // make a copy for the goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			segments := s.flattenTimelineSegments(clip)
			if len(segments) == 0 {
				log.Debug().Msgf("Clip %s has no segments, skipping audio processing.", clip.ID)
				return
			}

			safeID := util.Slugify(clip.ID) // Pastikan ID aman untuk nama file
			outputFilename := fmt.Sprintf("editor_%s_%s.ogg", newRevision, safeID)
			localOutputPath := filepath.Join(pageDirectory, outputFilename)
			
			log.Info().Msgf("Processing clip %s for Jilid %d Hal %d. Local output: %s", clip.ID, req.Jilid, req.Page, localOutputPath)

			err := s.createMergedAudio(segments, localOutputPath)
			if err != nil {
				errChan <- fmt.Errorf("failed to create merged audio for clip %s: %w", clip.ID, err)
				return
			}
			log.Info().Msgf("Successfully created merged audio for clip %s locally.", clip.ID)

			// Upload ke Supabase Storage
			supabasePath := fmt.Sprintf("tilawati/jilid_%d/hal%d/%s", req.Jilid, req.Page, outputFilename)
			uploadedURL, err := s.storageSvc.UploadFile(ctx, supabasePath, localOutputPath, "audio/ogg")
			if err != nil {
				errChan <- fmt.Errorf("failed to upload audio for clip %s to Supabase: %w", clip.ID, err)
				return
			}
			outputByClip[clip.ID] = uploadedURL // Simpan URL yang di-upload
			log.Info().Msgf("Successfully uploaded clip %s to Supabase. URL: %s", clip.ID, uploadedURL)

			// Hapus file lokal setelah diupload
			if err := os.Remove(localOutputPath); err != nil {
				log.Warn().Err(err).Msgf("Failed to remove local audio file %s", localOutputPath)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Periksa apakah ada error dari goroutine
	for err := range errChan {
		return nil, nil, err
	}

	// Perbarui URL output di timeline yang dipublikasikan
	for i, clip := range req.Timeline.Clips {
		if url, ok := outputByClip[clip.ID]; ok {
			req.Timeline.Clips[i].OutputURL = url
		} else {
			req.Timeline.Clips[i].OutputURL = "" // Jika tidak ada URL yang dihasilkan/di-upload
		}
	}

	// Konversi PublishedTimeline ke JSONB
	publishedJSON, err := json.Marshal(req.Timeline)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal published timeline to JSON: %w", err)
	}

	upsertQuery := `
		INSERT INTO tilawati_timeline_edits (jilid, page, published_timeline_json, revision)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (jilid, page) DO UPDATE
		SET published_timeline_json = EXCLUDED.published_timeline_json,
			revision = EXCLUDED.revision,
			updated_at = NOW()
		RETURNING jilid, page, published_timeline_json, revision, created_at, updated_at
	`
	edit := TilawatiTimelineEdit{}
	err = tx.QueryRow(ctx, upsertQuery, req.Jilid, req.Page, publishedJSON, newRevision).Scan(
		&edit.Jilid, &edit.Page, &edit.PublishedTimelineJSON, &edit.Revision, &edit.CreatedAt, &edit.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert tilawati timeline edit: %w", err)
	}

	// List audio assets (current state)
	currentAssets, err := s.ListPageAudio(req.Jilid, req.Page)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list page audio after publish, returning empty list")
		currentAssets = []AudioAsset{}
	}

	return &edit, currentAssets, nil
}

// createMergedAudio menggabungkan segmen-segmen audio menggunakan FFmpeg
func (s *TilawatiEditorService) createMergedAudio(segments []struct { Start float64; End float64 }, outputPath string) error {
	tempDir, err := os.MkdirTemp("", "tilawati-merge-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	clipFiles := []string{}
	for i, seg := range segments {
		sourcePath := s.audioUrlToFile(seg.SourceAudioURL)
		if !strings.HasPrefix(sourcePath, s.audioRoot) {
            // Ini untuk memastikan path aman dan berasal dari direktori audio aplikasi
            // Jika audio master Anda di luar public/audio/tilawati, Anda mungkin perlu menyesuaikan ini.
            return fmt.Errorf("invalid source audio URL: %s is outside expected audio root", seg.SourceAudioURL)
        }

		clipOutputPath := filepath.Join(tempDir, fmt.Sprintf("clip-%d.ogg", i))
		
		cmdArgs := []string{
			"-ss", fmt.Sprintf("%f", seg.Start),
			"-i", sourcePath,
			"-t", fmt.Sprintf("%f", seg.End-seg.Start),
			"-vn",
			"-c:a", "libvorbis",
			"-q:a", "5",
			clipOutputPath,
		}
		cmd := exec.Command("ffmpeg", cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Str("ffmpeg_output", string(output)).Err(err).Msg("FFmpeg clip creation failed")
			return fmt.Errorf("failed to create audio clip %d: %w, output: %s", i, err, string(output))
		}
		clipFiles = append(clipFiles, clipOutputPath)
	}

	if len(clipFiles) == 0 {
		return fmt.Errorf("no audio clips to merge")
	}

	concatListPath := filepath.Join(tempDir, "concat.txt")
	concatContent := ""
	for _, file := range clipFiles {
		concatContent += fmt.Sprintf("file '%s'
", strings.ReplaceAll(file, "'", "'\''"))
	}
	err = os.WriteFile(concatListPath, []byte(concatContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write concat list: %w", err)
	}

	finalCmdArgs := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatListPath,
		"-c:a", "libvorbis",
		"-q:a", "5",
		outputPath,
	}
	finalCmd := exec.Command("ffmpeg", finalCmdArgs...)
	output, err := finalCmd.CombinedOutput()
	if err != nil {
		log.Error().Str("ffmpeg_output", string(output)).Err(err).Msg("FFmpeg merge failed")
		return fmt.Errorf("failed to merge audio clips: %w, output: %s", err, string(output))
	}

	return nil
}

// audioUrlToFile mengkonversi URL audio menjadi path file lokal yang bisa diakses FFmpeg
func (s *TilawatiEditorService) audioUrlToFile(audioUrl string) string {
	// Asumsi audioUrl adalah path relatif dari /public/audio/tilawati
	// Contoh: /audio/tilawati/jilid_1/hal1/hal1_isi_01.ogg
	// Kita perlu mengubahnya menjadi path absolut di server
	if strings.HasPrefix(audioUrl, "/audio/tilawati/") {
		return filepath.Join(s.frontendDir, "public", audioUrl)
	}
	// Fallback, mungkin ini adalah path absolut atau URL lain, sesuaikan jika perlu
	return audioUrl
}

// flattenTimelineSegments mengambil clip dari frontend dan "merata"kannya menjadi segmen audio mentah
// Ini adalah duplikasi logic dari frontend/vite.config.ts
func (s *TilawatiEditorService) flattenTimelineSegments(clip TimelineClip) []struct { SourceAudioURL string; Start float64; End float64 } {
	trimStart := clip.TrimStart
	trimEnd := clip.TrimEnd
	var offset float64 = 0
	output := []struct { SourceAudioURL string; Start float64; End float64 }{}

	for _, rawSegment := range clip.Segments {
		segmentStart := rawSegment.Start
		segmentEnd := rawSegment.End
		length := segmentEnd - segmentStart
		
		visibleStart := math.Max(trimStart, offset)
		visibleEnd := math.Min(trimEnd, offset + length)

		if visibleEnd > visibleStart {
			output = append(output, struct { SourceAudioURL string; Start float64; End float64 }{
				SourceAudioURL: clip.SourceAudioURL, // Source audio asli untuk clip ini
				Start:          segmentStart + (visibleStart - offset),
				End:            segmentStart + (visibleEnd - offset),
			})
		}
		offset += length
	}
	return output
}

// ListPageAudio membaca file audio dari direktori public/audio/tilawati/jilid_X/halY
// Ini adalah duplikasi logic dari frontend/vite.config.ts
func (s *TilawatiEditorService) ListPageAudio(jilid int, page int) ([]AudioAsset, error) {
	pageAudioPath := filepath.Join(s.audioRoot, fmt.Sprintf("jilid_%d", jilid), fmt.Sprintf("hal%d", page))
	files, err := os.ReadDir(pageAudioPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []AudioAsset{}, nil // Directory doesn't exist, no audio files
		}
		return nil, fmt.Errorf("failed to read audio directory %s: %w", pageAudioPath, err)
	}

	assets := []AudioAsset{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".ogg") {
			continue
		}

		fullPath := filepath.Join(pageAudioPath, file.Name())
		audioURL := fmt.Sprintf("/audio/tilawati/jilid_%d/hal%d/%s", jilid, page, file.Name())
		
		// Dapatkan durasi audio menggunakan ffprobe
		duration, err := s.getAudioDuration(fullPath)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to get duration for %s, setting to 0", file.Name())
			duration = 0
		}

		asset := AudioAsset{
			ID:        strings.TrimSuffix(file.Name(), ".ogg"),
			Filename:  file.Name(),
			AudioURL:  audioURL,
			Duration:  duration,
			IsOpening: strings.Contains(file.Name(), "header_"), // Simple heuristic
		}
		assets = append(assets, asset)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Filename < assets[j].Filename
	})

	return assets, nil
}

// getAudioDuration menggunakan ffprobe untuk mendapatkan durasi audio
func (s *TilawatiEditorService) getAudioDuration(filePath string) (float64, error) {
    cmd := exec.Command("ffprobe",
        "-v", "error",
        "-show_entries", "format=duration",
        "-of", "default=noprint_wrappers=1:nokey=1",
        filePath,
    )
    output, err := cmd.CombinedOutput()
    if err != nil {
        return 0, fmt.Errorf("ffprobe failed for %s: %w, output: %s", filePath, err, string(output))
    }

    durationStr := strings.TrimSpace(string(output))
    duration, err := strconv.ParseFloat(durationStr, 64)
    if err != nil {
        return 0, fmt.Errorf("failed to parse duration %s: %w", durationStr, err)
    }
    return duration, nil
}

// Pastikan import math ada di awal file (ditambahkan secara manual jika tidak ada)
import "math"
