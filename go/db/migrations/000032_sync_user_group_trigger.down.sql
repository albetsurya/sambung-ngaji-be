DROP TRIGGER IF EXISTS trg_sync_user_group ON members;
DROP FUNCTION IF EXISTS sync_user_group_on_member_update();
-- FK dibiarkan (tidak di-drop) agar data tetap aman.
