-- 000011: Tandai reminder WA per jadwal petugas Jumat (anti double-kirim cron).
ALTER TABLE friday_schedules
    ADD COLUMN IF NOT EXISTS reminder_sent_at timestamptz;
