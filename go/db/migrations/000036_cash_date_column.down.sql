-- 000036 rollback: kembalikan date -> tanggal.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'cash_transactions' AND column_name = 'date'
  ) THEN
    ALTER TABLE cash_transactions RENAME COLUMN date TO tanggal;
  END IF;
END $$;

DROP INDEX IF EXISTS idx_cash_group_type_date;
CREATE INDEX IF NOT EXISTS idx_cash_group_type_date
  ON cash_transactions(group_id, cash_type, tanggal);
