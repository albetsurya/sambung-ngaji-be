var SPREADSHEET_ID = "";
var DRIVE_FOLDER_ID = "";
var DRIVE_ARCHIVE_FOLDER_ID = "";
var SESSION_TTL_HOURS = 12;

var PROP_KEY_SPREADSHEET_ID = "SPREADSHEET_ID";
var PROP_KEY_DRIVE_FOLDER_ID = "DRIVE_FOLDER_ID";
var PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID = "DRIVE_ARCHIVE_FOLDER_ID";

var SHEETS = {
  users: {
    name: "users",
    headers: [
      "user_id",
      "username",
      "password_hash",
      "nama",
      "role",
      "member_id",
      "status_aktif",
      "created_at",
      "updated_at",
      "last_login_at",
    ],
  },
  members: {
    name: "members",
    headers: [
      "member_id",
      "nama_lengkap",
      "nama_panggilan",
      "jenis_kelamin",
      "tempat_lahir",
      "tanggal_lahir",
      "foto_url",
      "no_wa",
      "alamat_rumah",
      "desa",
      "daerah",
      "kelompok",
      "is_muballigh",
      "is_kerja",
      "is_nikah",
      "tinggi_badan",
      "berat_badan",
      "hobi",
      "pekerjaan",
      "status_pembinaan",
      "status_aktif",
      "tanggal_masuk",
      "tanggal_keluar",
      "jenjang_pendidikan",
      "sekolah",
      "jurusan",
      "tahun_mulai_pendidikan",
      "tahun_selesai_pendidikan",
      "created_at",
      "updated_at",
    ],
  },
  groups: {
    name: "groups",
    headers: [
      "group_id",
      "group_code",
      "group_name",
      "pembina",
      "penandatangan",
      "jadwal",
      "status_aktif",
      "created_at",
      "updated_at",
    ],
  },
  meetings: {
    name: "meetings",
    headers: [
      "meeting_id",
      "tanggal",
      "hari",
      "jam",
      "group_id",
      "acara",
      "materi",
      "status",
      "catatan",
      "kategori_target",
      "created_by",
      "created_at",
      "updated_at",
    ],
  },
  attendance: {
    name: "attendance",
    headers: [
      "attendance_id",
      "meeting_id",
      "member_id",
      "status",
      "catatan",
      "created_by",
      "created_at",
      "updated_at",
    ],
  },
  monitoring: {
    name: "monitoring",
    headers: [
      "monitoring_id",
      "member_id",
      "tanggal",
      "jenis",
      "status",
      "catatan",
      "tindak_lanjut",
      "created_by",
      "created_at",
      "updated_at",
    ],
  },
  announcements: {
    name: "announcements",
    headers: [
      "announcement_id",
      "template_id",
      "meeting_id",
      "group_id",
      "tanggal",
      "hari",
      "jam",
      "acara",
      "materi",
      "catatan",
      "generated_text",
      "status",
      "created_by",
      "created_at",
      "updated_at",
    ],
  },
  announcement_templates: {
    name: "announcement_templates",
    headers: [
      "template_id",
      "nama_template",
      "kode",
      "isi_template",
      "status_aktif",
      "created_at",
      "updated_at",
    ],
  },
  audit_logs: {
    name: "audit_logs",
    headers: [
      "log_id",
      "user_id",
      "action",
      "target_type",
      "target_id",
      "timestamp",
    ],
  },
  settings: {
    name: "settings",
    headers: ["key", "value", "updated_at"],
  },
  sessions: {
    name: "sessions",
    headers: ["token", "user_id", "created_at", "expires_at"],
  },
  ai_usage: {
    name: "ai_usage",
    headers: [
      "usage_id",
      "user_id",
      "user_nama",
      "role",
      "provider",
      "input_tokens",
      "output_tokens",
      "total_tokens",
      "timestamp",
    ],
  },
  pending_members: {
    name: "pending_members",
    headers: [
      "submission_id",
      "nama_lengkap",
      "nama_panggilan",
      "jenis_kelamin",
      "tempat_lahir",
      "tanggal_lahir",
      "no_wa",
      "alamat_rumah",
      "desa",
      "daerah",
      "pekerjaan",
      "hobi",
      "is_nikah",
      "jenjang_pendidikan",
      "sekolah",
      "jurusan",
      "tahun_mulai_pendidikan",
      "tahun_selesai_pendidikan",
      "foto_url",
      "username",
      "password_hash",
      "status",
      "submitted_at",
      "submitted_ip",
      "reviewed_by",
      "reviewed_at",
      "rejection_reason",
      "created_member_id",
    ],
  },
};

var ROLES = {
  SUPER_ADMIN: "SUPER_ADMIN",
  ADMIN: "ADMIN",
  TIM_PNKB: "TIM_PNKB",
  TIM_ABSENSI: "TIM_ABSENSI",
  PENGAWAS: "PENGAWAS",
  MEMBER: "MEMBER",
};

/* ============ KATEGORI JAMAAH ============ */
/* Urutan dari termuda ke tertua:
 *   BALITA      → < 6 tahun
 *   CABERAWIT   → 6–12 tahun (SD)
 *   PRA_REMAJA  → 13–15 tahun (SMP)
 *   REMAJA      → 16–18 tahun (SMA/SMK)
 *   PRA_NIKAH   → 19+ tahun, belum menikah
 *   DEWASA      → sudah menikah
 *   ISTIMEWA    → sudah menikah & 60+ tahun (sebelumnya "MANULA")
 */
var MEMBER_CATEGORY = {
  BALITA: "BALITA", // ← baru
  CABERAWIT: "CABERAWIT",
  PRA_REMAJA: "PRA_REMAJA",
  REMAJA: "REMAJA",
  PRA_NIKAH: "PRA_NIKAH",
  DEWASA: "DEWASA",
  ISTIMEWA: "ISTIMEWA", // ← rename dari MANULA
};

var ATTENDANCE_STATUS = {
  HADIR: "HADIR",
  IJIN: "IJIN",
  SAKIT: "SAKIT",
  TANPA_KETERANGAN: "TANPA_KETERANGAN",
};

var MONITORING_STATUS = {
  AKTIF: "AKTIF",
  PERLU_PERHATIAN: "PERLU_PERHATIAN",
  KURANG_AKTIF: "KURANG_AKTIF",
  TIDAK_AKTIF: "TIDAK_AKTIF",
};

var ANNOUNCEMENT_STATUS = {
  DRAFT: "DRAFT",
  READY: "READY",
  SHARED: "SHARED",
  CANCELLED: "CANCELLED",
};

var PENDING_STATUS = {
  PENDING: "PENDING",
  APPROVED: "APPROVED",
  REJECTED: "REJECTED",
};

var FIELD_VISIBILITY = {
  UMUM: [
    "member_id",
    "nama_lengkap",
    "nama_panggilan",
    "jenis_kelamin",
    "foto_url",
    "kelompok",
    "status_aktif",
  ],
  INTERNAL: [
    "tempat_lahir",
    "tanggal_lahir",
    "desa",
    "daerah",
    "alamat_rumah",
    "no_wa",
    "pekerjaan",
    "status_pembinaan",
    "tanggal_masuk",
    "tanggal_keluar",
    "created_at",
    "updated_at",
  ],
  PNKB: [
    "is_muballigh",
    "is_kerja",
    "is_nikah",
    "tinggi_badan",
    "berat_badan",
    "hobi",
  ],
  PENDIDIKAN: [
    "jenjang_pendidikan",
    "sekolah",
    "jurusan",
    "tahun_mulai_pendidikan",
    "tahun_selesai_pendidikan",
  ],
};

var AI_DAILY_LIMIT = {
  SUPER_ADMIN: 100,
  ADMIN: 50,
  TIM_PNKB: 30,
  TIM_ABSENSI: 30,
  MEMBER: 10,
};

var ATTENDANCE_CACHE_TTL = 600; // 10 menit
var ATTENDANCE_CACHE_PREFIX = "att_";
var ATTENDANCE_RECENT_LIMIT = 3000; // untuk dashboard (≈ 6 bulan Grup C)
var AUDIT_LOG_RECENT_LIMIT = 500; // audit log tidak perlu semua
var WRITE_ACTIONS = {
  createMember: true,
  updateMember: true,
  deactivateMember: true,
  saveGroup: true,
  createMeeting: true,
  updateMeeting: true,
  deleteMeeting: true,
  saveAttendance: true,
  bulkSaveAttendance: true,
  deleteAttendance: true,
  deleteAttendanceByMeeting: true,
  deleteAttendanceByMember: true,
  createMonitoring: true,
  updateMonitoring: true,
  createAnnouncement: true,
  updateAnnouncement: true,
  uploadPhoto: true,
  deletePhoto: true,
  createUser: true,
  updateUser: true,
  updateUserRole: true,
  updateSettings: true,
  changeMyPassword: true,
  resetUserPassword: true,
  updateMyProfile: true,
  submitPublicRegistration: true,
  approvePendingMember: true,
  rejectPendingMember: true,
  logout: true,
};
var SESSION_CACHE_PREFIX = "sess:";
var USER_SESSION_PREFIX = "usess:";
var SESSION_CACHE_TTL = 21600;
var SESSION_MAX_ROWS_BEFORE_CLEANUP = 500;

var MEMBER_LIST_FIELDS = [
  "member_id",
  "nama_lengkap",
  "nama_panggilan",
  "jenis_kelamin",
  "kelompok",
  "kategori",
  "foto_url",
];

var ATTENDANCE_MEMBER_FIELDS = [
  "member_id",
  "nama_lengkap",
  "kelompok",
  "kategori",
  "jenis_kelamin",
];

var MEMBER_DETAIL_FIELDS = [
  "member_id",
  "nama_lengkap",
  "nama_panggilan",
  "jenis_kelamin",
  "tempat_lahir",
  "tanggal_lahir",
  "foto_url",
  "no_wa",
  "alamat_rumah",
  "desa",
  "daerah",
  "kelompok",
  "is_muballigh",
  "is_kerja",
  "is_nikah",
  "tinggi_badan",
  "berat_badan",
  "hobi",
  "pekerjaan",
  "status_pembinaan",
  "status_aktif",
  "tanggal_masuk",
  "tanggal_keluar",
  "jenjang_pendidikan",
  "sekolah",
  "jurusan",
  "tahun_mulai_pendidikan",
  "tahun_selesai_pendidikan",
  "created_at",
  "updated_at",
  "kategori",
  "usia",
];

var MEMBER_DEFAULT_LIMIT = 100;
var MEMBER_MAX_LIMIT = 500;

var MEMBER_EXPORT_FIELDS = [
  "nama_lengkap",
  "nama_panggilan",
  "jenis_kelamin",
  "tempat_lahir",
  "tanggal_lahir",
  "usia",
  "kategori",
  "kelompok",
  "desa",
  "daerah",
  "alamat_rumah",
  "no_wa",
  "pekerjaan",
  "hobi",
  "status_pembinaan",
];
