-- Down: revert status absensi ke set lama.
-- DISPENSASI (nilai baru) tidak valid di set lama -> konversi ke HADIR agar tidak orphan.

ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_status_check;

UPDATE attendance SET status = 'HADIR' WHERE status = 'DISPENSASI';
UPDATE attendance SET status = 'IJIN' WHERE status = 'IZIN';
UPDATE attendance SET status = 'TANPA_KETERANGAN' WHERE status = 'ALPA';

ALTER TABLE attendance
  ADD CONSTRAINT attendance_status_check
  CHECK (status IN ('HADIR','IJIN','SAKIT','TANPA_KETERANGAN'));