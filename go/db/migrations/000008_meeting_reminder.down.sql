DROP INDEX IF EXISTS idx_meetings_reminder_pending;
ALTER TABLE meetings DROP COLUMN IF EXISTS reminder_sent_at;
