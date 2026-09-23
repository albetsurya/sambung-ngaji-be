-- Migration: member_requests — permintaan user yang sudah login untuk menjadi member
-- (berbeda dari pending_members yang untuk registrasi publik/user baru)
CREATE TABLE IF NOT EXISTS member_requests (
  request_id  TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL UNIQUE,
  nama        TEXT NOT NULL DEFAULT '',
  status      TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  member_id   TEXT,
  reason      TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  reviewed_by TEXT,
  reviewed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_member_requests_status ON member_requests(status);
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_member_requests_user') THEN
        ALTER TABLE member_requests ADD CONSTRAINT fk_member_requests_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE NOT VALID;
    END IF;
END$$;