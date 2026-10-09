CREATE TABLE IF NOT EXISTS ngaji_ceria_progress (
    user_id TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    xp BIGINT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    total_score BIGINT NOT NULL DEFAULT 0 CHECK (total_score >= 0),
    games_played INTEGER NOT NULL DEFAULT 0 CHECK (games_played >= 0),
    coins BIGINT NOT NULL DEFAULT 0 CHECK (coins >= 0),
    current_streak INTEGER NOT NULL DEFAULT 0 CHECK (current_streak >= 0),
    last_activity_date DATE,
    daily_date DATE NOT NULL DEFAULT CURRENT_DATE,
    daily_games_played INTEGER NOT NULL DEFAULT 0 CHECK (daily_games_played >= 0),
    daily_total_score BIGINT NOT NULL DEFAULT 0 CHECK (daily_total_score >= 0),
    daily_claimed_missions TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ngaji_ceria_progress_events (
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    game_id TEXT NOT NULL,
    score BIGINT NOT NULL DEFAULT 0 CHECK (score >= 0),
    base_xp INTEGER NOT NULL DEFAULT 0 CHECK (base_xp >= 0),
    achievement_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_ngaji_ceria_events_leaderboard
    ON ngaji_ceria_progress_events (game_id, created_at DESC, score DESC);

CREATE TABLE IF NOT EXISTS ngaji_ceria_achievements (
    user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id)
);

CREATE OR REPLACE FUNCTION ngaji_ceria_set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END; $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ngaji_ceria_progress_updated ON ngaji_ceria_progress;
CREATE TRIGGER trg_ngaji_ceria_progress_updated
    BEFORE UPDATE ON ngaji_ceria_progress
    FOR EACH ROW EXECUTE FUNCTION ngaji_ceria_set_updated_at();

CREATE TABLE IF NOT EXISTS tilawati_timeline_edit_history (
    history_id BIGSERIAL PRIMARY KEY,
    jilid INT NOT NULL,
    page INT NOT NULL,
    revision TEXT NOT NULL,
    published_timeline_json JSONB NOT NULL,
    published_by TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tilawati_timeline_history_page
    ON tilawati_timeline_edit_history (jilid, page, created_at DESC);
