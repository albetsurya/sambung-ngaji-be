ALTER TABLE meetings
  ADD COLUMN gender_target CHAR(1)
  CHECK (gender_target IN ('L', 'P') OR gender_target IS NULL);
