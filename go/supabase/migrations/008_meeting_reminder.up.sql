-- Track kapan reminder WA dikirim untuk meeting.
-- NULL = belum kirim. Non-null = sudah kirim, jangan kirim lagi.
ALTER TABLE meetings
  ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_meetings_reminder_pending
  ON meetings (tanggal)
  WHERE reminder_sent_at IS NULL AND status != 'LIBUR';
