-- ============================================================
-- 000013_down: Rollback group-based role assignment
-- ============================================================

ALTER TABLE users DROP COLUMN IF EXISTS group_id;
ALTER TABLE members DROP COLUMN IF EXISTS group_id;
