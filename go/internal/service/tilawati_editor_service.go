package service

import (
	"context"
	"encoding/json"
	"errors"
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

	"github.com/jackc/pgx/v5"
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

type TilawatiTimelineHistory struct {
	HistoryID             int64           `json:"historyId"`
	Jilid                 int             `json:"jilid"`
	Page                  int             `json:"page"`
	Revision              string          `json:"revision"`
	PublishedTimelineJSON json.RawMessage `json:"publishedTimelineJson"`
	PublishedBy           *string         `json:"publishedBy,omitempty"`
	CreatedAt             time.Time       `json:"createdAt"`
}

type TimelineClipSegment struct {
	SourceAudioURL string  `json:"sourceUrl"`
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
}

type TimelineClip struct {
	ID             string                `json:"id"`
	Label          string                `json:"label"`
	Kind           string                `json:"kind"`
	SourceAudioURL string                `json:"sourceAudioUrl"`
	Segments       []TimelineClipSegment `json:"segments"`
	TrimStart      float64               `json:"trimStart"`
	TrimEnd        float64               `json:"trimEnd"`
	OutputURL      string                `json:"outputUrl,omitempty"`
}

// PublishedTimeline merefleksikan struktur data dari frontend
type PublishedTimeline struct {
	Version int            `json:"version"`
	Clips   []TimelineClip `json:"clips"`
	Texts   []interface{}  `json:"texts"`
}

type AudioAsset struct {
	ID        string  `json:"id"`
	Filename  string  `json:"filename"`
	AudioURL  string  `json:"audioUrl"`
	Duration  float64 `json:"duration"`
	IsOpening bool    `json:"isOpening"`
	Type      string  `json:"type"`
}

type TimelineDraft struct {
	Timeline  PublishedTimeline `json:"timeline"`
	UpdatedAt int64             `json:"updatedAt"`
	Revision  string            `json:"revision"`
}

func (s *TilawatiEditorService) GetDraft(ctx context.Context, userID string, jilid, page int) (*TimelineDraft, error) {
	var draft TimelineDraft
	var raw json.RawMessage
	err := s.db.QueryRow(ctx, `SELECT timeline_json, updated_at_ms, base_revision FROM tilawati_timeline_drafts WHERE user_id=$1 AND jilid=$2 AND page=$3`, userID, jilid, page).Scan(&raw, &draft.UpdatedAt, &draft.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &draft.Timeline); err != nil {
		return nil, err
	}
	return &draft, nil
}

func validateTimelineDraft(jilid, page int, draft TimelineDraft) error {
	if jilid < 1 || jilid > 6 || page < 1 || page > 44 || draft.Timeline.Version != 2 || draft.UpdatedAt <= 0 || draft.UpdatedAt > time.Now().Add(5*time.Minute).UnixMilli() {
		return fmt.Errorf("draft tidak valid")
	}
	if len(draft.Timeline.Clips) > 500 || len(draft.Timeline.Texts) > 1000 {
		return fmt.Errorf("draft terlalu besar")
	}
	return nil
}

func (s *TilawatiEditorService) SaveDraft(ctx context.Context, userID string, jilid, page int, draft TimelineDraft) error {
	if err := validateTimelineDraft(jilid, page, draft); err != nil {
		return err
	}
	raw, err := json.Marshal(draft.Timeline)
	if err != nil {
		return err
	}
	if len(raw) > 2*1024*1024 {
		return fmt.Errorf("draft terlalu besar")
	}
	// Advisory lock menyamakan urutan draft-save dan publish untuk halaman ini.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1,$2)`, jilid, page); err != nil {
		return err
	}
	var revision string
	err = tx.QueryRow(ctx, `SELECT revision FROM tilawati_timeline_edits WHERE jilid=$1 AND page=$2`, jilid, page).Scan(&revision)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if revision != draft.Revision {
		return fmt.Errorf("conflict: halaman sudah dipublikasikan ulang; muat ulang sebelum menyimpan draft")
	}
	result, err := tx.Exec(ctx, `INSERT INTO tilawati_timeline_drafts (user_id,jilid,page,timeline_json,base_revision,updated_at_ms)
		VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (user_id,jilid,page) DO UPDATE SET timeline_json=EXCLUDED.timeline_json, base_revision=EXCLUDED.base_revision, updated_at_ms=EXCLUDED.updated_at_ms
		WHERE tilawati_timeline_drafts.updated_at_ms <= EXCLUDED.updated_at_ms`, userID, jilid, page, raw, draft.Revision, draft.UpdatedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("conflict: draft cloud lebih baru; buka ulang editor untuk sinkronisasi")
	}
	return tx.Commit(ctx)
}

type PublishTimelineRequest struct {
	Jilid            int               `json:"jilid"`
	Page             int               `json:"page"`
	Timeline         PublishedTimeline `json:"timeline"`
	ExpectedRevision string            `json:"expectedRevision"`
	PublishedBy      string            `json:"-"`
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
	if err := s.validatePublishRequest(req); err != nil {
		return nil, nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1,$2)`, req.Jilid, req.Page); err != nil {
		return nil, nil, err
	}
	if s.storageSvc == nil {
		return nil, nil, fmt.Errorf("storage audio belum dikonfigurasi")
	}

	// Optimistic locking check
	var currentRevision string
	var currentTimeline json.RawMessage
	err = tx.QueryRow(ctx, `SELECT revision, published_timeline_json FROM tilawati_timeline_edits WHERE jilid = $1 AND page = $2 FOR UPDATE`, req.Jilid, req.Page).Scan(&currentRevision, &currentTimeline)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, nil, fmt.Errorf("failed to check current revision: %w", err)
	}
	if currentRevision != req.ExpectedRevision {
		return nil, nil, fmt.Errorf("conflict: timeline has been modified by another user (expected %s, got %s)", req.ExpectedRevision, currentRevision)
	}
	if currentRevision != "" && len(currentTimeline) > 0 {
		if _, err := tx.Exec(ctx, `INSERT INTO tilawati_timeline_edit_history (jilid, page, revision, published_timeline_json, published_by) VALUES ($1,$2,$3,$4,$5)`, req.Jilid, req.Page, currentRevision, currentTimeline, nullableString(req.PublishedBy)); err != nil {
			return nil, nil, fmt.Errorf("failed to save timeline history: %w", err)
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
	committed := false
	defer func() {
		if committed || s.storageSvc == nil {
			return
		}
		for _, uploadedURL := range outputByClip {
			_ = s.storageSvc.DeletePhoto(context.Background(), uploadedURL)
		}
	}()

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
	committed = true

	currentAssets, err := s.ListPageAudio(req.Jilid, req.Page)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list page audio after publish, returning empty list")
		currentAssets = []AudioAsset{}
	}

	return &edit, currentAssets, nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *TilawatiEditorService) ListHistory(ctx context.Context, jilid, page, limit int) ([]TilawatiTimelineHistory, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `SELECT history_id, jilid, page, revision, published_timeline_json, published_by, created_at FROM tilawati_timeline_edit_history WHERE jilid=$1 AND page=$2 ORDER BY created_at DESC LIMIT $3`, jilid, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list timeline history: %w", err)
	}
	defer rows.Close()
	history := []TilawatiTimelineHistory{}
	for rows.Next() {
		var item TilawatiTimelineHistory
		if err := rows.Scan(&item.HistoryID, &item.Jilid, &item.Page, &item.Revision, &item.PublishedTimelineJSON, &item.PublishedBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func (s *TilawatiEditorService) createMergedAudio(segments []FlattenedSegment, outputPath string) error {
	tempDir, err := os.MkdirTemp("", "tilawati-merge-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	clipFiles := []string{}
	for i, seg := range segments {
		sourcePath, err := s.audioURLToFile(seg.SourceAudioURL)
		if err != nil {
			return fmt.Errorf("invalid source audio for segment %d: %w", i, err)
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

func (s *TilawatiEditorService) audioURLToFile(audioURL string) (string, error) {
	const prefix = "/audio/tilawati/"
	if !strings.HasPrefix(audioURL, prefix) {
		return "", fmt.Errorf("source audio must be a local Tilawati asset")
	}

	relative := filepath.Clean(strings.TrimPrefix(audioURL, prefix))
	if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("source audio path escapes the Tilawati asset directory")
	}
	return filepath.Join(s.audioRoot, relative), nil
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
				SourceAudioURL: firstNonEmpty(rawSegment.SourceAudioURL, clip.SourceAudioURL),
				Start:          segmentStart + (visibleStart - offset),
				End:            segmentStart + (visibleEnd - offset),
			})
		}
		offset += length
	}
	return output
}

func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (s *TilawatiEditorService) validatePublishRequest(req PublishTimelineRequest) error {
	if req.Jilid < 1 || req.Jilid > 6 || req.Page < 1 || req.Page > 44 {
		return fmt.Errorf("jilid dan halaman harus positif")
	}
	if len(req.Timeline.Clips) == 0 {
		return fmt.Errorf("timeline harus memiliki setidaknya satu klip")
	}

	seenIDs := make(map[string]struct{}, len(req.Timeline.Clips))
	for _, clip := range req.Timeline.Clips {
		if clip.ID == "" {
			return fmt.Errorf("setiap klip harus memiliki id")
		}
		if _, exists := seenIDs[clip.ID]; exists {
			return fmt.Errorf("id klip duplikat: %s", clip.ID)
		}
		seenIDs[clip.ID] = struct{}{}
		if len(clip.Segments) == 0 {
			return fmt.Errorf("klip %s tidak memiliki segmen audio", clip.ID)
		}
		if !isFiniteNonNegative(clip.TrimStart) || !isFiniteNonNegative(clip.TrimEnd) || clip.TrimEnd <= clip.TrimStart {
			return fmt.Errorf("rentang trim klip %s tidak valid", clip.ID)
		}

		var total float64
		for _, segment := range clip.Segments {
			if !isFiniteNonNegative(segment.Start) || !isFiniteNonNegative(segment.End) || segment.End <= segment.Start {
				return fmt.Errorf("rentang segmen klip %s tidak valid", clip.ID)
			}
			if _, err := s.audioURLToFile(firstNonEmpty(segment.SourceAudioURL, clip.SourceAudioURL)); err != nil {
				return fmt.Errorf("sumber audio klip %s tidak valid: %w", clip.ID, err)
			}
			total += segment.End - segment.Start
		}
		if clip.TrimEnd > total+0.001 {
			return fmt.Errorf("rentang trim klip %s melebihi durasi segmen", clip.ID)
		}
	}
	return nil
}

func isFiniteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func (s *TilawatiEditorService) AudioToolsAvailable() (bool, bool) {
	_, ffmpegErr := exec.LookPath("ffmpeg")
	_, ffprobeErr := exec.LookPath("ffprobe")
	return ffmpegErr == nil, ffprobeErr == nil
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
			Type:      audioAssetType(file.Name()),
		}
		assets = append(assets, asset)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Filename < assets[j].Filename
	})

	return assets, nil
}

func audioAssetType(filename string) string {
	for _, kind := range []string{"header", "demo", "baris", "footer"} {
		if strings.Contains(filename, "_"+kind) {
			return kind
		}
	}
	return "isi"
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
