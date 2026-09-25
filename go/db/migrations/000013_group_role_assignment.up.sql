-- ============================================================
-- 000013: Group-based role assignment
-- Tambah group_id ke users & members agar role per kelompok
-- SUPER_ADMIN tetap global (group_id = NULL)
-- ============================================================

-- 1a. Tambah group_id ke users
ALTER TABLE users ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_users_group_id ON users(group_id);

-- 1b. Tambah group_id ke members (FK yg benar, bukan sekadar string kelompok)
ALTER TABLE members ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_members_group_id ON members(group_id);

-- 1c. Backfill members.group_id dari kelompok name → group_id
UPDATE members m
SET group_id = g.group_id
FROM groups g
WHERE m.kelompok = g.group_name;

-- 1d. Backfill users.group_id (lewat relasi: users -> members -> groups)
-- SUPER_ADMIN sengaja tidak di-set (global)
UPDATE users u
SET group_id = m.group_id
FROM members m
WHERE u.member_id = m.member_id
  AND m.group_id IS NOT NULL
  AND u.role != 'SUPER_ADMIN';

UPDATE users u
SET group_id = g.group_id
FROM groups g
WHERE LOWER(TRIM(u.username)) = LOWER(TRIM(g.group_code))
  AND u.group_id IS NULL
  AND u.role != 'SUPER_ADMIN';
