-- ============================================================
-- 018: Finance per-group (SabilKas): kas, shodaqoh, zakat
-- ============================================================

CREATE TABLE IF NOT EXISTS kas_transactions (
  kas_id        TEXT PRIMARY KEY,
  group_id      TEXT REFERENCES groups(group_id),
  kas_type      TEXT NOT NULL DEFAULT 'main',
  tanggal       DATE NOT NULL,
  account_name  TEXT NOT NULL DEFAULT '',
  description   TEXT NOT NULL DEFAULT '',
  debit         NUMERIC NOT NULL DEFAULT 0,
  credit        NUMERIC NOT NULL DEFAULT 0,
  created_by    TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_kas_group_type_tanggal
  ON kas_transactions(group_id, kas_type, tanggal);

CREATE TABLE IF NOT EXISTS shodaqoh_members (
  member_id       TEXT PRIMARY KEY,
  group_id        TEXT REFERENCES groups(group_id),
  member_name     TEXT NOT NULL DEFAULT '',
  monthly_target  NUMERIC NOT NULL DEFAULT 0,
  status          TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_shodaqoh_members_group
  ON shodaqoh_members(group_id);

CREATE TABLE IF NOT EXISTS shodaqoh_payments (
  payment_id          TEXT PRIMARY KEY,
  group_id            TEXT REFERENCES groups(group_id),
  member_id           TEXT NOT NULL DEFAULT '',
  payment_date        DATE NOT NULL,
  total_amount        NUMERIC NOT NULL DEFAULT 0,
  carryover_ir        NUMERIC NOT NULL DEFAULT 0,
  carryover_months    TEXT NOT NULL DEFAULT '',
  carryover_breakdown TEXT NOT NULL DEFAULT '',
  connecting_fund     NUMERIC NOT NULL DEFAULT 0,
  community_dues      NUMERIC NOT NULL DEFAULT 0,
  outreach_fund       NUMERIC NOT NULL DEFAULT 0,
  thousand_fund       NUMERIC NOT NULL DEFAULT 0,
  funeral_fund        NUMERIC NOT NULL DEFAULT 0,
  ukhro_mt            NUMERIC NOT NULL DEFAULT 0,
  notes               TEXT NOT NULL DEFAULT '',
  status              TEXT NOT NULL DEFAULT 'ACTIVE',
  created_by          TEXT NOT NULL DEFAULT '',
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_shodaqoh_payments_group_member
  ON shodaqoh_payments(group_id, member_id, payment_date);

CREATE TABLE IF NOT EXISTS zakat_records (
  zakat_id        TEXT PRIMARY KEY,
  group_id        TEXT REFERENCES groups(group_id),
  zakat_type      TEXT NOT NULL DEFAULT 'FITRAH',
  muzakki_name    TEXT NOT NULL DEFAULT '',
  soul_count      INTEGER NOT NULL DEFAULT 1,
  total_rice_kg   NUMERIC NOT NULL DEFAULT 0,
  total_money_rp  NUMERIC NOT NULL DEFAULT 0,
  status          TEXT NOT NULL DEFAULT 'PENDING',
  transaction_date DATE,
  details         JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by      TEXT NOT NULL DEFAULT '',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_zakat_records_group
  ON zakat_records(group_id);
