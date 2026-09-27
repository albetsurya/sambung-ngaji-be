-- ============================================================
-- 000015: Repair meetings FK + gender_target check
-- Root cause:
--  a) fk_meetings_group gagal VALIDATE karena ada meetings.group_id
--     yang tidak ada di groups(group_id) (orphan, huruf kecil/typo).
--  b) meetings_gender_target_check gagal karena ada nilai liar
--     selain 'L'/'P'/NULL (mis. '', 'LAKI', dsb).
-- Strategi: bersihkan data dulu, baru VALIDATE constraint.
-- Idempotent: aman dijalankan ulang.
-- ============================================================

-- 1) Normalisasi gender_target liar -> NULL (hanya L/P/NULL yang valid)
UPDATE meetings
SET gender_target = NULL
WHERE gender_target IS NOT NULL
  AND gender_target NOT IN ('L', 'P');

-- Normalisasi string kosong -> NULL (konsisten dengan check OR IS NULL)
UPDATE meetings
SET gender_target = NULL
WHERE gender_target = '';

-- 2) Orphan group_id -> NULL (ON DELETE SET NULL semantics:
--    jadwal tanpa kelompok valid menjadi jadwal global)
UPDATE meetings m
SET group_id = NULL
WHERE m.group_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM groups g WHERE g.group_id = m.group_id);

-- 3) Backfill members.group_id dari kolom teks kelompok -> group_id
--    (butuh kolom dulu; IF NOT EXISTS agar idempotent)
ALTER TABLE members ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_members_group_id ON members(group_id);

UPDATE members m
SET group_id = g.group_id
FROM groups g
WHERE m.group_id IS NULL
  AND m.kelompok IS NOT NULL
  AND lower(trim(m.kelompok)) = lower(trim(g.group_name));

-- 4) Validasi constraint yang tadinya NOT VALID.
--    Gagal di sini = masih ada data kotor -> perbaiki data, bukan drop constraint.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_meetings_group'
  ) THEN
    ALTER TABLE meetings VALIDATE CONSTRAINT fk_meetings_group;
  END IF;
  IF EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'meetings_gender_target_check'
  ) THEN
    ALTER TABLE meetings VALIDATE CONSTRAINT meetings_gender_target_check;
  END IF;
END $$;
