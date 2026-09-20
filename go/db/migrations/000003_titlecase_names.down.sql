-- Down: tidak ada revert. initcap() bersifat lossy (tidak bisa restore casing asli).
-- Kalau butuh revert, restore dari backup.
SELECT 'no-op: titlecase migration cannot be reversed' AS notice;
