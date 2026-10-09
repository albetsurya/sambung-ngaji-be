-- 036: Rename cash_transactions.tanggal -> date (terlewat migrasi 033
-- yang menargetkan nama tabel lama kas_transactions). Idempotent.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'cash_transactions' AND column_name = 'tanggal'
  ) THEN
    ALTER TABLE cash_transactions RENAME COLUMN tanggal TO date;
  END IF;
END $$;

DROP INDEX IF EXISTS idx_cash_group_type_date;
CREATE INDEX IF NOT EXISTS idx_cash_group_type_date
  ON cash_transactions(group_id, cash_type, date);
