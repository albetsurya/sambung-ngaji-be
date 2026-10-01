-- 000026 down: lepas hardening tanpa hapus data
DROP TRIGGER IF EXISTS trg_cash_set_updated ON cash_transactions;
DROP TRIGGER IF EXISTS trg_due_members_set_updated ON due_members;
DROP TRIGGER IF EXISTS trg_due_payments_set_updated ON due_payments;
DROP TRIGGER IF EXISTS trg_zakat_set_updated ON zakat_records;
DROP TRIGGER IF EXISTS trg_master_payers_set_updated ON master_payers;
DROP TRIGGER IF EXISTS trg_master_recipients_set_updated ON master_recipients;

ALTER TABLE cash_transactions DROP CONSTRAINT IF EXISTS chk_cash_type;
ALTER TABLE cash_transactions DROP CONSTRAINT IF EXISTS chk_cash_amount_nonneg;
ALTER TABLE due_members DROP CONSTRAINT IF EXISTS chk_due_member_status;
ALTER TABLE due_members DROP CONSTRAINT IF EXISTS chk_due_member_target_nonneg;
ALTER TABLE due_payments DROP CONSTRAINT IF EXISTS chk_due_payment_status;
ALTER TABLE due_payments DROP CONSTRAINT IF EXISTS chk_due_payment_amount_nonneg;
ALTER TABLE zakat_records DROP CONSTRAINT IF EXISTS chk_zakat_status;
ALTER TABLE zakat_records DROP CONSTRAINT IF EXISTS chk_zakat_money_nonneg;

DROP INDEX IF EXISTS idx_cash_group_type_date;
DROP INDEX IF EXISTS idx_due_members_group_status;
DROP INDEX IF EXISTS idx_due_payments_group_member_date;
DROP INDEX IF EXISTS idx_due_payments_group_date;
DROP INDEX IF EXISTS idx_zakat_group_status;
DROP INDEX IF EXISTS idx_finance_sync_errors_group_time;

ALTER TABLE master_payers DROP COLUMN IF EXISTS sync_source;
ALTER TABLE master_payers DROP COLUMN IF EXISTS sheet_row;
ALTER TABLE master_recipients DROP COLUMN IF EXISTS sync_source;
ALTER TABLE master_recipients DROP COLUMN IF EXISTS sheet_row;
