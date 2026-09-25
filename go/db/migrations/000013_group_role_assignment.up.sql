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

-- 1d. Backfill users.group_id (assign user ke kelompok via group_code match)
UPDATE users u
SET group_id = g.group_id
FROM groups g
WHERE u.username = g.group_code;

-- 1e. Tambah updated_at trigger jika belum ada
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at'
  ) THEN
    CREATE TRIGGER update_users_updated_at
      BEFORE UPDATE ON users
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END$$;
