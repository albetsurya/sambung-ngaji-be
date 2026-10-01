-- db/migrations/000025_zakat_relational.down.sql
DROP TRIGGER IF EXISTS trg_zakat_payers_cache ON zakat_payers;
DROP FUNCTION IF EXISTS trg_zakat_refresh_totals();

ALTER TABLE zakat_records ADD COLUMN IF NOT EXISTS details JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE zakat_records z SET details = jsonb_build_object(
  'muzakki_list', COALESCE((
    SELECT jsonb_agg(jsonb_build_object(
      'id', payer_id, 'nama', name, 'nominal', amount,
      'jenis_zakat', zakat_category,
      'jumlah_anggota_keluarga', family_members_count
    ) ORDER BY sort_order)
    FROM zakat_payers WHERE zakat_id = z.zakat_id
  ), '[]'::jsonb),
  'mustahik_list', COALESCE((
    SELECT jsonb_agg(jsonb_build_object(
      'id', recipient_id, 'nama', name, 'nominal', amount,
      'jenis_zakat', zakat_category
    ) ORDER BY sort_order)
    FROM zakat_recipients WHERE zakat_id = z.zakat_id
  ), '[]'::jsonb)
);

DROP TABLE IF EXISTS zakat_allocations;
DROP TABLE IF EXISTS zakat_recipients;
DROP TABLE IF EXISTS master_recipients;
DROP TABLE IF EXISTS master_payers;

ALTER TABLE zakat_records
  DROP COLUMN IF EXISTS title,
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS location,
  DROP COLUMN IF EXISTS zakat_category,
  DROP COLUMN IF EXISTS completed_at,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS version,
  DROP COLUMN IF EXISTS updated_by;