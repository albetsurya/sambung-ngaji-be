-- db/migrations/000025_zakat_relational.up.sql
-- ============================================================
-- 000025: Zakat relational (multi-payer, multi-recipient, allocations)
-- snake_case English naming untuk semua kolom/tabel baru.
-- Kolom legacy zakat_records dipertahankan (muzakki_name, soul_count, dst).
-- Idempotent: IF NOT EXISTS di mana pun memungkinkan.
-- ============================================================

-- 1. Extend header zakat_records
ALTER TABLE zakat_records
  ADD COLUMN IF NOT EXISTS title           TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS description     TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS location        TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS zakat_category  TEXT NOT NULL DEFAULT 'FITRAH',
  ADD COLUMN IF NOT EXISTS completed_at    TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS deleted_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS version         INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS updated_by      TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_zakat_category'
  ) THEN
    ALTER TABLE zakat_records ADD CONSTRAINT chk_zakat_category
      CHECK (zakat_category IN ('FITRAH','MAL','TIJAROH','ZURU','LIVESTOCK','OTHER'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_zakat_group_date
  ON zakat_records(group_id, transaction_date DESC) WHERE deleted_at IS NULL;

-- 2. Master payers & recipients
CREATE TABLE IF NOT EXISTS master_payers (
  master_id   TEXT PRIMARY KEY,
  group_id    TEXT NOT NULL REFERENCES groups(group_id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_master_payers_group_name UNIQUE (group_id, name)
);
CREATE INDEX IF NOT EXISTS idx_master_payers_group ON master_payers(group_id);

CREATE TABLE IF NOT EXISTS master_recipients (
  master_id   TEXT PRIMARY KEY,
  group_id    TEXT NOT NULL REFERENCES groups(group_id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_master_recipients_group_name UNIQUE (group_id, name)
);
CREATE INDEX IF NOT EXISTS idx_master_recipients_group ON master_recipients(group_id);

-- 3. Zakat payers (anak)
CREATE TABLE IF NOT EXISTS zakat_payers (
  payer_id              TEXT PRIMARY KEY,
  zakat_id              TEXT NOT NULL REFERENCES zakat_records(zakat_id) ON DELETE CASCADE,
  master_id             TEXT REFERENCES master_payers(master_id) ON DELETE SET NULL,
  name                  TEXT NOT NULL,
  amount                NUMERIC(15,2) NOT NULL DEFAULT 0,
  zakat_category        TEXT NOT NULL,
  family_members_count  INTEGER NOT NULL DEFAULT 0,
  sort_order            INTEGER NOT NULL DEFAULT 0,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_zakat_payers_category
    CHECK (zakat_category IN ('FITRAH','MAL','TIJAROH','ZURU','LIVESTOCK','OTHER'))
);
CREATE INDEX IF NOT EXISTS idx_zakat_payers_zakat  ON zakat_payers(zakat_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_zakat_payers_master ON zakat_payers(master_id);
CREATE INDEX IF NOT EXISTS idx_zakat_payers_name   ON zakat_payers(name);

-- 4. Zakat recipients (anak)
CREATE TABLE IF NOT EXISTS zakat_recipients (
  recipient_id   TEXT PRIMARY KEY,
  zakat_id       TEXT NOT NULL REFERENCES zakat_records(zakat_id) ON DELETE CASCADE,
  master_id      TEXT REFERENCES master_recipients(master_id) ON DELETE SET NULL,
  name           TEXT NOT NULL,
  amount         NUMERIC(15,2) NOT NULL DEFAULT 0,
  zakat_category TEXT NOT NULL,
  sort_order     INTEGER NOT NULL DEFAULT 0,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_zakat_recipients_category
    CHECK (zakat_category IN ('FITRAH','MAL','TIJAROH','ZURU','LIVESTOCK','OTHER'))
);
CREATE INDEX IF NOT EXISTS idx_zakat_recipients_zakat  ON zakat_recipients(zakat_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_zakat_recipients_master ON zakat_recipients(master_id);

-- 5. Zakat allocations (1 baris per zakat per category)
CREATE TABLE IF NOT EXISTS zakat_allocations (
  zakat_id                    TEXT NOT NULL REFERENCES zakat_records(zakat_id) ON DELETE CASCADE,
  category                    TEXT NOT NULL,
  recipient_percent           INTEGER NOT NULL DEFAULT 0,
  recipient_amount            NUMERIC(15,2) NOT NULL DEFAULT 0,
  recipient_group_percent     INTEGER NOT NULL DEFAULT 0,
  recipient_group_amount      NUMERIC(15,2) NOT NULL DEFAULT 0,
  recipient_region_percent    INTEGER NOT NULL DEFAULT 0,
  recipient_region_amount     NUMERIC(15,2) NOT NULL DEFAULT 0,
  sabilillah_percent          INTEGER NOT NULL DEFAULT 0,
  sabilillah_amount           NUMERIC(15,2) NOT NULL DEFAULT 0,
  amil_percent                INTEGER NOT NULL DEFAULT 0,
  amil_amount                 NUMERIC(15,2) NOT NULL DEFAULT 0,
  amil_group_percent          INTEGER NOT NULL DEFAULT 0,
  amil_group_amount           NUMERIC(15,2) NOT NULL DEFAULT 0,
  amil_village_percent        INTEGER NOT NULL DEFAULT 0,
  amil_village_amount         NUMERIC(15,2) NOT NULL DEFAULT 0,
  amil_region_percent         INTEGER NOT NULL DEFAULT 0,
  amil_region_amount          NUMERIC(15,2) NOT NULL DEFAULT 0,
  PRIMARY KEY (zakat_id, category),
  CONSTRAINT chk_zakat_allocations_category CHECK (category IN ('FITRAH','MAAL'))
);

-- 6. Backfill dari details JSONB lama
DO $$
DECLARE
  r RECORD;
  item JSONB;
  idx INT;
  has_details BOOLEAN;
BEGIN
  SELECT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'zakat_records' AND column_name = 'details'
  ) INTO has_details;
  IF NOT has_details THEN RETURN; END IF;

  FOR r IN SELECT z.zakat_id, z.details, z.zakat_type, z.muzakki_name,
                  z.soul_count, z.total_money_rp
           FROM zakat_records z
           WHERE z.details IS NOT NULL AND z.details::text NOT IN ('', '{}')
             AND NOT EXISTS (SELECT 1 FROM zakat_payers p WHERE p.zakat_id = z.zakat_id)
  LOOP
    idx := 0;
    FOR item IN SELECT * FROM jsonb_array_elements(
      COALESCE(r.details->'muzakki_list', '[]'::jsonb))
    LOOP
      INSERT INTO zakat_payers
        (payer_id, zakat_id, name, amount, zakat_category, family_members_count, sort_order)
      VALUES (
        'PYR' || substr(md5(random()::text || r.zakat_id || idx::text), 1, 10),
        r.zakat_id,
        COALESCE(item->>'nama', r.muzakki_name, ''),
        COALESCE((item->>'nominal')::numeric, r.total_money_rp, 0),
        COALESCE(NULLIF(item->>'jenis_zakat',''),
                 CASE WHEN r.zakat_type='MAL' THEN 'MAL' ELSE 'FITRAH' END),
        COALESCE((item->>'jumlah_anggota_keluarga')::int, 0),
        idx
      );
      idx := idx + 1;
    END LOOP;

    IF idx = 0 AND r.muzakki_name IS NOT NULL AND btrim(r.muzakki_name) <> '' THEN
      INSERT INTO zakat_payers
        (payer_id, zakat_id, name, amount, zakat_category, family_members_count, sort_order)
      VALUES (
        'PYR' || substr(md5(random()::text || r.zakat_id), 1, 10),
        r.zakat_id,
        r.muzakki_name,
        COALESCE(r.total_money_rp, 0),
        CASE WHEN r.zakat_type='MAL' THEN 'MAL' ELSE 'FITRAH' END,
        COALESCE(r.soul_count, 0),
        0
      );
    END IF;

    idx := 0;
    FOR item IN SELECT * FROM jsonb_array_elements(
      COALESCE(r.details->'mustahik_list', '[]'::jsonb))
    LOOP
      INSERT INTO zakat_recipients
        (recipient_id, zakat_id, name, amount, zakat_category, sort_order)
      VALUES (
        'RCP' || substr(md5(random()::text || r.zakat_id || idx::text), 1, 10),
        r.zakat_id,
        COALESCE(item->>'nama', ''),
        COALESCE((item->>'nominal')::numeric, 0),
        COALESCE(NULLIF(item->>'jenis_zakat',''), 'FITRAH'),
        idx
      );
      idx := idx + 1;
    END LOOP;
  END LOOP;
END $$;

ALTER TABLE zakat_records DROP COLUMN IF EXISTS details;

-- 7. Trigger update cache total di header
CREATE OR REPLACE FUNCTION trg_zakat_refresh_totals() RETURNS TRIGGER AS $$
DECLARE
  v_zakat_id TEXT := COALESCE(NEW.zakat_id, OLD.zakat_id);
BEGIN
  UPDATE zakat_records SET
    total_money_rp = COALESCE((
      SELECT SUM(amount) FROM zakat_payers WHERE zakat_id = v_zakat_id
    ), 0),
    soul_count = COALESCE((
      SELECT SUM(family_members_count) FROM zakat_payers WHERE zakat_id = v_zakat_id
    ), 0),
    updated_at = now(),
    version = version + 1
  WHERE zakat_id = v_zakat_id;
  RETURN COALESCE(NEW, OLD);
END; $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_zakat_payers_cache ON zakat_payers;
CREATE TRIGGER trg_zakat_payers_cache
  AFTER INSERT OR UPDATE OR DELETE ON zakat_payers
  FOR EACH ROW EXECUTE FUNCTION trg_zakat_refresh_totals();