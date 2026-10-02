-- Schema init untuk aplikasi Pengajian
-- 15 tabel dari Config.js + ai_quota

-- ============================================================
-- 1. members
-- ============================================================
CREATE TABLE members (
  member_id                TEXT PRIMARY KEY,
  nama_lengkap             TEXT NOT NULL DEFAULT '',
  nama_panggilan           TEXT NOT NULL DEFAULT '',
  jenis_kelamin            CHAR(1) CHECK (jenis_kelamin IN ('L','P') OR jenis_kelamin IS NULL),
  tempat_lahir             TEXT NOT NULL DEFAULT '',
  tanggal_lahir            DATE,
  foto_url                 TEXT NOT NULL DEFAULT '',
  no_wa                    TEXT NOT NULL DEFAULT '',
  alamat_rumah             TEXT NOT NULL DEFAULT '',
  desa                     TEXT NOT NULL DEFAULT '',
  daerah                   TEXT NOT NULL DEFAULT '',
  kelompok                 TEXT NOT NULL DEFAULT '',
  is_muballigh             BOOLEAN NOT NULL DEFAULT false,
  is_kerja                 BOOLEAN NOT NULL DEFAULT false,
  is_nikah                 BOOLEAN NOT NULL DEFAULT false,
  tinggi_badan             TEXT NOT NULL DEFAULT '',
  berat_badan              TEXT NOT NULL DEFAULT '',
  hobi                     TEXT NOT NULL DEFAULT '',
  pekerjaan                TEXT NOT NULL DEFAULT '',
  status_pembinaan         TEXT NOT NULL DEFAULT 'AKTIF' CHECK (status_pembinaan IN ('AKTIF','PERLU_PERHATIAN','KURANG_AKTIF','TIDAK_AKTIF')),
  status_aktif             BOOLEAN NOT NULL DEFAULT true,
  tanggal_masuk            DATE,
  tanggal_keluar           DATE,
  jenjang_pendidikan       TEXT NOT NULL DEFAULT '',
  sekolah                  TEXT NOT NULL DEFAULT '',
  jurusan                  TEXT NOT NULL DEFAULT '',
  tahun_mulai_pendidikan   TEXT NOT NULL DEFAULT '',
  tahun_selesai_pendidikan TEXT NOT NULL DEFAULT '',
  created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_members_status_aktif ON members(status_aktif) WHERE status_aktif = true;
CREATE INDEX idx_members_kelompok ON members(kelompok);
CREATE INDEX idx_members_nama_lengkap ON members(nama_lengkap);
CREATE INDEX idx_members_no_wa ON members(no_wa);
CREATE INDEX idx_members_jenis_kelamin ON members(jenis_kelamin);

-- ============================================================
-- 2. users
-- ============================================================
CREATE TABLE users (
  user_id        TEXT PRIMARY KEY,
  username       TEXT NOT NULL UNIQUE,
  password_hash  TEXT NOT NULL,
  nama           TEXT NOT NULL DEFAULT '',
  role           TEXT NOT NULL CHECK (role IN ('SUPER_ADMIN','ADMIN','TIM_PNKB','TIM_ABSENSI','PENGAWAS','TIM_KU','MEMBER')),
  member_id      TEXT,
  status_aktif   BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at  TIMESTAMPTZ
);
CREATE INDEX idx_users_member_id ON users(member_id);
ALTER TABLE users ADD CONSTRAINT fk_users_member FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE SET NULL NOT VALID;

-- ============================================================
-- 3. groups
-- ============================================================
CREATE TABLE groups (
  group_id        TEXT PRIMARY KEY,
  group_code      TEXT NOT NULL DEFAULT '',
  group_name      TEXT NOT NULL DEFAULT '',
  pembina         TEXT NOT NULL DEFAULT '',
  penandatangan   TEXT NOT NULL DEFAULT '',
  jadwal          TEXT NOT NULL DEFAULT '',
  status_aktif    BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- 4. meetings
-- ============================================================
CREATE TABLE meetings (
  meeting_id        TEXT PRIMARY KEY,
  tanggal           DATE NOT NULL,
  hari              TEXT NOT NULL DEFAULT '',
  jam               TEXT NOT NULL DEFAULT '',
  jam_start         TEXT NOT NULL DEFAULT '',
  group_id          TEXT,
  acara             TEXT NOT NULL DEFAULT '',
  materi            TEXT NOT NULL DEFAULT '',
  status            TEXT NOT NULL DEFAULT 'SCHEDULED',
  catatan           TEXT NOT NULL DEFAULT '',
  kategori_target   JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_by        TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_meetings_tanggal ON meetings(tanggal);
CREATE INDEX idx_meetings_group_tanggal ON meetings(group_id, tanggal);
ALTER TABLE meetings ADD CONSTRAINT fk_meetings_group FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE SET NULL NOT VALID;

-- ============================================================
-- 5. announcement_templates
-- ============================================================
CREATE TABLE announcement_templates (
  template_id    TEXT PRIMARY KEY,
  nama_template  TEXT NOT NULL DEFAULT '',
  kode           TEXT NOT NULL UNIQUE,
  isi_template   TEXT NOT NULL DEFAULT '',
  status_aktif   BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- 6. attendance
-- ============================================================
CREATE TABLE attendance (
  attendance_id  TEXT PRIMARY KEY,
  meeting_id     TEXT NOT NULL,
  member_id      TEXT NOT NULL,
  status         TEXT NOT NULL CHECK (status IN ('HADIR','IJIN','SAKIT','TANPA_KETERANGAN')),
  catatan        TEXT NOT NULL DEFAULT '',
  created_by     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_attendance_meeting ON attendance(meeting_id);
CREATE INDEX idx_attendance_member ON attendance(member_id);
CREATE UNIQUE INDEX idx_attendance_meeting_member ON attendance(meeting_id, member_id);
ALTER TABLE attendance ADD CONSTRAINT fk_attendance_meeting FOREIGN KEY (meeting_id) REFERENCES meetings(meeting_id) ON DELETE CASCADE NOT VALID;
ALTER TABLE attendance ADD CONSTRAINT fk_attendance_member FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE NOT VALID;

-- ============================================================
-- 7. monitoring
-- ============================================================
CREATE TABLE monitoring (
  monitoring_id   TEXT PRIMARY KEY,
  member_id       TEXT NOT NULL,
  tanggal         DATE NOT NULL,
  jenis           TEXT NOT NULL DEFAULT 'UMUM',
  status          TEXT NOT NULL CHECK (status IN ('AKTIF','PERLU_PERHATIAN','KURANG_AKTIF','TIDAK_AKTIF')),
  catatan         TEXT NOT NULL DEFAULT '',
  tindak_lanjut   TEXT NOT NULL DEFAULT '',
  created_by      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_monitoring_member ON monitoring(member_id);
CREATE INDEX idx_monitoring_tanggal ON monitoring(tanggal DESC);
ALTER TABLE monitoring ADD CONSTRAINT fk_monitoring_member FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE NOT VALID;

-- ============================================================
-- 8. announcements
-- ============================================================
CREATE TABLE announcements (
  announcement_id  TEXT PRIMARY KEY,
  template_id      TEXT,
  meeting_id       TEXT,
  group_id         TEXT,
  tanggal          DATE NOT NULL,
  hari             TEXT NOT NULL DEFAULT '',
  jam              TEXT NOT NULL DEFAULT '',
  acara            TEXT NOT NULL DEFAULT '',
  materi           TEXT NOT NULL DEFAULT '',
  catatan          TEXT NOT NULL DEFAULT '',
  generated_text   TEXT NOT NULL DEFAULT '',
  status           TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','READY','SHARED','CANCELLED')),
  created_by       TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_announcements_group ON announcements(group_id);
CREATE INDEX idx_announcements_tanggal ON announcements(tanggal DESC);
CREATE INDEX idx_announcements_status ON announcements(status);
ALTER TABLE announcements ADD CONSTRAINT fk_announcements_template FOREIGN KEY (template_id) REFERENCES announcement_templates(template_id) ON DELETE SET NULL NOT VALID;
ALTER TABLE announcements ADD CONSTRAINT fk_announcements_meeting FOREIGN KEY (meeting_id) REFERENCES meetings(meeting_id) ON DELETE SET NULL NOT VALID;
ALTER TABLE announcements ADD CONSTRAINT fk_announcements_group FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE SET NULL NOT VALID;

-- ============================================================
-- 9. pending_members
-- ============================================================
CREATE TABLE pending_members (
  submission_id            TEXT PRIMARY KEY,
  nama_lengkap             TEXT NOT NULL DEFAULT '',
  nama_panggilan           TEXT NOT NULL DEFAULT '',
  jenis_kelamin            CHAR(1) CHECK (jenis_kelamin IN ('L','P') OR jenis_kelamin IS NULL),
  tempat_lahir             TEXT NOT NULL DEFAULT '',
  tanggal_lahir            DATE,
  no_wa                    TEXT NOT NULL DEFAULT '',
  alamat_rumah             TEXT NOT NULL DEFAULT '',
  desa                     TEXT NOT NULL DEFAULT '',
  daerah                   TEXT NOT NULL DEFAULT '',
  pekerjaan                TEXT NOT NULL DEFAULT '',
  hobi                     TEXT NOT NULL DEFAULT '',
  is_nikah                 BOOLEAN NOT NULL DEFAULT false,
  jenjang_pendidikan       TEXT NOT NULL DEFAULT '',
  sekolah                  TEXT NOT NULL DEFAULT '',
  jurusan                  TEXT NOT NULL DEFAULT '',
  tahun_mulai_pendidikan   TEXT NOT NULL DEFAULT '',
  tahun_selesai_pendidikan TEXT NOT NULL DEFAULT '',
  foto_url                 TEXT NOT NULL DEFAULT '',
  username                 TEXT NOT NULL DEFAULT '',
  password_hash            TEXT NOT NULL DEFAULT '',
  status                   TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  submitted_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  submitted_ip             TEXT NOT NULL DEFAULT '',
  reviewed_by              TEXT,
  reviewed_at              TIMESTAMPTZ,
  rejection_reason         TEXT NOT NULL DEFAULT '',
  created_member_id        TEXT
);
CREATE INDEX idx_pending_status ON pending_members(status);
CREATE INDEX idx_pending_no_wa ON pending_members(no_wa);
CREATE INDEX idx_pending_username ON pending_members(username);
CREATE INDEX idx_pending_submitted_at ON pending_members(submitted_at DESC);

-- ============================================================
-- 10. sessions
-- ============================================================
CREATE TABLE sessions (
  token       TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE NOT VALID;

-- ============================================================
-- 11. audit_logs
-- ============================================================
CREATE TABLE audit_logs (
  log_id       TEXT PRIMARY KEY,
  user_id      TEXT,
  user_nama    TEXT NOT NULL DEFAULT '',
  action       TEXT NOT NULL DEFAULT '',
  target_type  TEXT NOT NULL DEFAULT '',
  target_id    TEXT NOT NULL DEFAULT '',
  timestamp    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id);

-- ============================================================
-- 12. ai_usage
-- ============================================================
CREATE TABLE ai_usage (
  usage_id       TEXT PRIMARY KEY,
  user_id        TEXT,
  user_nama      TEXT NOT NULL DEFAULT '',
  role           TEXT NOT NULL DEFAULT '',
  provider       TEXT NOT NULL DEFAULT '',
  input_tokens   INTEGER NOT NULL DEFAULT 0,
  output_tokens  INTEGER NOT NULL DEFAULT 0,
  total_tokens   INTEGER NOT NULL DEFAULT 0,
  timestamp      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_usage_user_timestamp ON ai_usage(user_id, timestamp DESC);
CREATE INDEX idx_ai_usage_timestamp ON ai_usage(timestamp DESC);

-- ============================================================
-- 13. ai_quota
-- ============================================================
CREATE TABLE ai_quota (
  user_id  TEXT NOT NULL,
  date     DATE NOT NULL,
  count    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, date)
);
CREATE INDEX idx_ai_quota_date ON ai_quota(date);

-- ============================================================
-- 14. wa_queue
-- ============================================================
CREATE TABLE wa_queue (
  queue_id      TEXT PRIMARY KEY,
  meeting_id    TEXT,
  meeting_date  DATE,
  jam_start     TEXT NOT NULL DEFAULT '',
  send_at       TIMESTAMPTZ,
  status        TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','SENT','FAILED','CANCELLED')),
  template_id   TEXT,
  member_count  INTEGER NOT NULL DEFAULT 0,
  sent_count    INTEGER NOT NULL DEFAULT 0,
  failed_count  INTEGER NOT NULL DEFAULT 0,
  error_log     TEXT NOT NULL DEFAULT '',
  sent_at       TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wa_queue_status ON wa_queue(status);
CREATE INDEX idx_wa_queue_send_at ON wa_queue(send_at);
CREATE INDEX idx_wa_queue_meeting ON wa_queue(meeting_id);
ALTER TABLE wa_queue ADD CONSTRAINT fk_wa_queue_meeting FOREIGN KEY (meeting_id) REFERENCES meetings(meeting_id) ON DELETE CASCADE NOT VALID;

-- ============================================================
-- 15. settings
-- ============================================================
CREATE TABLE settings (
  key         TEXT PRIMARY KEY,
  value       TEXT NOT NULL DEFAULT '',
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
