-- ============================================================
-- 000016: friday_schedules per-group (group_id, tanggal) unique
-- Root cause: tanggal UNIQUE global -> dua kelompok tidak bisa
-- punya jadwal Jumat di tanggal yang sama; upsert ON CONFLICT(tanggal)
-- menimpa baris kelompok lain.
-- Strategi: unique per (group_id, tanggal).
-- CATATAN: NULL group_id tidak dianggap sama oleh UNIQUE, jadi baris
-- legacy global (group_id NULL) tetap boleh duplikat tanggal antar NULL.
-- Itu disengaja: data lama global, data baru selalu ber-group.
-- ============================================================

-- 1) Hapus constraint UNIQUE global di tanggal (nama auto-generated,
--    cari lewat pg_constraint agar idempotent di semua env).
DO $$
DECLARE r RECORD;
BEGIN
  FOR r IN
    SELECT conname FROM pg_constraint
    WHERE conrelid = 'friday_schedules'::regclass
      AND contype = 'u'
  LOOP
    -- Pertahankan unique (group_id, tanggal) jika sudah ada
    IF r.conname <> 'uq_friday_group_tanggal' THEN
      EXECUTE 'ALTER TABLE friday_schedules DROP CONSTRAINT ' || quote_ident(r.conname);
    END IF;
  END LOOP;
END $$;

-- 2) Unique per kelompok per tanggal
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'uq_friday_group_tanggal'
  ) THEN
    ALTER TABLE friday_schedules
      ADD CONSTRAINT uq_friday_group_tanggal UNIQUE (group_id, tanggal);
  END IF;
END $$;
