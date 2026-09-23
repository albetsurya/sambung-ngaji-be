-- Add gender_target column (idempotent, aman dijalankan ulang)
ALTER TABLE meetings
  ADD COLUMN IF NOT EXISTS gender_target CHAR(1)
  CHECK (gender_target IN ('L', 'P') OR gender_target IS NULL);
