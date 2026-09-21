-- Hapus DISPENSASI dari status per-member.
-- Status valid sekarang: HADIR, IZIN, SAKIT, ALPA.
-- Konsep "libur" dipindah sepenuhnya ke level meeting (meeting.status = 'LIBUR').

-- Konversi data existing (no-op kalau tidak ada)
UPDATE attendance SET status = 'IZIN' WHERE status = 'DISPENSASI';

-- Ganti CHECK constraint
ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_status_check;

ALTER TABLE attendance
  ADD CONSTRAINT attendance_status_check
  CHECK (status IN ('HADIR','IZIN','SAKIT','ALPA'));
