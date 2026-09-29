-- Down: lepas index backfill saja, data group_id dipertahankan (tidak di-NULL-kan agar aman).
DROP INDEX IF EXISTS idx_pending_group_id;
DROP INDEX IF EXISTS idx_users_group_id;
DROP INDEX IF EXISTS idx_members_group_id;
