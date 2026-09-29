-- Standardisasi group_id sebagai FK tunggal.
-- Backfill baris lama yang group_id-nya masih NULL tapi kelompok (nama) terisi,
-- dan turunkan users.group_id dari members via member_id.
-- Idempotent dan aman di-rerun.

-- 1. Pastikan kolom ada (untuk DB yang loncat migrasi 013/015)
ALTER TABLE members ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
ALTER TABLE pending_members ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);

-- 2. Isi members.group_id dari groups.group_name (case-insensitive, trim)
UPDATE members m
SET group_id = g.group_id, updated_at = now()
FROM groups g
WHERE (m.group_id IS NULL OR m.group_id = '')
  AND m.kelompok IS NOT NULL AND btrim(m.kelompok) <> ''
  AND lower(btrim(g.group_name)) = lower(btrim(m.kelompok));

-- 3. Sinkronkan members.kelompok dari groups.group_name agar tidak stale
UPDATE members m
SET kelompok = g.group_name, updated_at = now()
FROM groups g
WHERE m.group_id IS NOT NULL AND m.group_id <> ''
  AND g.group_id = m.group_id
  AND (m.kelompok IS DISTINCT FROM g.group_name);

-- 4. Isi users.group_id dari members via member_id
UPDATE users u
SET group_id = m.group_id, updated_at = now()
FROM members m
WHERE (u.group_id IS NULL OR u.group_id = '')
  AND u.member_id IS NOT NULL AND u.member_id <> ''
  AND m.member_id = u.member_id
  AND m.group_id IS NOT NULL AND m.group_id <> '';

-- 5. Index untuk filter per-kelompok (IF NOT EXISTS)
CREATE INDEX IF NOT EXISTS idx_members_group_id ON members(group_id);
CREATE INDEX IF NOT EXISTS idx_users_group_id ON users(group_id);
CREATE INDEX IF NOT EXISTS idx_pending_group_id ON pending_members(group_id);
