-- Toggle per-meeting: apakah reminder WA dikirim.
-- Default true agar existing meeting tetap terkirim reminder.
ALTER TABLE meetings
  ADD COLUMN send_reminder BOOLEAN NOT NULL DEFAULT true;

-- Index untuk query pending reminder (termasuk filter send_reminder).
CREATE INDEX idx_meetings_reminder_pending_v2
  ON meetings (tanggal)
  WHERE reminder_sent_at IS NULL AND status != 'LIBUR' AND send_reminder = true;

-- Drop index lama (000008) supaya tidak redundant.
DROP INDEX IF EXISTS idx_meetings_reminder_pending;
