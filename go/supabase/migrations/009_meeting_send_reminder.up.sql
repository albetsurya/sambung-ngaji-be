-- Toggle per-meeting: apakah reminder WA dikirim.
-- Default true agar existing meeting tetap terkirim reminder.
-- Migrasi ini dibuat idempotent sehingga bisa dijalankan berulang kali.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'meetings' AND column_name = 'send_reminder') THEN
        ALTER TABLE meetings ADD COLUMN send_reminder BOOLEAN NOT NULL DEFAULT true;
    END IF;
END$$;

-- Index untuk query pending reminder (termasuk filter send_reminder).
CREATE INDEX IF NOT EXISTS idx_meetings_reminder_pending_v2
  ON meetings (tanggal)
  WHERE reminder_sent_at IS NULL AND status != 'LIBUR' AND send_reminder = true;

-- Drop index lama (000008) supaya tidak redundant.
DROP INDEX IF EXISTS idx_meetings_reminder_pending;
