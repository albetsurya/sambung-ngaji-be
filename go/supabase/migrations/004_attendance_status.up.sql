-- Migration: normalisasi status absensi
--   IJIN             -> IZIN
--   TANPA_KETERANGAN -> ALPA
--   + tambah nilai DISPENSASI
-- Drop & recreate CHECK constraint karena inline constraint auto-named.

ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_status_check;

UPDATE attendance SET status = 'IZIN' WHERE status = 'IJIN';
UPDATE attendance SET status = 'ALPA' WHERE status = 'TANPA_KETERANGAN';

ALTER TABLE attendance
  ADD CONSTRAINT attendance_status_check
  CHECK (status IN ('HADIR','IZIN','SAKIT','ALPA','DISPENSASI'));