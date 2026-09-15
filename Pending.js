var MAX_SUBMISSION_PER_IP_PER_DAY = 3;
var REGISTRATION_MIN_AGE = 0;
var REGISTRATION_MAX_AGE = 120;
var USERNAME_REGEX = /^[a-z0-9_]{3,20}$/;
var MIN_PASSWORD_LENGTH = 6;

/* -------------------------------------------------------------------------- */
/*                          Helper sapaan & doa                               */
/* -------------------------------------------------------------------------- */

/**
 * Hitung usia dari tanggal lahir (ISO date).
 * Return null kalau tidak valid.
 */
function hitungUsia_(tanggalLahir) {
  if (!tanggalLahir) return null;
  var dob = new Date(tanggalLahir);
  if (isNaN(dob.getTime())) return null;
  var today = new Date();
  var usia = today.getFullYear() - dob.getFullYear();
  var m = today.getMonth() - dob.getMonth();
  if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) usia--;
  return usia;
}

/**
 * Sapaan berdasarkan usia + jenis kelamin.
 * - Caberawit (< 13 th)         -> "Adik"
 * - Laki-laki >= 40 th          -> "Bapak"
 * - Perempuan >= 40 th          -> "Ibu"
 * - Lainnya                     -> "Saudara/Saudari"
 */
function buildSapaan_(jenisKelamin, tanggalLahir) {
  var usia = hitungUsia_(tanggalLahir);
  var jk = String(jenisKelamin || "").toUpperCase();

  if (usia !== null && usia < 13) {
    return "Adik";
  }
  if (usia !== null && usia >= 40) {
    return jk === "P" ? "Ibu" : "Bapak";
  }
  return jk === "P" ? "Saudari" : "Saudara";
}

/**
 * Doa penutup sesuai jenis kelamin.
 */
function buildDoa_(jenisKelamin) {
  var jk = String(jenisKelamin || "").toUpperCase();
  return jk === "P" ? "Jazaakillahu khoiro" : "Jazaakallahu khoiro";
}

/* -------------------------------------------------------------------------- */
/*                     Cek ketersediaan username (real-time)                  */
/* -------------------------------------------------------------------------- */

function checkUsernameAvailability_(ctx, params) {
  var username = String(params.username || "")
    .trim()
    .toLowerCase();

  if (!USERNAME_REGEX.test(username)) {
    return ok_({ available: false, reason: "invalid" });
  }

  var usersRepo = new SheetRepository_("users");
  if (usersRepo.findById("username", username)) {
    return ok_({ available: false, reason: "taken" });
  }

  var pendingRepo = new SheetRepository_("pending_members");
  var dup = pendingRepo.find(function (p) {
    return (
      String(p.status).trim().toUpperCase() === PENDING_STATUS.PENDING &&
      String(p.username || "").toLowerCase() === username
    );
  });
  if (dup.length > 0) {
    return ok_({ available: false, reason: "pending" });
  }

  return ok_({ available: true });
}

/* -------------------------------------------------------------------------- */
/*                              Submit pendaftaran                            */
/* -------------------------------------------------------------------------- */

function submitPublicRegistration_(ctx, params) {
  var namaLengkap = String(params.nama_lengkap || "").trim();
  var jenisKelamin = String(params.jenis_kelamin || "")
    .trim()
    .toUpperCase();
  var noWa = String(params.no_wa || "").trim();
  var username = String(params.username || "")
    .trim()
    .toLowerCase();
  var password = String(params.password || "");

  if (namaLengkap.length < 3) {
    return fail_("Nama lengkap minimal 3 karakter");
  }
  if (jenisKelamin !== "L" && jenisKelamin !== "P") {
    return fail_("Jenis kelamin harus L atau P");
  }
  if (!noWa) {
    return fail_("Nomor WhatsApp wajib diisi");
  }

  var normalizedWa = normalizePhoneNumber(noWa);
  if (normalizedWa.length < 10 || normalizedWa.length > 15) {
    return fail_("Nomor WhatsApp tidak valid");
  }

  if (!USERNAME_REGEX.test(username)) {
    return fail_(
      "Username tidak valid. Gunakan huruf kecil, angka, atau underscore (3-20 karakter).",
    );
  }
  if (password.length < MIN_PASSWORD_LENGTH) {
    return fail_("Password minimal " + MIN_PASSWORD_LENGTH + " karakter");
  }

  // Cek username belum dipakai di users
  var usersRepo = new SheetRepository_("users");
  if (usersRepo.findById("username", username)) {
    return fail_("Username sudah dipakai. Coba yang lain.");
  }

  var clientIp = String(params._client_ip || "unknown").substring(0, 50);

  // FASE 2B FIX: hapus SpreadsheetApp.flush() pertama yang sia-sia.
  // Flush hanya perlu setelah semua write selesai (di akhir fungsi).

  var pendingRepo = new SheetRepository_("pending_members");
  var membersRepo = new SheetRepository_("members");

  pendingRepo._invalidateCache();
  membersRepo._invalidateCache();

  var today = formatDate(nowIso_());

  var todayByIp = pendingRepo.find(function (p) {
    return (
      String(p.submitted_ip) === clientIp &&
      String(p.submitted_at).indexOf(today) === 0
    );
  });

  if (todayByIp.length >= MAX_SUBMISSION_PER_IP_PER_DAY) {
    return fail_(
      "Terlalu banyak pendaftaran dari perangkat ini. Coba lagi besok.",
    );
  }

  var allMembers = membersRepo.getAll();
  var existingMember = null;
  for (var i = 0; i < allMembers.length; i++) {
    if (String(allMembers[i].no_wa) === String(normalizedWa)) {
      existingMember = allMembers[i];
      break;
    }
  }
  if (existingMember) {
    return fail_("Nomor WhatsApp sudah terdaftar sebagai jamaah");
  }

  var allPending = pendingRepo.getAll();
  var existingPending = null;
  var usernamePending = null;
  for (var j = 0; j < allPending.length; j++) {
    var p = allPending[j];
    var status = String(p.status).trim().toUpperCase();
    var wa = String(p.no_wa).trim();
    var uname = String(p.username || "").toLowerCase();

    if (wa === normalizedWa && status === PENDING_STATUS.PENDING) {
      existingPending = p;
      break;
    }
    if (uname === username && status === PENDING_STATUS.PENDING) {
      usernamePending = p;
    }
  }
  if (existingPending) {
    return fail_("Pendaftaran dengan nomor ini sedang menunggu verifikasi");
  }
  if (usernamePending) {
    return fail_("Username sedang menunggu verifikasi. Coba yang lain.");
  }

  var now = nowIso_();
  var submissionId = generateSubmissionId();

  var row = {
    submission_id: submissionId,
    nama_lengkap: namaLengkap,
    nama_panggilan: String(params.nama_panggilan || "").trim(),
    jenis_kelamin: jenisKelamin,
    tempat_lahir: String(params.tempat_lahir || "").trim(),
    tanggal_lahir: params.tanggal_lahir ? formatDate(params.tanggal_lahir) : "",
    no_wa: normalizedWa,
    alamat_rumah: String(params.alamat_rumah || "").trim(),
    desa: String(params.desa || "").trim(),
    daerah: String(params.daerah || "").trim(),
    pekerjaan: String(params.pekerjaan || "").trim(),
    hobi: String(params.hobi || "").trim(),
    is_nikah: toBool_(params.is_nikah),
    jenjang_pendidikan: String(params.jenjang_pendidikan || "").trim(),
    sekolah: String(params.sekolah || "").trim(),
    jurusan: String(params.jurusan || "").trim(),
    tahun_mulai_pendidikan: String(params.tahun_mulai_pendidikan || "").trim(),
    tahun_selesai_pendidikan: String(
      params.tahun_selesai_pendidikan || "",
    ).trim(),
    foto_url: String(params.foto_url || "").trim(),
    username: username,
    password_hash: hashPassword_(password),
    status: PENDING_STATUS.PENDING,
    submitted_at: now,
    submitted_ip: clientIp,
    reviewed_by: "",
    reviewed_at: "",
    rejection_reason: "",
    created_member_id: "",
  };

  pendingRepo.insert(row);

  // Flush di akhir (satu-satunya) — pastikan write selesai sebelum return
  SpreadsheetApp.flush();

  invalidateDashboardCache_();

  return ok_({
    submission_id: submissionId,
    nama_lengkap: namaLengkap,
    submitted_at: now,
  });
}

/* -------------------------------------------------------------------------- */
/*                              List & detail                                 */
/* -------------------------------------------------------------------------- */

function getPendingMembers_(ctx, params) {
  var repo = new SheetRepository_("pending_members");
  var all = repo.getAll();

  if (params.status) {
    all = all.filter(function (p) {
      return p.status === params.status;
    });
  }

  all.sort(function (a, b) {
    return new Date(b.submitted_at) - new Date(a.submitted_at);
  });

  var isPaged = String(params.paged) === "true";

  var limit = params.limit !== undefined ? Number(params.limit) : 50;
  if (isNaN(limit) || limit <= 0) limit = 50;
  if (limit > 200) limit = 200;

  var offset = params.offset !== undefined ? Number(params.offset) : 0;
  if (isNaN(offset) || offset < 0) offset = 0;

  var total = all.length;
  var paged = all.slice(offset, offset + limit);

  var items = paged.map(function (p) {
    var c = Object.assign({}, p);
    delete c._row;
    delete c.password_hash;
    return c;
  });

  if (!isPaged) {
    return ok_(items);
  }

  return ok_({
    items: items,
    total: total,
    limit: limit,
    offset: offset,
    has_more: offset + items.length < total,
  });
}

function getPendingMemberDetail_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var repo = new SheetRepository_("pending_members");
  var pending = repo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");

  var c = Object.assign({}, pending);
  delete c._row;
  delete c.password_hash;
  return ok_(c);
}

/* -------------------------------------------------------------------------- */
/*                                  Approve                                   */
/* -------------------------------------------------------------------------- */

function approvePendingMember_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var pendingRepo = new SheetRepository_("pending_members");
  var pending = pendingRepo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");
  if (String(pending.status).toUpperCase() !== PENDING_STATUS.PENDING) {
    return fail_("Pendaftaran sudah diproses");
  }

  var username = String(pending.username || "")
    .trim()
    .toLowerCase();
  var passwordHash = String(pending.password_hash || "");

  if (!USERNAME_REGEX.test(username)) {
    return fail_("Data pendaftar tidak memiliki username yang valid");
  }
  if (!passwordHash) {
    return fail_("Data pendaftar tidak memiliki password");
  }

  var usersRepo = new SheetRepository_("users");
  if (usersRepo.findById("username", username)) {
    return fail_(
      "Username sudah dipakai. Tolak pendaftar dan minta daftar ulang dengan username lain.",
    );
  }

  var membersRepo = new SheetRepository_("members");
  var now = nowIso_();
  var memberId = generateMemberId();

  var memberRow = {
    member_id: memberId,
    nama_lengkap: pending.nama_lengkap,
    nama_panggilan: pending.nama_panggilan || "",
    jenis_kelamin: pending.jenis_kelamin,
    tempat_lahir: pending.tempat_lahir || "",
    tanggal_lahir: pending.tanggal_lahir || "",
    foto_url: pending.foto_url || "",
    no_wa: pending.no_wa || "",
    alamat_rumah: pending.alamat_rumah || "",
    desa: pending.desa || "",
    daerah: pending.daerah || "",
    kelompok: params.kelompok || "",
    is_muballigh: false,
    is_kerja: false,
    is_nikah: toBool_(pending.is_nikah),
    tinggi_badan: "",
    berat_badan: "",
    hobi: pending.hobi || "",
    pekerjaan: pending.pekerjaan || "",
    status_pembinaan: MONITORING_STATUS.AKTIF,
    status_aktif: true,
    tanggal_masuk: formatDate(now),
    tanggal_keluar: "",
    jenjang_pendidikan: pending.jenjang_pendidikan || "",
    sekolah: pending.sekolah || "",
    jurusan: pending.jurusan || "",
    tahun_mulai_pendidikan: pending.tahun_mulai_pendidikan || "",
    tahun_selesai_pendidikan: pending.tahun_selesai_pendidikan || "",
    created_at: now,
    updated_at: now,
  };
  membersRepo.insert(memberRow);

  var userId = generateUserId();
  usersRepo.insert({
    user_id: userId,
    username: username,
    password_hash: passwordHash,
    nama: pending.nama_lengkap,
    role: ROLES.MEMBER,
    member_id: memberId,
    status_aktif: true,
    created_at: now,
    updated_at: now,
    last_login_at: "",
  });

  pendingRepo.updateById("submission_id", params.submission_id, {
    status: PENDING_STATUS.APPROVED,
    reviewed_by: ctx.user.user_id,
    reviewed_at: now,
    created_member_id: memberId,
  });

  invalidateDashboardCache_();

  writeAuditLog_(
    ctx.user.user_id,
    "APPROVE_PENDING",
    "PENDING",
    params.submission_id,
  );
  writeAuditLog_(ctx.user.user_id, "CREATE_USER_FROM_PENDING", "USER", userId);

  if (pending.no_wa) {
    var sapaan = buildSapaan_(pending.jenis_kelamin, pending.tanggal_lahir);
    var doa = buildDoa_(pending.jenis_kelamin);

    try {
      sendWhatsApp_(
        pending.no_wa,
        "Assalamu'alaikum " +
          sapaan +
          ",\n\n" +
          "Alhamdulillah, pendaftaran " +
          sapaan +
          " di *Sambung Ngaji* sudah disetujui.\n\n" +
          "Silakan masuk menggunakan:\n" +
          "Username: *" +
          username +
          "*\n" +
          "Password: sesuai yang " +
          sapaan +
          " daftarkan\n\n" +
          "Barakallahu fiik.\n" +
          doa +
          " \uD83E\uDD0D",
      );
    } catch (e) {
      Logger.log("Gagal kirim WA approve: " + e);
    }
  }

  return ok_({
    member_id: memberId,
    submission_id: params.submission_id,
    user_id: userId,
    username: username,
  });
}

/* -------------------------------------------------------------------------- */
/*                                  Reject                                    */
/* -------------------------------------------------------------------------- */

function rejectPendingMember_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var pendingRepo = new SheetRepository_("pending_members");
  var pending = pendingRepo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");
  if (String(pending.status).toUpperCase() !== PENDING_STATUS.PENDING) {
    return fail_("Pendaftaran sudah diproses");
  }

  var now = nowIso_();
  var reason = String(params.reason || "Tidak memenuhi syarat").trim();

  pendingRepo.updateById("submission_id", params.submission_id, {
    status: PENDING_STATUS.REJECTED,
    reviewed_by: ctx.user.user_id,
    reviewed_at: now,
    rejection_reason: reason,
  });

  invalidateDashboardCache_();

  writeAuditLog_(
    ctx.user.user_id,
    "REJECT_PENDING",
    "PENDING",
    params.submission_id,
  );

  if (pending.no_wa) {
    var sapaan = buildSapaan_(pending.jenis_kelamin, pending.tanggal_lahir);
    var doa = buildDoa_(pending.jenis_kelamin);

    try {
      sendWhatsApp_(
        pending.no_wa,
        "Assalamu'alaikum " +
          sapaan +
          ",\n\n" +
          "Mohon maaf, pendaftaran " +
          sapaan +
          " di *Sambung Ngaji* belum bisa kami setujui.\n\n" +
          "Alasan: " +
          reason +
          "\n\n" +
          "Silakan hubungi admin untuk informasi lebih lanjut.\n\n" +
          doa +
          " \uD83D\uDE4F",
      );
    } catch (e) {
      Logger.log("Gagal kirim WA reject: " + e);
    }
  }

  return ok_({
    submission_id: params.submission_id,
    status: PENDING_STATUS.REJECTED,
  });
}
