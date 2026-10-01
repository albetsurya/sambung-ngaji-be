-- Fix: column "details" does not exist error
-- Tambah kolom details kembali untuk kompatibilitas query lama
ALTER TABLE zakat_records ADD COLUMN IF NOT EXISTS details JSONB NOT NULL DEFAULT '{}'::jsonb;
