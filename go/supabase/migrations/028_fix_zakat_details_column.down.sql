-- Rollback: hapus kolom details
ALTER TABLE zakat_records DROP COLUMN IF EXISTS details;
