-- db/migrations/000027_zakat_multitype.up.sql
-- ============================================================
-- 000027: Zakat multi-tipe per record.
--  - Satukan ejaan kategori ke MAL (sebelumnya alokasi memakai MAAL).
--  - Buka CHECK alokasi ke 6 kategori (rincian per tipe).
--  - Hapus kolom tipe/kategori/nama dari header: tipe kini milik
--    tiap muzakki/mustahik/alokasi (zakat_category di tabel anak).
--  - Backfill judul kosong dari muzakki_name sebelum kolom dihapus.
-- Idempotent: IF EXISTS / guard UPDATE dipakai di mana memungkinkan.
-- ============================================================

-- 1. Kanonis MAAL -> MAL (kode setara yang dipakai CHECK anak).
UPDATE zakat_allocations SET category = 'MAL' WHERE category = 'MAAL';
UPDATE zakat_payers SET zakat_category = 'MAL' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'MAAL';
UPDATE zakat_recipients SET zakat_category = 'MAL' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'MAAL';
UPDATE zakat_records SET zakat_category = 'MAL' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'MAAL';
UPDATE zakat_records SET zakat_type = 'MAL' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_type)) IN ('MAAL', 'ZAKAT MAAL', 'ZAKAT MAL');
UPDATE zakat_records SET zakat_category = 'LIVESTOCK' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'TERNAK';
UPDATE zakat_payers SET zakat_category = 'LIVESTOCK' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'TERNAK';
UPDATE zakat_recipients SET zakat_category = 'LIVESTOCK' WHERE UPPER(TRIM(BOTH ' ' FROM zakat_category)) = 'TERNAK';

-- 2. Rincian boleh untuk semua tipe (1 baris per zakat per kategori).
ALTER TABLE zakat_allocations DROP CONSTRAINT IF EXISTS chk_zakat_allocations_category;
ALTER TABLE zakat_allocations ADD CONSTRAINT chk_zakat_allocations_category
  CHECK (category IN ('FITRAH','MAL','TIJAROH','ZURU','LIVESTOCK','OTHER'));

-- 3. Judul jangan sampai kosong setelah muzakki_name dihapus.
UPDATE zakat_records SET title = muzakki_name
WHERE (title IS NULL OR btrim(title) = '')
  AND muzakki_name IS NOT NULL AND btrim(muzakki_name) <> '';

-- 4. Hapus kolom header yang kini redundan.
ALTER TABLE zakat_records DROP COLUMN IF EXISTS zakat_type;
ALTER TABLE zakat_records DROP COLUMN IF EXISTS zakat_category;
ALTER TABLE zakat_records DROP COLUMN IF EXISTS muzakki_name;
