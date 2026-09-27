-- ============================================================
-- 000017: Hapus mekanisme kirim WA otomatis (total).
-- Menghapus: meetings.send_reminder, meetings.reminder_sent_at,
-- friday_schedules.reminder_sent_at, tabel wa_queue,
-- index reminder 008/009.
-- Idempotent: semua DROP pakai IF EXISTS.
-- CATATAN: kode backend yang membaca kolom ini dihapus di commit
-- yang sama — migrasi ini jalan BERDUAAN dengan perubahan kode.
-- Jangan apply migrasi ini tanpa deploy kode baru.
-- ============================================================

DROP INDEX IF EXISTS idx_meetings_reminder_pending;
DROP INDEX IF EXISTS idx_meetings_reminder_pending_v2;

ALTER TABLE meetings DROP COLUMN IF EXISTS send_reminder;
ALTER TABLE meetings DROP COLUMN IF EXISTS reminder_sent_at;

ALTER TABLE friday_schedules DROP COLUMN IF EXISTS reminder_sent_at;

-- Hapus FK dulu agar DROP TABLE tidak gagal di env yang constraint-nya tervalidasi
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_wa_queue_meeting'
  ) THEN
    ALTER TABLE wa_queue DROP CONSTRAINT fk_wa_queue_meeting;
  END IF;
END $$;

DROP TABLE IF EXISTS wa_queue;

DROP INDEX IF EXISTS idx_wa_queue_status;
DROP INDEX IF EXISTS idx_wa_queue_send_at;
DROP INDEX IF EXISTS idx_wa_queue_meeting;
