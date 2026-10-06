package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// TilawatiTimelineEdit merepresentasikan entri di tabel tilawati_timeline_edits
type TilawatiTimelineEdit struct {
	Jilid                 int             `json:"jilid"`
	Page                  int             `json:"page"`
	PublishedTimelineJSON json.RawMessage `json:"publishedTimelineJson"`
	Revision              string          `json:"revision"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type TimelineClipSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type TimelineClip struct {
	ID             string                `json:"id"`
	SourceAudioURL string                `json:"sourceAudioUrl"`
	Segments       []TimelineClipSegment `json:"segments"`
	TrimStart      float64               `json:"trimStart"`
	TrimEnd        float64               `json:"trimEnd"`
	OutputURL      string                `json:"outputUrl,omitempty"`
}

// PublishedTimeline merefleksikan struktur data dari frontend
type PublishedTimeline struct {
	Clips []TimelineClip `json:"clips"`
	Texts []interface{}  `json:"texts"`
}

type AudioAsset struct {
	ID        string  `json:"id"`
	Filename  string  `json:"filename"`
	AudioURL  string  `json:"audioUrl"`
	Duration  float64 `json:"duration"`
	IsOpening bool    `json:"isOpening"`
}

type PublishTimelineRequest struct {
	Jilid            int               `json:"jilid"`
	Page             int               `json:"page"`
	Timeline         PublishedTimeline `json:"timeline"`
	ExpectedRevision string            `json:"expectedRevision"`
}

type FlattenedSegment struct {
	SourceAudioURL string
	Start          float64
	End            float64
}

type TilawatiEditorService struct {
	db          *pgxpool.Pool
	storageSvc  *StorageService
	audioRoot   string
	frontendDir string
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func NewTilawatiEditorService(db *pgxpool.Pool, storageSvc *StorageService, frontendDir string) *TilawatiEditorService {
	return &TilawatiEditorService{
		db:          db,
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
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get published timeline from DB: %w", err)
	}
	return &edit, nil
}

func (s *TilawatiEditorService) PublishTimeline(ctx context.Context, req PublishTimelineRequest) (*TilawatiTimelineEdit, []AudioAsset, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Optimistic locking check
	if req.ExpectedRevision != "" {
		var currentRevision string
		err := tx.QueryRow(ctx, `SELECT revision FROM tilawati_timeline_edits WHERE jilid = $1 AND page = $2 FOR UPDATE`, req.Jilid, req.Page).Scan(&currentRevision)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil, fmt.Errorf("failed to check current revision: %w", err)
		}
		if currentRevision != "" && currentRevision != req.ExpectedRevision {
			return nil, nil, fmt.Errorf("conflict: timeline has been modified by another user (expected %s, got %s)", req.ExpectedRevision, currentRevision)
		}
	}

	tempDir, err := os.MkdirTemp("", "tilawati-publish-")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp directory for publishing: %w", err)
	}
	defer os.RemoveAll(tempDir)

	pageDirectory := filepath.Join(tempDir, fmt.Sprintf("jilid_%d_hal%d", req.Jilid, req.Page))
	if err := os.MkdirAll(pageDirectory, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create page directory: %w", err)
	}

	newRevision := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	outputByClip := make(map[string]string)
	var outputMutex sync.Mutex

	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Timeline.Clips))

	for _, rawClip := range req.Timeline.Clips {
		clip := rawClip
		wg.Add(1)
		go func() {
			defer wg.Done()
			segments := s.flattenTimelineSegments(clip)
			if len(segments) == 0 {
				log.Debug().Msgf("Clip %s has no segments, skipping audio processing.", clip.ID)
				return
			}

			safeID := slugify(clip.ID)
			outputFilename := fmt.Sprintf("editor_%s_%s.ogg", newRevision, safeID)
			localOutputPath := filepath.Join(pageDirectory, outputFilename)

			log.Info().Msgf("Processing clip %s for Jilid %d Hal %d. Local output: %s", clip.ID, req.Jilid, req.Page, localOutputPath)

			err := s.createMergedAudio(segments, localOutputPath)
			if err != nil {
				errChan <- fmt.Errorf("failed to create merged audio for clip %s: %w", clip.ID, err)
				return
			}
			log.Info().Msgf("Successfully created merged audio for clip %s locally.", clip.ID)

			supabasePath := fmt.Sprintf("tilawati/jilid_%d/hal%d/%s", req.Jilid, req.Page, outputFilename)
			uploadedURL, err := s.storageSvc.UploadFile(ctx, supabasePath, localOutputPath, "audio/ogg")
			if err != nil {
				errChan <- fmt.Errorf("failed to upload audio for clip %s to Supabase: %w", clip.ID, err)
				return
			}
			outputMutex.Lock()
			outputByClip[clip.ID] = uploadedURL
			outputMutex.Unlock()
			log.Info().Msgf("Successfully uploaded clip %s to Supabase. URL: %s", clip.ID, uploadedURL)

			if err := os.Remove(localOutputPath); err != nil {
				log.Warn().Err(err).Msgf("Failed to remove local audio file %s", localOutputPath)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return nil, nil, err
	}

	for i, clip := range req.Timeline.Clips {
		if url, ok := outputByClip[clip.ID]; ok {
			req.Timeline.Clips[i].OutputURL = url
		} else {
			req.Timeline.Clips[i].OutputURL = ""
		}
	}

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

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	currentAssets, err := s.ListPageAudio(req.Jilid, req.Page)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list page audio after publish, returning empty list")
		currentAssets = []AudioAsset{}
	}

	return &edit, currentAssets, nil
}

func (s *TilawatiEditorService) createMergedAudio(segments []FlattenedSegment, outputPath string) error {
	tempDir, err := os.MkdirTemp("", "tilawati-merge-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	clipFiles := []string{}
	for i, seg := range segments {
		sourcePath := s.audioUrlToFile(seg.SourceAudioURL)
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
		concatContent += fmt.Sprintf("file '%s'\n", strings.ReplaceAll(file, "'", "'\\''"))
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

func (s *TilawatiEditorService) audioUrlToFile(audioUrl string) string {
	if strings.HasPrefix(audioUrl, "/audio/tilawati/") {
		return filepath.Join(s.frontendDir, "public", audioUrl)
	}
	return audioUrl
}

func (s *TilawatiEditorService) flattenTimelineSegments(clip TimelineClip) []FlattenedSegment {
	trimStart := clip.TrimStart
	trimEnd := clip.TrimEnd
	var offset float64 = 0
	output := []FlattenedSegment{}

	for _, rawSegment := range clip.Segments {
		segmentStart := rawSegment.Start
		segmentEnd := rawSegment.End
		length := segmentEnd - segmentStart

		visibleStart := math.Max(trimStart, offset)
		visibleEnd := math.Min(trimEnd, offset+length)

		if visibleEnd > visibleStart {
			output = append(output, FlattenedSegment{
				SourceAudioURL: clip.SourceAudioURL,
				Start:          segmentStart + (visibleStart - offset),
				End:            segmentStart + (visibleEnd - offset),
			})
		}
		offset += length
	}
	return output
}

func (s *TilawatiEditorService) ListPageAudio(jilid int, page int) ([]AudioAsset, error) {
	pageAudioPath := filepath.Join(s.audioRoot, fmt.Sprintf("jilid_%d", jilid), fmt.Sprintf("hal%d", page))
	files, err := os.ReadDir(pageAudioPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []AudioAsset{}, nil
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
			IsOpening: strings.Contains(file.Name(), "header_"),
		}
		assets = append(assets, asset)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Filename < assets[j].Filename
	})

	return assets, nil
}

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
