-- Hapus DISPENSASI dari status per-member.
-- Status valid sekarang: HADIR, IZIN, SAKIT, ALPA.
-- Konsep "libur" dipindah sepenuhnya ke level meeting (meeting.status = 'LIBUR').

-- Konversi data existing (no-op kalau tidak ada)
UPDATE attendance SET status = 'IZIN' WHERE status = 'DISPENSASI';

-- Ganti CHECK constraint
ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_status_check;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'attendance_status_check') THEN
        ALTER TABLE attendance ADD CONSTRAINT attendance_status_check CHECK (status IN ('HADIR','IZIN','SAKIT','ALPA'));
    END IF;
END$$;
