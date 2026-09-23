-- Rollback: drop column and restore old index
DROP INDEX IF EXISTS idx_meetings_reminder_pending_v2;
ALTER TABLE meetings DROP COLUMN IF EXISTS send_reminder;
CREATE INDEX idx_meetings_reminder_pending
  ON meetings (tanggal)
  WHERE reminder_sent_at IS NULL AND status != 'LIBUR';
