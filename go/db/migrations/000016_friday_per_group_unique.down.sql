-- 000016 rollback: kembalikan UNIQUE global di tanggal.
-- Gagal jika sudah ada duplikat tanggal antar kelompok -> hapus duplikat manual dulu.
ALTER TABLE friday_schedules DROP CONSTRAINT IF EXISTS uq_friday_group_tanggal;
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'friday_schedules'::regclass AND contype = 'u'
  ) THEN
    ALTER TABLE friday_schedules ADD CONSTRAINT friday_schedules_tanggal_key UNIQUE (tanggal);
  END IF;
END $$;
