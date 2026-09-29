-- ============================================================
-- 000019: Finance full-English naming + sheet sync columns
-- ============================================================

ALTER TABLE IF EXISTS kas_transactions RENAME TO cash_transactions;
ALTER TABLE IF EXISTS shodaqoh_members RENAME TO due_members;
ALTER TABLE IF EXISTS shodaqoh_payments RENAME TO due_payments;

ALTER TABLE IF EXISTS cash_transactions RENAME COLUMN kas_id TO cash_id;
ALTER TABLE IF EXISTS cash_transactions RENAME COLUMN kas_type TO cash_type;
ALTER TABLE IF EXISTS due_members RENAME COLUMN member_id TO due_member_id;
UPDATE cash_transactions SET cash_type = 'amil' WHERE cash_type = 'kas_amil';

ALTER TABLE IF EXISTS cash_transactions
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app';
ALTER TABLE IF EXISTS cash_transactions
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;
ALTER TABLE IF EXISTS due_members
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app';
ALTER TABLE IF EXISTS due_members
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;
ALTER TABLE IF EXISTS due_payments
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app';
ALTER TABLE IF EXISTS due_payments
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;
ALTER TABLE IF EXISTS zakat_records
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app';
ALTER TABLE IF EXISTS zakat_records
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;

CREATE TABLE IF NOT EXISTS finance_sync_errors (
  error_id    TEXT PRIMARY KEY,
  group_id    TEXT REFERENCES groups(group_id),
  entity      TEXT NOT NULL DEFAULT '',
  entity_id   TEXT NOT NULL DEFAULT '',
  direction   TEXT NOT NULL DEFAULT '',
  message     TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_finance_sync_errors_group
  ON finance_sync_errors(group_id);
