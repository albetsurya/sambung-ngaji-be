-- 000027 down: kembalikan kolom header + ejaan MAAL alokasi.
-- Catatan: isi zakat_type/zakat_category/muzakki_name tidak bisa dipulihkan
-- (diisi default aman), dan rincian non FITRAH/MAAL harus dihapus manual
-- sebelum rollback agar CHECK lama lolos.
ALTER TABLE zakat_records ADD COLUMN IF NOT EXISTS zakat_type TEXT NOT NULL DEFAULT 'FITRAH';
ALTER TABLE zakat_records ADD COLUMN IF NOT EXISTS zakat_category TEXT NOT NULL DEFAULT 'FITRAH';
ALTER TABLE zakat_records ADD COLUMN IF NOT EXISTS muzakki_name TEXT NOT NULL DEFAULT '';

ALTER TABLE zakat_allocations DROP CONSTRAINT IF EXISTS chk_zakat_allocations_category;
ALTER TABLE zakat_allocations ADD CONSTRAINT chk_zakat_allocations_category
  CHECK (category IN ('FITRAH','MAAL'));
UPDATE zakat_allocations SET category = 'MAAL' WHERE category = 'MAL';
