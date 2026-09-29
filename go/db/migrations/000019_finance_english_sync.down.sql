DROP TABLE IF EXISTS finance_sync_errors;

ALTER TABLE IF EXISTS zakat_records DROP COLUMN IF EXISTS sheet_row;
ALTER TABLE IF EXISTS zakat_records DROP COLUMN IF EXISTS sync_source;
ALTER TABLE IF EXISTS due_payments DROP COLUMN IF EXISTS sheet_row;
ALTER TABLE IF EXISTS due_payments DROP COLUMN IF EXISTS sync_source;
ALTER TABLE IF EXISTS due_members DROP COLUMN IF EXISTS sheet_row;
ALTER TABLE IF EXISTS due_members DROP COLUMN IF EXISTS sync_source;
ALTER TABLE IF EXISTS cash_transactions DROP COLUMN IF EXISTS sheet_row;
ALTER TABLE IF EXISTS cash_transactions DROP COLUMN IF EXISTS sync_source;

ALTER TABLE IF EXISTS due_members RENAME COLUMN due_member_id TO member_id;
ALTER TABLE IF EXISTS cash_transactions RENAME COLUMN cash_type TO kas_type;
ALTER TABLE IF EXISTS cash_transactions RENAME COLUMN cash_id TO kas_id;

ALTER TABLE IF EXISTS cash_transactions RENAME TO kas_transactions;
ALTER TABLE IF EXISTS due_members RENAME TO shodaqoh_members;
ALTER TABLE IF EXISTS due_payments RENAME TO shodaqoh_payments;
