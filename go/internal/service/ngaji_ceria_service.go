package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ngajiIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

var allowedNgajiAchievements = map[string]bool{
	"first-game": true, "doa-dungeon-clear": true, "nabi-chronicles-complete": true,
	"malaikat-master": true, "ayat-hunter-clear": true, "tetris-lines": true,
	"qari-cepat": true, "pejuang-tajwid": true,
	"jilid-1-clear": true, "jilid-2-clear": true, "jilid-3-clear": true,
	"jilid-4-clear": true, "jilid-5-clear": true, "jilid-6-clear": true, "pejuang-istiqomah": true,
}

type NgajiCeriaService struct{ pool *pgxpool.Pool }

type NgajiCeriaProgress struct {
	XP                   int64    `json:"xp"`
	TotalScore           int64    `json:"totalScore"`
	GamesPlayed          int      `json:"gamesPlayed"`
	Coins                int64    `json:"coins"`
	CurrentStreak        int      `json:"currentStreak"`
	LastActivityDate     *string  `json:"lastActivityDate,omitempty"`
	DailyDate            string   `json:"dailyDate"`
	DailyGamesPlayed     int      `json:"dailyGamesPlayed"`
	DailyTotalScore      int64    `json:"dailyTotalScore"`
	DailyClaimedMissions []string `json:"dailyClaimedMissions"`
	Achievements         []string `json:"achievements"`
}

type NgajiProgressEvent struct {
	EventID        string   `json:"eventId"`
	GameID         string   `json:"gameId"`
	Score          int64    `json:"score"`
	BaseXP         int      `json:"baseXp"`
	AchievementIDs []string `json:"achievementIds"`
}

type NgajiMissionReward struct {
	MissionID string             `json:"missionId"`
	Claimed   bool               `json:"claimed"`
	Coins     int64              `json:"coins"`
	XP        int                `json:"xp"`
	Progress  NgajiCeriaProgress `json:"progress"`
}

type NgajiLeaderboardEntry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Score     int64     `json:"score"`
	GameID    string    `json:"gameId"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ngajiMission struct {
	ID        string
	Target    int64
	Coins     int64
	XP        int
	UsesScore bool
}

var ngajiMissions = map[string]ngajiMission{
	"daily-games-1":    {ID: "daily-games-1", Target: 2, Coins: 30, XP: 50},
	"daily-score-1":    {ID: "daily-score-1", Target: 200, Coins: 50, XP: 80, UsesScore: true},
	"daily-first-game": {ID: "daily-first-game", Target: 1, Coins: 20, XP: 30},
}

func NewNgajiCeriaService(pool *pgxpool.Pool) *NgajiCeriaService {
	return &NgajiCeriaService{pool: pool}
}

func (s *NgajiCeriaService) GetProgress(ctx context.Context, userID string) (*NgajiCeriaProgress, error) {
	if userID == "" {
		return nil, errors.New("user ID wajib diisi")
	}
	if err := s.ensureProgress(ctx, s.pool, userID); err != nil {
		return nil, err
	}
	return s.scanProgress(ctx, s.pool, userID, false)
}

func (s *NgajiCeriaService) AwardProgress(ctx context.Context, userID string, input NgajiProgressEvent) (*NgajiCeriaProgress, error) {
	if err := validateNgajiEvent(userID, input); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("mulai transaksi progress: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := s.ensureProgress(ctx, tx, userID); err != nil {
		return nil, err
	}

	var insertedEvent string
	err = tx.QueryRow(ctx, `
		INSERT INTO ngaji_ceria_progress_events (event_id, user_id, game_id, score, base_xp, achievement_ids)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id, event_id) DO NOTHING
		RETURNING event_id
	`, input.EventID, userID, input.GameID, input.Score, input.BaseXP, input.AchievementIDs).Scan(&insertedEvent)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return s.GetProgress(ctx, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("simpan event progress: %w", err)
	}

	var currentXP, totalScore, coins int64
	var gamesPlayed, currentStreak, dailyGames int
	var lastDate, dailyDate *time.Time
	var dailyScore int64
	var claimed []string
	err = tx.QueryRow(ctx, `
		SELECT xp, total_score, games_played, coins, current_streak, last_activity_date,
		       daily_date, daily_games_played, daily_total_score, daily_claimed_missions
		FROM ngaji_ceria_progress WHERE user_id=$1 FOR UPDATE
	`, userID).Scan(&currentXP, &totalScore, &gamesPlayed, &coins, &currentStreak, &lastDate, &dailyDate, &dailyGames, &dailyScore, &claimed)
	if err != nil {
		return nil, fmt.Errorf("baca progress: %w", err)
	}

	today := dateOnly(time.Now().UTC())
	if dailyDate == nil || dateOnly(*dailyDate) != today {
		dailyGames, dailyScore, claimed = 0, 0, []string{}
	}
	newStreak := nextStreak(currentStreak, lastDate, today)
	xpAward := maxInt64(int64(input.BaseXP), input.Score/10)
	coinAward := maxInt64(5, int64(input.BaseXP)+input.Score) / 10
	if coinAward < 5 {
		coinAward = 5
	}

	_, err = tx.Exec(ctx, `
		UPDATE ngaji_ceria_progress
		SET xp=$2, total_score=$3, games_played=$4, coins=$5, current_streak=$6,
		    last_activity_date=$7, daily_date=$7, daily_games_played=$8, daily_total_score=$9,
		    daily_claimed_missions=$10
		WHERE user_id=$1
	`, userID, currentXP+xpAward, totalScore+input.Score, gamesPlayed+1, coins+coinAward, newStreak, today, dailyGames+1, dailyScore+input.Score, claimed)
	if err != nil {
		return nil, fmt.Errorf("update progress: %w", err)
	}
	achievements := append([]string{"first-game"}, input.AchievementIDs...)
	if newStreak >= 5 {
		achievements = append(achievements, "pejuang-istiqomah")
	}
	for _, achievementID := range achievements {
		if _, err := tx.Exec(ctx, `INSERT INTO ngaji_ceria_achievements (user_id, achievement_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, achievementID); err != nil {
			return nil, fmt.Errorf("simpan achievement: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit progress: %w", err)
	}
	return s.GetProgress(ctx, userID)
}

func (s *NgajiCeriaService) ClaimMission(ctx context.Context, userID, missionID string) (*NgajiMissionReward, error) {
	mission, ok := ngajiMissions[missionID]
	if !ok {
		return nil, errors.New("misi harian tidak dikenal")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err := s.ensureProgress(ctx, tx, userID); err != nil {
		return nil, err
	}

	var xp, coins int64
	var dailyDate *time.Time
	var dailyGames int
	var dailyScore int64
	var claimed []string
	err = tx.QueryRow(ctx, `SELECT xp, coins, daily_date, daily_games_played, daily_total_score, daily_claimed_missions FROM ngaji_ceria_progress WHERE user_id=$1 FOR UPDATE`, userID).
		Scan(&xp, &coins, &dailyDate, &dailyGames, &dailyScore, &claimed)
	if err != nil {
		return nil, err
	}
	today := dateOnly(time.Now().UTC())
	if dailyDate == nil || dateOnly(*dailyDate) != today {
		dailyGames, dailyScore, claimed = 0, 0, []string{}
	}
	if containsNgajiString(claimed, missionID) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		progress, err := s.GetProgress(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &NgajiMissionReward{MissionID: missionID, Claimed: false, Progress: *progress}, nil
	}
	current := dailyGames
	if mission.UsesScore {
		current = int(dailyScore)
	}
	if int64(current) < mission.Target {
		return nil, errors.New("target misi belum tercapai")
	}
	claimed = append(claimed, missionID)
	_, err = tx.Exec(ctx, `UPDATE ngaji_ceria_progress SET xp=xp+$2, coins=coins+$3, daily_date=$4, daily_games_played=$5, daily_total_score=$6, daily_claimed_missions=$7 WHERE user_id=$1`, userID, mission.XP, mission.Coins, today, dailyGames, dailyScore, claimed)
	if err != nil {
		return nil, fmt.Errorf("klaim misi: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	progress, err := s.GetProgress(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &NgajiMissionReward{MissionID: missionID, Claimed: true, Coins: mission.Coins, XP: mission.XP, Progress: *progress}, nil
}

func (s *NgajiCeriaService) GetLeaderboard(ctx context.Context, gameID, period string, limit int) ([]NgajiLeaderboardEntry, error) {
	if !ngajiIDPattern.MatchString(gameID) {
		return nil, errors.New("game ID tidak valid")
	}
	if period != "daily" && period != "weekly" {
		period = "weekly"
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	days := 7
	if period == "daily" {
		days = 1
	}
	rows, err := s.pool.Query(ctx, `
		WITH best AS (
			SELECT DISTINCT ON (e.user_id) e.user_id, e.game_id, e.score, e.created_at
			FROM ngaji_ceria_progress_events e
			WHERE e.game_id=$1 AND e.created_at >= now() - ($2::integer * interval '1 day')
			ORDER BY e.user_id, e.score DESC, e.created_at ASC
		)
		SELECT b.user_id, COALESCE(NULLIF(u.name,''), u.username), b.score, b.game_id, b.created_at
		FROM best b JOIN users u ON u.user_id=b.user_id
		ORDER BY b.score DESC, b.created_at ASC LIMIT $3
	`, gameID, days, limit)
	if err != nil {
		return nil, fmt.Errorf("ambil leaderboard: %w", err)
	}
	defer rows.Close()
	entries := []NgajiLeaderboardEntry{}
	for rows.Next() {
		var entry NgajiLeaderboardEntry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Score, &entry.GameID, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

type progressDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (s *NgajiCeriaService) ensureProgress(ctx context.Context, db progressDB, userID string) error {
	_, err := db.Exec(ctx, `INSERT INTO ngaji_ceria_progress (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, userID)
	if err != nil {
		return fmt.Errorf("inisialisasi progress: %w", err)
	}
	return nil
}

func (s *NgajiCeriaService) scanProgress(ctx context.Context, db progressDB, userID string, lock bool) (*NgajiCeriaProgress, error) {
	lockClause := ""
	if lock {
		lockClause = " FOR UPDATE"
	}
	var progress NgajiCeriaProgress
	var lastDate, dailyDate *time.Time
	var claimed []string
	err := db.QueryRow(ctx, `SELECT xp, total_score, games_played, coins, current_streak, last_activity_date, daily_date, daily_games_played, daily_total_score, daily_claimed_missions FROM ngaji_ceria_progress WHERE user_id=$1`+lockClause, userID).
		Scan(&progress.XP, &progress.TotalScore, &progress.GamesPlayed, &progress.Coins, &progress.CurrentStreak, &lastDate, &dailyDate, &progress.DailyGamesPlayed, &progress.DailyTotalScore, &claimed)
	if err != nil {
		return nil, err
	}
	if lastDate != nil {
		value := dateOnly(*lastDate)
		progress.LastActivityDate = &value
	}
	if dailyDate != nil {
		progress.DailyDate = dateOnly(*dailyDate)
	}
	progress.DailyClaimedMissions = claimed
	if progress.DailyClaimedMissions == nil {
		progress.DailyClaimedMissions = []string{}
	}
	rows, err := s.pool.Query(ctx, `SELECT achievement_id FROM ngaji_ceria_achievements WHERE user_id=$1 ORDER BY earned_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		progress.Achievements = append(progress.Achievements, id)
	}
	return &progress, rows.Err()
}

func validateNgajiEvent(userID string, input NgajiProgressEvent) error {
	if userID == "" {
		return errors.New("user ID wajib diisi")
	}
	if input.EventID == "" || len(input.EventID) > 128 {
		return errors.New("event ID tidak valid")
	}
	if !ngajiIDPattern.MatchString(input.GameID) {
		return errors.New("game ID tidak valid")
	}
	if input.Score < 0 || input.Score > 10_000_000 {
		return errors.New("score di luar batas")
	}
	if input.BaseXP < 0 || input.BaseXP > 10_000 {
		return errors.New("base XP di luar batas")
	}
	seen := make(map[string]struct{}, len(input.AchievementIDs))
	for _, id := range input.AchievementIDs {
		if !allowedNgajiAchievements[id] {
			return fmt.Errorf("achievement tidak dikenal: %s", id)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("achievement duplikat: %s", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func dateOnly(value time.Time) string { return value.UTC().Format("2006-01-02") }

func nextStreak(current int, last *time.Time, today string) int {
	if last == nil {
		return 1
	}
	lastDate := dateOnly(*last)
	if lastDate == today {
		return current
	}
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	if lastDate == yesterday {
		return current + 1
	}
	return 1
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func containsNgajiString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
