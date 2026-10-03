-- 032: Relasi group_id users<->members<->groups + auto-sync trigger
-- Best practice: FK referential integrity + trigger cegah update anomaly.
-- User tanpa member_id (NULL) tidak tersentuh trigger, tetap pakai users.group_id sendiri.

-- 1) Pastikan FK members.group_id -> groups(group_id)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_members_group') THEN
    ALTER TABLE members ADD CONSTRAINT fk_members_group
      FOREIGN KEY (group_id) REFERENCES groups(group_id) ON UPDATE CASCADE ON DELETE SET NULL;
  END IF;
END $$;

-- 2) Pastikan FK users.group_id -> groups(group_id)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_group') THEN
    ALTER TABLE users ADD CONSTRAINT fk_users_group
      FOREIGN KEY (group_id) REFERENCES groups(group_id) ON UPDATE CASCADE ON DELETE SET NULL;
  END IF;
END $$;

-- 3) Pastikan FK users.member_id -> members(member_id) (mungkin sudah ada sebagai fk_users_member)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_member') THEN
    ALTER TABLE users ADD CONSTRAINT fk_users_member
      FOREIGN KEY (member_id) REFERENCES members(member_id) ON UPDATE CASCADE ON DELETE SET NULL;
  END IF;
END $$;

-- 4) One-time backfill: samakan users.group_id dengan members.group_id untuk user yang terhubung jamaah
UPDATE users u
SET group_id = m.group_id, updated_at = NOW()
FROM members m
WHERE u.member_id = m.member_id
  AND m.group_id IS NOT NULL
  AND (u.group_id IS DISTINCT FROM m.group_id);

-- 5) Trigger auto-sync: setiap members.group_id berubah, ikutkan users yang terhubung
CREATE OR REPLACE FUNCTION sync_user_group_on_member_update()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.group_id IS DISTINCT FROM OLD.group_id THEN
    UPDATE users
    SET group_id = NEW.group_id, updated_at = NOW()
    WHERE member_id = NEW.member_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_user_group ON members;
CREATE TRIGGER trg_sync_user_group
AFTER UPDATE OF group_id ON members
FOR EACH ROW
EXECUTE FUNCTION sync_user_group_on_member_update();
