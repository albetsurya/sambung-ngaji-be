-- Rollback: kembalikan DISPENSASI ke CHECK constraint.
-- Data yang sudah dikonversi ke IZIN tidak bisa di-restore otomatis.

ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_status_check;

ALTER TABLE attendance
  ADD CONSTRAINT attendance_status_check
  CHECK (status IN ('HADIR','IZIN','SAKIT','ALPA','DISPENSASI'));
