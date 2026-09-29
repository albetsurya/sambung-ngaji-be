-- ============================================================
-- 000022: Normalisasi susulan (carryover) due_payments
-- ------------------------------------------------------------
-- Masalah lama (1NF dilanggar):
--   carryover_months    TEXT  : daftar bulan multi-nilai ("Jun, Jul")
--   carryover_breakdown TEXT  : teks bebas duplikat months+amount,
--                               bisa kontradiksi dengan carryover_ir
--   carryover_ir        NUMERIC: satu-satunya yang atomic
-- Aturan baru:
--   * Rincian per bulan pindah ke anak tabel due_payment_carryovers
--     (payment_id, month YYYY-MM, amount). UNIQUE(payment_id, month).
--   * carryover_ir dipertahankan sebagai CACHE = SUM(anak),
--     selalu dihitung server (save & sync), bukan input mentah.
--   * carryover_months / carryover_breakdown dihapus setelah backfill.
--     Isi breakdown yang tak terpindahkan disambung ke notes
--     (data user tidak dihilangkan).
-- ============================================================

CREATE TABLE IF NOT EXISTS due_payment_carryovers (
  carryover_id TEXT PRIMARY KEY,
  payment_id   TEXT NOT NULL REFERENCES due_payments(payment_id) ON DELETE CASCADE,
  month        TEXT NOT NULL,
  amount       NUMERIC NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_carryover_payment_month UNIQUE (payment_id, month),
  CONSTRAINT chk_carryover_month_fmt CHECK (month ~ '^[0-9]{4}-(0[1-9]|1[0-2])$'),
  CONSTRAINT chk_carryover_amount_nonneg CHECK (amount >= 0)
);
CREATE INDEX IF NOT EXISTS idx_carryovers_payment ON due_payment_carryovers(payment_id);

-- Backfill: pecah carryover_months (koma/semicolon) jadi baris anak,
-- bagi carryover_ir rata (sisa ke bulan terakhir). Idempotent: hanya
-- untuk payment yang belum punya anak.
DO $$
DECLARE
  r          RECORD;
  tok        TEXT;
  months     TEXT[];
  norm       TEXT;
  m          TEXT;
  per        NUMERIC;
  rem        NUMERIC;
  i          INT;
  yr         TEXT;
  mon        TEXT;
  name2num   JSONB := '{"jan":"01","januari":"01","january":"01","feb":"02","februari":"02","february":"02","mar":"03","maret":"03","march":"03","apr":"04","april":"04","mei":"05","may":"05","jun":"06","juni":"06","june":"06","jul":"07","juli":"07","july":"07","agu":"08","agustus":"08","aug":"08","august":"08","sep":"09","september":"09","sept":"09","okt":"10","oktober":"10","oct":"10","october":"10","nov":"11","november":"11","des":"12","desember":"12","dec":"12","december":"12"}';
  parts      TEXT[];
BEGIN
  FOR r IN SELECT payment_id, payment_date, carryover_ir, carryover_months, carryover_breakdown, notes
           FROM due_payments
  LOOP
    IF EXISTS (SELECT 1 FROM due_payment_carryovers WHERE payment_id = r.payment_id) THEN
      CONTINUE;
    END IF;

    -- Pindahkan breakdown tak-terstruktur ke notes (jangan hilangkan data user).
    IF r.carryover_breakdown IS NOT NULL AND btrim(r.carryover_breakdown) <> '' THEN
      UPDATE due_payments
      SET notes = CASE WHEN btrim(COALESCE(notes,'')) = '' THEN '[Susulan lama] ' || btrim(r.carryover_breakdown)
                       ELSE btrim(notes) || ' | [Susulan lama] ' || btrim(r.carryover_breakdown) END,
          updated_at = now()
      WHERE payment_id = r.payment_id;
    END IF;

    IF r.carryover_ir IS NULL OR r.carryover_ir <= 0 THEN
      CONTINUE;
    END IF;

    months := ARRAY[]::TEXT[];
    IF r.carryover_months IS NOT NULL AND btrim(r.carryover_months) <> '' THEN
      FOREACH tok IN ARRAY string_to_array(replace(r.carryover_months, ';', ','), ',')
      LOOP
        tok := lower(btrim(tok));
        CONTINUE WHEN tok = '';
        norm := NULL;
        -- Format kanonis YYYY-MM
        IF tok ~ '^[0-9]{4}-(0[1-9]|1[0-2])$' THEN
          norm := tok;
        -- MM/YYYY atau M/YYYY
        ELSIF tok ~ '^([0]?[1-9]|1[0-2])/([0-9]{4})$' THEN
          parts := string_to_array(tok, '/');
          mon := lpad(parts[1], 2, '0');
          norm := parts[2] || '-' || mon;
        -- "NamaBulan YYYY"
        ELSIF tok ~ '^[a-z]+ +[0-9]{4}$' THEN
          parts := regexp_split_to_array(tok, '\s+');
          mon := name2num ->> parts[1];
          IF mon IS NOT NULL THEN
            norm := parts[2] || '-' || mon;
          END IF;
        END IF;
        IF norm IS NOT NULL AND NOT (norm = ANY(months)) THEN
          months := months || norm;
        END IF;
      END LOOP;
    END IF;

    IF array_length(months, 1) IS NULL THEN
      -- Tak ada bulan terbaca: satu baris di bulan payment_date.
      INSERT INTO due_payment_carryovers (carryover_id, payment_id, month, amount)
      VALUES ('CRY' || substr(md5(random()::text || r.payment_id), 1, 8),
              r.payment_id, to_char(r.payment_date, 'YYYY-MM'), r.carryover_ir)
      ON CONFLICT DO NOTHING;
    ELSE
      -- Bagi rata; sisa pembulatan ke bulan terakhir.
      per := floor((r.carryover_ir / array_length(months, 1)) * 100) / 100;
      rem := r.carryover_ir - per * (array_length(months, 1) - 1);
      FOR i IN 1 .. array_length(months, 1)
      LOOP
        m := months[i];
        INSERT INTO due_payment_carryovers (carryover_id, payment_id, month, amount)
        VALUES ('CRY' || substr(md5(random()::text || r.payment_id || m), 1, 8),
                r.payment_id, m, CASE WHEN i = array_length(months, 1) THEN rem ELSE per END)
        ON CONFLICT DO NOTHING;
      END LOOP;
    END IF;
  END LOOP;
END $$;

-- Samakan cache carryover_ir dengan SUM anak (untuk baris yang punya anak).
UPDATE due_payments p
SET carryover_ir = COALESCE(c.s, 0), updated_at = now()
FROM (SELECT payment_id, SUM(amount) s FROM due_payment_carryovers GROUP BY payment_id) c
WHERE c.payment_id = p.payment_id
  AND p.carryover_ir IS DISTINCT FROM COALESCE(c.s, 0);

-- Hapus kolom multi-nilai / redundan.
ALTER TABLE due_payments DROP COLUMN IF EXISTS carryover_months;
ALTER TABLE due_payments DROP COLUMN IF EXISTS carryover_breakdown;
