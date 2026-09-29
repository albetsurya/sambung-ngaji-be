-- Down 000022: kembalikan kolom teks dari data anak (derived, bukan data asli
-- breakdown bebas — data asli yang dipindah ke notes tetap di notes).
ALTER TABLE due_payments ADD COLUMN IF NOT EXISTS carryover_months TEXT NOT NULL DEFAULT '';
ALTER TABLE due_payments ADD COLUMN IF NOT EXISTS carryover_breakdown TEXT NOT NULL DEFAULT '';

UPDATE due_payments p
SET carryover_months = COALESCE(c.months, ''),
    carryover_breakdown = COALESCE(c.breakdown, ''),
    updated_at = now()
FROM (
  SELECT payment_id,
         string_agg(month, ', ' ORDER BY month) AS months,
         string_agg(month || ': ' || amount::text, ', ' ORDER BY month) AS breakdown
  FROM due_payment_carryovers GROUP BY payment_id
) c
WHERE c.payment_id = p.payment_id;

DROP TABLE IF EXISTS due_payment_carryovers;
