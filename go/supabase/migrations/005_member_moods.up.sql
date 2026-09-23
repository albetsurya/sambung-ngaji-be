-- Migration: tabel member_moods — pencatatan mood harian member (Phase 2 sync tracker)
CREATE TABLE IF NOT EXISTS member_moods (
  mood_id    TEXT PRIMARY KEY,
  member_id  TEXT NOT NULL,
  mood_key   TEXT NOT NULL CHECK (mood_key IN ('sedih','cemas','syukur','marah','lelah','takut','putus-asa','tenang')),
  tanggal    DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_member_moods_member_date ON member_moods(member_id, tanggal);
CREATE INDEX IF NOT EXISTS idx_member_moods_member ON member_moods(member_id);
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_member_moods_member') THEN
        ALTER TABLE member_moods ADD CONSTRAINT fk_member_moods_member FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE NOT VALID;
    END IF;
END$$;