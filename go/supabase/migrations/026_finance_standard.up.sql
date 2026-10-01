-- ============================================================
-- 000026: Finance industry-standard hardening (non-breaking)
-- Prinsip: tidak rename kolom, tidak hapus data, idempotent.
-- Standar: PK sudah ada, tambah CHECK, INDEX, updated_at trigger,
-- sync_source/sheet_row untuk tabel master zakat.
-- ============================================================

-- 0. Shared trigger: set updated_at = now() on UPDATE
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END; $$ LANGUAGE plpgsql;

-- 1. cash_transactions: CHECK + INDEX + trigger
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_cash_type') THEN
    ALTER TABLE cash_transactions ADD CONSTRAINT chk_cash_type
      CHECK (cash_type IN ('main','amil'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_cash_amount_nonneg') THEN
    ALTER TABLE cash_transactions ADD CONSTRAINT chk_cash_amount_nonneg
      CHECK (debit >= 0 AND credit >= 0);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_cash_group_type_date
  ON cash_transactions(group_id, cash_type, tanggal);
DROP TRIGGER IF EXISTS trg_cash_set_updated ON cash_transactions;
CREATE TRIGGER trg_cash_set_updated
  BEFORE UPDATE ON cash_transactions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 2. due_members: CHECK + INDEX + trigger
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_due_member_status') THEN
    ALTER TABLE due_members ADD CONSTRAINT chk_due_member_status
      CHECK (status IN ('ACTIVE','INACTIVE','DELETED'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_due_member_target_nonneg') THEN
    ALTER TABLE due_members ADD CONSTRAINT chk_due_member_target_nonneg
      CHECK (monthly_target >= 0);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_due_members_group_status
  ON due_members(group_id, status);
DROP TRIGGER IF EXISTS trg_due_members_set_updated ON due_members;
CREATE TRIGGER trg_due_members_set_updated
  BEFORE UPDATE ON due_members
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 3. due_payments: CHECK + INDEX + trigger
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_due_payment_status') THEN
    ALTER TABLE due_payments ADD CONSTRAINT chk_due_payment_status
      CHECK (status IN ('ACTIVE','INACTIVE','REVERSED','DELETED'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_due_payment_amount_nonneg') THEN
    ALTER TABLE due_payments ADD CONSTRAINT chk_due_payment_amount_nonneg
      CHECK (total_amount >= 0 AND carryover_ir >= 0 AND connecting_fund >= 0
        AND community_dues >= 0 AND outreach_fund >= 0 AND thousand_fund >= 0
        AND funeral_fund >= 0 AND ukhro_mt >= 0);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_due_payments_group_member_date
  ON due_payments(group_id, member_id, payment_date DESC);
CREATE INDEX IF NOT EXISTS idx_due_payments_group_date
  ON due_payments(group_id, payment_date DESC);
DROP TRIGGER IF EXISTS trg_due_payments_set_updated ON due_payments;
CREATE TRIGGER trg_due_payments_set_updated
  BEFORE UPDATE ON due_payments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 4. due_payment_carryovers: trigger (checks sudah ada di 024)
DROP TRIGGER IF EXISTS trg_carryovers_no_updated ON due_payment_carryovers;
-- carryovers immutable-ish: tidak ada updated_at, skip trigger.

-- 5. zakat_records: CHECK status + INDEX + trigger
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_zakat_status') THEN
    ALTER TABLE zakat_records ADD CONSTRAINT chk_zakat_status
      CHECK (status IN ('PENDING','ACTIVE','COMPLETED','CANCELLED','DELETED'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_zakat_money_nonneg') THEN
    ALTER TABLE zakat_records ADD CONSTRAINT chk_zakat_money_nonneg
      CHECK (total_money_rp >= 0 AND total_rice_kg >= 0 AND soul_count >= 0);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_zakat_group_status
  ON zakat_records(group_id, status) WHERE deleted_at IS NULL;
DROP TRIGGER IF EXISTS trg_zakat_set_updated ON zakat_records;
CREATE TRIGGER trg_zakat_set_updated
  BEFORE UPDATE ON zakat_records
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 6. master payers/recipients: sync_source + sheet_row + trigger + index
ALTER TABLE master_payers
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app',
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;
ALTER TABLE master_recipients
  ADD COLUMN IF NOT EXISTS sync_source TEXT NOT NULL DEFAULT 'app',
  ADD COLUMN IF NOT EXISTS sheet_row INTEGER;
DROP TRIGGER IF EXISTS trg_master_payers_set_updated ON master_payers;
CREATE TRIGGER trg_master_payers_set_updated
  BEFORE UPDATE ON master_payers
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
DROP TRIGGER IF EXISTS trg_master_recipients_set_updated ON master_recipients;
CREATE TRIGGER trg_master_recipients_set_updated
  BEFORE UPDATE ON master_recipients
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 7. finance_sync_errors: index waktu untuk pruning/dashboard
CREATE INDEX IF NOT EXISTS idx_finance_sync_errors_group_time
  ON finance_sync_errors(group_id, created_at DESC);
