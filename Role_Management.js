/**
 * ROLE_PERMISSIONS — Daftar izin aksi per role.
 *
 * Setiap role punya daftar action yang boleh diakses.
 * - SUPER_ADMIN: ['*'] = akses penuh ke semua action
 * - Action yang tidak ada di daftar = forbidden (ditolak)
 *
 * Kategorisasi action:
 * - READ:     get*, list — hanya lihat data
 * - WRITE:    create*, update*, save*, bulk* — mengubah data
 * - DELETE:   delete*, deactivate* — menghapus/nonaktifkan data
 * - MONITOR:  createMonitoring, updateMonitoring — catatan pembinaan
 * - SESSION:  validateSession, changeMyPassword — autentikasi
 */
var ROLE_PERMISSIONS = {
  /**
   * SUPER_ADMIN — Akses penuh.
   * Satu-satunya yang bisa: kelola user, ubah pengaturan, lihat audit log.
   */
  SUPER_ADMIN: ["*"],

  /**
   * ADMIN — Pengurus harian.
   * Bisa: kelola jamaah, kelompok, jadwal, absensi, pengumuman, approval pendaftar.
   * Tidak bisa: kelola user, ubah pengaturan, lihat audit log (khusus SUPER_ADMIN).
   */
  ADMIN: [
    // — Sesi & keamanan
    "validateSession",
    "changeMyPassword",

    // — Jamaah: baca, tulis, nonaktifkan
    "getDashboard",
    "getMembers",
    "getMembersPaged",
    "getMembersForExport",
    "getMemberDetail",
    "createMember",
    "updateMember",
    "deactivateMember",
    "getMemberUserStatus",
    "getUserDetail",

    // — Kelompok: baca & tulis
    "getGroups",
    "saveGroup",

    // — Jadwal pengajian: baca, buat, ubah
    "getMeetings",
    "createMeeting",
    "updateMeeting",
    "deleteMeeting",

    // — Bulk Meeting
    "previewBulkMeetings",
    "bulkCreateMeetings",
    "getBulkMeetingTemplates",

    // — Absensi: baca, input, bulk, hapus (per record / per meeting / per member)
    "getAttendance",
    "getAttendancePage",
    "saveAttendance",
    "bulkSaveAttendance",
    "deleteAttendance",
    "deleteAttendanceByMeeting",
    "deleteAttendanceByMember",

    // — Monitoring: baca, tulis, ubah
    "getMonitoring",
    "createMonitoring",
    "updateMonitoring",

    // — Pengumuman: baca template, buat, ubah, generate
    "getAnnouncementTemplates",
    "createAnnouncement",
    "updateAnnouncement",
    "generateAnnouncement",

    // — Foto & pengaturan
    "uploadPhoto",
    "deletePhoto",
    "getSettings",

    // — AI
    "aiChat",
    "getCurrentProvider",
    "setAIProvider",

    // — Pendaftar: verifikasi & approval
    "getPendingMembers",
    "getPendingMemberDetail",
    "approvePendingMember",
    "rejectPendingMember",

    // — Monitoring AI usage
    "getAiUsageStats",
  ],

  /**
   * TIM_PNKB — Tim Pembinaan Pra Nikah.
   * Bisa: lihat jamaah PNKB saja, tulis monitoring untuk mereka.
   * Tidak bisa: kelola jamaah, absensi, kelompok, pengumuman.
   */
  TIM_PNKB: [
    // — Sesi & keamanan
    "validateSession",
    "changeMyPassword",

    // — Dashboard & jamaah PNKB saja (getPNKBMembers = filter otomatis)
    "getDashboard",
    "getPNKBMembers",
    "getMembersForExport",
    "getPNKBMembersPaged",
    "getMemberDetail",

    // — Monitoring: fokus utama role ini
    "getMonitoring",
    "createMonitoring",
    "updateMonitoring",

    // — AI
    "aiChat",
    "getCurrentProvider",
    "setAIProvider",
  ],

  /**
   * TIM_ABSENSI — Tim Absensi Pengajian.
   * Bisa: kelola jadwal pengajian & input absensi.
   * Tidak bisa: kelola jamaah, monitoring, pengumuman.
   */
  TIM_ABSENSI: [
    // — Sesi & keamanan
    "validateSession",
    "changeMyPassword",

    // — Dashboard & jadwal pengajian
    "getMeetings",
    "createMeeting",
    "updateMeeting",
    "deleteMeeting",

    // — Absensi: fokus utama role ini
    "getAttendanceMembers",
    "getAttendance",
    "getAttendancePage",
    "saveAttendance",
    "bulkSaveAttendance",
    "deleteAttendance",
    "deleteAttendanceByMeeting",
    "deleteAttendanceByMember",

    // — AI
    "aiChat",
    "getCurrentProvider",
    "setAIProvider",
  ],

  /**
   * PENGAWAS — Pengawas / Viewer.
   * Bisa: LIHAT semua data (jamaah, kelompok, jadwal, absensi, pengumuman) + TULIS monitoring.
   * Tidak bisa: create/update/delete data apapun selain monitoring.
   * Cocok untuk: pengawas lapangan, yayasan, donatur, pembina tamu.
   */
  PENGAWAS: [
    // — Sesi & keamanan
    "validateSession",
    "changeMyPassword",

    // — Dashboard
    "getDashboard",

    // — Jamaah: hanya lihat (tidak ada create/update/deactivate)
    "getMembers",
    "getMembersPaged",
    "getMemberDetail",

    // — Kelompok: hanya lihat (tidak ada saveGroup)
    "getGroups",

    // — Jadwal pengajian: hanya lihat (tidak ada create/update)
    "getMeetings",

    // — Absensi: hanya lihat (tidak ada save/bulk/delete)
    "getAttendance",
    "getAttendancePage",
    "getAttendanceMembers",

    // — Monitoring: SATU-SATUNYA write yang diizinkan
    "getMonitoring",
    "createMonitoring",
    "updateMonitoring",

    // — Pengumuman: hanya lihat
    "getAnnouncementTemplates",
    "getAnnouncements",

    // — AI (read-only, tidak bisa ganti provider)
    "aiChat",
    "getCurrentProvider",
  ],

  /**
   * MEMBER — Jamaah biasa.
   * Bisa: lihat & edit profil sendiri, lihat absensi & monitoring sendiri, chat AI.
   * Tidak bisa: lihat data jamaah lain, akses fitur admin.
   */
  MEMBER: [
    // — Sesi & keamanan
    "validateSession",
    "changeMyPassword",

    // — Dashboard pribadi (getMyDashboard = data diri sendiri)
    "getMyDashboard",

    // — Profil sendiri: lihat & edit (field terbatas)
    "getMyProfile",
    "updateMyProfile",

    // — Riwayat sendiri
    "getMyAttendance",
    "getMyMonitoring",

    // — Jadwal pengajian mendatang
    "getUpcomingMeetings",

    // — Foto profil sendiri
    "uploadPhoto",
    "deletePhoto",

    // — AI (tidak bisa ganti provider)
    "aiChat",
    "getCurrentProvider",
  ],
};
