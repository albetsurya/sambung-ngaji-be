-- 000017 rollback: kembalikan kolom WA otomatis.
-- Data isi kolom (kapan terkirim) TIDAK bisa dikembalikan.
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS send_reminder BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE friday_schedules ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS wa_queue (
  queue_id      TEXT PRIMARY KEY,
  meeting_id    TEXT,
  meeting_date  DATE,
  jam_start     TEXT NOT NULL DEFAULT '',
  send_at       TIMESTAMPTZ,
  status        TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','SENT','FAILED','CANCELLED')),
  template_id   TEXT,
  member_count  INTEGER NOT NULL DEFAULT 0,
  sent_count    INTEGER NOT NULL DEFAULT 0,
  failed_count  INTEGER NOT NULL DEFAULT 0,
  error_log     TEXT NOT NULL DEFAULT '',
  sent_at       TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
