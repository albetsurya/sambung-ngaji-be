-- ============================================================
-- 000014: Add group_id to friday_schedules, pending_members, announcement_templates
-- ============================================================

-- 1. friday_schedules
ALTER TABLE friday_schedules ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_friday_schedules_group_id ON friday_schedules(group_id);

-- 2. pending_members
ALTER TABLE pending_members ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_pending_members_group_id ON pending_members(group_id);

-- 3. announcement_templates
ALTER TABLE announcement_templates ADD COLUMN IF NOT EXISTS group_id TEXT REFERENCES groups(group_id);
CREATE INDEX IF NOT EXISTS idx_announcement_templates_group_id ON announcement_templates(group_id);