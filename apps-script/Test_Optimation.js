function postDeploySmokeTest() {
  Logger.log("=== POST-DEPLOY SMOKE TEST ===");
  Logger.log("");

  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  if (!login.success) {
    Logger.log("❌ Login gagal: " + login.message);
    return;
  }
  var ctx = validateSession_(login.data.token);
  Logger.log("✅ Login OK");

  // Test 1: Create meeting (verify getHariFromDate fix)
  var meetingResult = createMeeting_(ctx, {
    tanggal: formatDate(nowIso_()),
    acara: "SMOKE TEST " + Date.now(),
    kategori_target: [],
  });
  if (!meetingResult.success) {
    Logger.log("❌ Create meeting gagal: " + meetingResult.message);
    return;
  }
  var testMeetingId = meetingResult.data.meeting_id;
  Logger.log("✅ Create meeting OK: " + testMeetingId);
  Logger.log("   Hari: " + meetingResult.data.hari);

  // Test 2: Bulk save attendance
  var members = getMembers_(ctx, { limit: 5 });
  var items = members.data.map(function (m) {
    return { member_id: m.member_id, status: "HADIR" };
  });
  var bulkResult = bulkSaveAttendance_(ctx, {
    meeting_id: testMeetingId,
    items: items,
  });
  if (!bulkResult.success) {
    Logger.log("❌ Bulk save gagal: " + bulkResult.message);
    return;
  }
  Logger.log("✅ Bulk save OK: " + JSON.stringify(bulkResult.data));

  // Test 3: Delete meeting (verify deleteMeeting_ + deleteAttendanceByMeeting_)
  var t1 = Date.now();
  var deleteResult = deleteMeeting_(ctx, { meeting_id: testMeetingId });
  var elapsed = Date.now() - t1;
  Logger.log("✅ Delete meeting OK: " + JSON.stringify(deleteResult.data));
  Logger.log("   Time: " + elapsed + "ms");

  // Test 4: AI chat quota (verify incrementAiQuota_)
  var key = "aiquota:" + ctx.user.user_id + ":" + formatDate(nowIso_());
  PropertiesService.getScriptProperties().setProperty(key, "0");
  var chatResult = handleAiChat_({ message: "Halo", history: [] }, ctx);
  var quotaAfter = getAiQuotaUsed_(ctx);
  Logger.log(
    "✅ AI chat: success=" + chatResult.success + ", quota=" + quotaAfter,
  );

  Logger.log("");
  Logger.log("=== SEMUA TEST LULUS ===");
}

function preDeployCheck() {
  Logger.log("=== PRE-DEPLOY CHECK ===");
  Logger.log("");

  Logger.log("1. Environment:");
  showEnvironmentConfig();
  Logger.log("");

  Logger.log("2. Validasi environment:");
  var v = validateEnvironmentConfig();
  Logger.log("   Success: " + v.success);
  if (!v.success) {
    Logger.log("   Errors: " + JSON.stringify(v.errors));
  }
  Logger.log("");

  Logger.log("3. Trigger aktif:");
  listAllTriggers();
  Logger.log("");

  Logger.log("4. Fungsi penting ada:");
  var fns = [
    "login_",
    "validateSession_",
    "handleAiChat_",
    "checkAiQuota_",
    "incrementAiQuota_",
    "cleanupOldAiQuota_",
    "submitPublicRegistration_",
    "deleteAttendanceByMeeting_",
    "deleteMeeting_",
    "createMeeting_",
  ];
  fns.forEach(function (fn) {
    var exists = "❌ TIDAK ADA";
    try {
      if (typeof eval(fn) === "function") exists = "✅ ADA";
    } catch (e) {
      exists = "❌ ERROR";
    }
    Logger.log("   " + fn + ": " + exists);
  });
  Logger.log("");

  Logger.log("5. Ukuran sheet penting:");
  [
    "sessions",
    "attendance",
    "audit_logs",
    "ai_usage",
    "pending_members",
  ].forEach(function (k) {
    var sheet = new SheetRepository_(k)._sheet();
    Logger.log("   " + k + ": " + (sheet.getLastRow() - 1) + " baris");
  });
  Logger.log("");

  Logger.log("=== SELESAI ===");
}

/**
 * AUDIT UNDEFINED FUNCTIONS
 *
 * Scan semua fungsi global, cari panggilan fungsi yang:
 * - Terlihat seperti fungsi project (akhiran _ atau format camelCase)
 * - Tidak ada di daftar fungsi global
 * - Tidak ada di whitelist built-in GAS
 *
 * Berguna untuk nemu typo seperti getHariFromDate_ vs getHariFromDate.
 */
function auditUndefinedFunctions() {
  Logger.log("=== AUDIT UNDEFINED FUNCTIONS ===");
  Logger.log("");

  // 1. Kumpulkan semua fungsi global yang ada di project
  var globalFunctions = {};
  var allKeys = Object.keys(this);
  allKeys.forEach(function (k) {
    if (typeof this[k] === "function") {
      globalFunctions[k] = true;
    }
  });

  var totalFunctions = Object.keys(globalFunctions).length;
  Logger.log("Total fungsi global terdeteksi: " + totalFunctions);
  Logger.log("");

  // 2. Whitelist: built-in GAS & JS
  var whitelist = {
    // JS built-in
    Object: true,
    Array: true,
    String: true,
    Number: true,
    Boolean: true,
    Date: true,
    Math: true,
    JSON: true,
    RegExp: true,
    Error: true,
    Promise: true,
    Map: true,
    Set: true,
    parseInt: true,
    parseFloat: true,
    isNaN: true,
    isFinite: true,
    encodeURIComponent: true,
    decodeURIComponent: true,
    encodeURI: true,
    decodeURI: true,
    eval: true,
    Function: true,
    require: true,
    console: true,

    // GAS built-in
    Logger: true,
    Utilities: true,
    SpreadsheetApp: true,
    DriveApp: true,
    Drive: true,
    GmailApp: true,
    MailApp: true,
    UrlFetchApp: true,
    CacheService: true,
    PropertiesService: true,
    LockService: true,
    ScriptApp: true,
    Session: true,
    ContentService: true,
    HtmlService: true,
    CalendarApp: true,
    Charts: true,
    Maps: true,
    XmlService: true,
    SitesApp: true,
    FormApp: true,
    DocumentApp: true,
    SlidesApp: true,
    translate: true,
    BigQuery: true,
    Firebase: true,

    // Common patterns
    toString: true,
    valueOf: true,
    hasOwnProperty: true,
    isPrototypeOf: true,
    propertyIsEnumerable: true,
    toLocaleString: true,
    constructor: true,

    // JS keywords yang sering salah dideteksi
    if: true,
    for: true,
    while: true,
    switch: true,
    function: true,
    return: true,
    typeof: true,
    catch: true,
    try: true,
    throw: true,
    new: true,
    delete: true,
    void: true,
    do: true,
    else: true,
    in: true,
    instanceof: true,
  };

  // 3. Untuk setiap fungsi, scan source code-nya
  var undefinedCalls = {}; // fnName -> [caller1, caller2, ...]

  Object.keys(globalFunctions).forEach(function (fnName) {
    var source = "";
    try {
      source =
        globalFunctions[fnName] && globalFunctions[fnName].toString ? "" : "";
      // Ambil source code
      source = this[fnName].toString();
    } catch (e) {
      return;
    }

    // Regex untuk panggilan fungsi: nama_fungsi(
    // Match identifier yang diikuti langsung tanda '('
    var callRegex = /\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(/g;
    var match;
    while ((match = callRegex.exec(source)) !== null) {
      var calledName = match[1];

      // Skip kalau di whitelist
      if (whitelist[calledName]) continue;

      // Skip kalau bukan pattern project (harus ada _ di akhir atau camelCase)
      // Pattern: diakhiri _, atau ada huruf besar setelah huruf kecil (camelCase)
      var looksLikeProjectFn =
        /_$/.test(calledName) || // diakhiri _
        /^[a-z]+[A-Z]/.test(calledName) || // camelCase
        /^[A-Z][a-z]+[A-Z]/.test(calledName); // PascalCase compound

      if (!looksLikeProjectFn) continue;

      // Skip kalau ternyata fungsi ini ada
      if (globalFunctions[calledName]) continue;

      // Catat sebagai undefined
      if (!undefinedCalls[calledName]) {
        undefinedCalls[calledName] = [];
      }
      if (undefinedCalls[calledName].indexOf(fnName) === -1) {
        undefinedCalls[calledName].push(fnName);
      }
    }
  });

  // 4. Report
  var undefinedList = Object.keys(undefinedCalls);
  Logger.log("=== HASIL ===");
  Logger.log("Fungsi dipanggil tapi TIDAK ADA: " + undefinedList.length);
  Logger.log("");

  if (undefinedList.length === 0) {
    Logger.log("✅ Tidak ditemukan typo/panggilan undefined.");
  } else {
    undefinedList.sort().forEach(function (name) {
      Logger.log("  ❌ " + name + "()");
      Logger.log("     Dipanggil dari: " + undefinedCalls[name].join(", "));
    });
  }

  Logger.log("");
  Logger.log("=== CATATAN ===");
  Logger.log(
    "- False positive mungkin terjadi (misal variabel lokal yang dianggap fungsi)",
  );
  Logger.log("- Verifikasi manual sebelum fix");

  return {
    totalFunctions: totalFunctions,
    undefinedCalls: undefinedCalls,
  };
}

function testDeleteAttendanceOptimization() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var meetingsRepo = new SheetRepository_("meetings");
  var attendanceRepo = new SheetRepository_("attendance");

  // Buat meeting test
  var meetingResult = createMeeting_(ctx, {
    tanggal: formatDate(nowIso_()),
    acara: "TEST OPTIMASI DELETE",
    kategori_target: [],
  });
  var testMeetingId = meetingResult.data.meeting_id;
  Logger.log("Meeting test: " + testMeetingId);

  // Ambil 30 member
  var members = getMembers_(ctx, { limit: 30 });
  if (members.data.length < 10) {
    Logger.log("Member tidak cukup untuk test");
    return;
  }
  Logger.log("Member tersedia: " + members.data.length);

  // Bulk save 30 attendance
  var items = members.data.map(function (m) {
    return { member_id: m.member_id, status: "HADIR" };
  });
  bulkSaveAttendance_(ctx, { meeting_id: testMeetingId, items: items });
  Logger.log("Attendance dibuat: " + items.length);

  // Ukur delete time
  var t1 = Date.now();
  var result = deleteAttendanceByMeeting_(ctx, { meeting_id: testMeetingId });
  var elapsed = Date.now() - t1;

  Logger.log("");
  Logger.log("=== HASIL ===");
  Logger.log("Deleted: " + result.data.deleted + " rows");
  Logger.log("Time: " + elapsed + "ms");
  Logger.log("");
  Logger.log("Baseline (24 rows): 1576ms");
  Logger.log("Target (30 rows): < 1000ms");

  // Cleanup: hapus meeting test
  meetingsRepo.deleteById("meeting_id", testMeetingId);
  Logger.log("Cleanup selesai");
}

function testRegistrationSpeed() {
  Logger.log("=== TEST REGISTRASI SPEED ===");

  var t1 = Date.now();
  var result = submitPublicRegistration_(null, {
    nama_lengkap: "Test Reg " + Date.now(),
    jenis_kelamin: "L",
    no_wa: "62812" + String(Date.now()).slice(-8),
    username: "testreg_" + String(Date.now()).slice(-6),
    password: "test123456",
    _client_ip: "127.0.0.1",
  });
  var elapsed = Date.now() - t1;

  Logger.log("Success: " + result.success);
  Logger.log("Time: " + elapsed + "ms");
  Logger.log("(Sebelum fix: ~1500-3000ms, sesudah fix: ~500-1500ms)");

  // Cleanup — hapus submission test
  if (result.success) {
    var repo = new SheetRepository_("pending_members");
    var sub = repo.findById("submission_id", result.data.submission_id);
    if (sub) {
      repo.deleteById("submission_id", result.data.submission_id);
      Logger.log("Cleanup: submission test dihapus");
    }
  }
}

function testAuditJsFix() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  Logger.log("=== SEBELUM increment ===");
  Logger.log("checkAiQuota_: " + JSON.stringify(checkAiQuota_(ctx)));
  Logger.log("getAiQuotaUsed_: " + getAiQuotaUsed_(ctx));

  incrementAiQuota_(ctx);
  incrementAiQuota_(ctx);
  incrementAiQuota_(ctx);

  Logger.log("");
  Logger.log("=== SESUDAH 3x increment ===");
  Logger.log("getAiQuotaUsed_: " + getAiQuotaUsed_(ctx));
  Logger.log("checkAiQuota_: " + JSON.stringify(checkAiQuota_(ctx)));

  Logger.log("");
  Logger.log("=== TEST BLOKIR (set kuota = limit) ===");
  var props = PropertiesService.getScriptProperties();
  var key = "aiquota:" + ctx.user.user_id + ":" + formatDate(nowIso_());
  var limit = AI_DAILY_LIMIT[ctx.user.role];
  props.setProperty(key, String(limit));
  Logger.log("Set counter = " + limit);
  Logger.log("checkAiQuota_: " + JSON.stringify(checkAiQuota_(ctx)));

  Logger.log("");
  Logger.log("=== CLEANUP ===");
  props.setProperty(key, "0");
  Logger.log("Reset counter ke 0");
}

function testAiQuotaEndToEnd() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  // Reset kuota user ini
  var key = "aiquota:" + ctx.user.user_id + ":" + formatDate(nowIso_());
  PropertiesService.getScriptProperties().setProperty(key, "0");

  Logger.log("=== Sebelum chat ===");
  Logger.log("Kuota terpakai: " + getAiQuotaUsed_(ctx));
  Logger.log("checkAiQuota_: " + JSON.stringify(checkAiQuota_(ctx)));

  Logger.log("");
  Logger.log("=== Kirim 1 pesan AI ===");
  var result = handleAiChat_({ message: "Halo, apa kabar?", history: [] }, ctx);
  Logger.log("Success: " + result.success);
  Logger.log(
    "Reply: " + (result.data ? String(result.data.reply).slice(0, 100) : "-"),
  );

  Logger.log("");
  Logger.log("=== Setelah chat ===");
  Logger.log("Kuota terpakai: " + getAiQuotaUsed_(ctx));
  Logger.log("checkAiQuota_: " + JSON.stringify(checkAiQuota_(ctx)));
}
function testAiQuotaEnforcement() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var today = formatDate(nowIso_());
  var cacheKey = "aiquota:" + ctx.user.user_id + ":" + today;
  var cache = CacheService.getScriptCache();

  // Clear kuota
  cache.remove(cacheKey);
  Logger.log("Kuota awal: " + (cache.get(cacheKey) || "0"));

  // Panggil increment 3x manual
  incrementAiQuota_(ctx);
  incrementAiQuota_(ctx);
  incrementAiQuota_(ctx);

  Logger.log("Setelah 3x increment: " + cache.get(cacheKey));

  // Check apakah diblokir (limit MEMBER = 10, tapi kita cek di bawah limit dulu)
  var check = checkAiQuota_(ctx);
  Logger.log(
    "checkAiQuota_: " +
      (check ? "BLOCKED: " + check.message : "OK (belum limit)"),
  );

  // Test TTL
  Logger.log("");
  Logger.log("=== TEST TTL ===");
  cache.put("test_ttl", "value", 86400);
  var retrieved = cache.get("test_ttl");
  Logger.log("Put dengan TTL 86400, get: " + (retrieved ? "ADA" : "KOSONG"));
  Logger.log(
    "(Catatan: kalau TTL max 21600, akan expire setelah 6 jam, bukan 24 jam)",
  );
}

function diagnoseAiFunctions() {
  Logger.log("=== DIAGNOSE AI FUNCTIONS ===");
  Logger.log("");

  var functions = [
    "checkAiQuota_",
    "incrementAiQuota_",
    "logAiUsage_",
    "getAiUsageToday_",
    "getAiUsageStats_",
    "handleAiChat_",
  ];

  functions.forEach(function (fn) {
    var type = "TIDAK ADA";
    try {
      var result = eval("typeof " + fn);
      if (result === "function") type = "ADA (function)";
      else type = "TIDAK ADA (" + result + ")";
    } catch (e) {
      type = "ERROR: " + e.message;
    }
    Logger.log("  " + fn + ": " + type);
  });

  Logger.log("");
  Logger.log("=== PANGGIL checkAiQuota_ ===");
  try {
    var login = login_({ username: "albetsurya", password: "albetsurya123" });
    if (!login.success) {
      Logger.log("Login gagal: " + login.message);
      return;
    }
    var ctx = validateSession_(login.data.token);
    if (!ctx) {
      Logger.log("Session invalid");
      return;
    }

    var result = checkAiQuota_(ctx);
    Logger.log("  checkAiQuota_ return: " + JSON.stringify(result));
    Logger.log(
      "  Tipe: " + (result === null ? "null (tidak diblokir)" : typeof result),
    );
  } catch (e) {
    Logger.log("  ERROR: " + e.message);
  }

  Logger.log("");
  Logger.log("=== PANGGIL incrementAiQuota_ ===");
  try {
    var login2 = login_({ username: "albetsurya", password: "albetsurya123" });
    var ctx2 = validateSession_(login2.data.token);
    incrementAiQuota_(ctx2);
    Logger.log("  incrementAiQuota_ BERHASIL dipanggil");

    var today = formatDate(nowIso_());
    var key = "aiquota:" + ctx2.user.user_id + ":" + today;
    var val = CacheService.getScriptCache().get(key);
    Logger.log("  Cache value setelah increment: " + (val || "(kosong)"));
  } catch (e) {
    Logger.log("  ERROR: " + e.message);
  }

  Logger.log("");
  Logger.log("=== DAFTAR FILE DI PROJECT ===");
  try {
    var files = DriveApp.getFolderById("root").getFiles();
    Logger.log("  (tidak bisa enumerate via DriveApp - cek manual di editor)");
  } catch (e) {}
}

function testNormalisasiJenisKelamin() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var members = getMembers_(ctx, { limit: 5 });
  Logger.log("=== getMembers_ ===");
  members.data.forEach(function (m) {
    Logger.log("  " + m.nama_lengkap + " → '" + m.jenis_kelamin + "'");
  });

  var exportData = getMembersForExport_(ctx, {});
  Logger.log("");
  Logger.log("=== getMembersForExport_ (3 sample) ===");
  exportData.data.slice(0, 3).forEach(function (m) {
    Logger.log("  " + m.nama_lengkap);
    Logger.log("    jenis_kelamin: '" + m.jenis_kelamin + "'");
    Logger.log("    tanggal_lahir: '" + m.tanggal_lahir + "'");
  });
}

function migrateJenisKelaminToShort() {
  Logger.log("=== MIGRATE JENIS KELAMIN → L/P ===");

  var sheets = ["members", "pending_members"];

  sheets.forEach(function (sheetKey) {
    var repo = new SheetRepository_(sheetKey);
    var all = repo.getAll();
    var sheet = repo._sheet();
    var headers = repo.def.headers;
    var colIndex = headers.indexOf("jenis_kelamin");

    if (colIndex === -1) {
      Logger.log("❌ Kolom jenis_kelamin tidak ada di " + sheetKey);
      return;
    }

    var migrated = 0;
    var skipped = 0;

    all.forEach(function (row) {
      var raw = String(row.jenis_kelamin || "").trim();
      var normalized = "";

      var lower = raw.toLowerCase();
      if (
        lower === "l" ||
        lower === "laki-laki" ||
        lower === "laki laki" ||
        lower === "pria" ||
        lower === "male"
      ) {
        normalized = "L";
      } else if (
        lower === "p" ||
        lower === "perempuan" ||
        lower === "wanita" ||
        lower === "female"
      ) {
        normalized = "P";
      }

      if (!normalized) {
        skipped++;
        if (raw)
          Logger.log(
            "  ⚠️ Skip (tidak dikenal): '" + raw + "' di " + row.member_id ||
              row.submission_id,
          );
        return;
      }

      if (normalized === raw) {
        skipped++;
        return;
      }

      // Update cell langsung (lebih cepat dari updateById loop)
      sheet.getRange(row._row, colIndex + 1).setValue(normalized);
      migrated++;
    });

    repo._invalidateCache();
    Logger.log(
      "📊 " +
        sheetKey +
        ": " +
        migrated +
        " dimigrasi, " +
        skipped +
        " di-skip",
    );
  });

  Logger.log("");
  Logger.log("=== SELESAI ===");
  Logger.log("Jalankan testVerifyJenisKelamin() untuk verifikasi.");
}

function testVerifyJenisKelamin() {
  Logger.log("=== VERIFIKASI JENIS KELAMIN ===");

  var sheets = ["members", "pending_members"];

  sheets.forEach(function (sheetKey) {
    var repo = new SheetRepository_(sheetKey);
    repo._invalidateCache();
    var all = repo.getAll();
    var variants = {};

    all.forEach(function (row) {
      var jk = String(row.jenis_kelamin || "").trim();
      variants[jk] = (variants[jk] || 0) + 1;
    });

    Logger.log("");
    Logger.log("📊 " + sheetKey + ":");
    Object.keys(variants)
      .sort()
      .forEach(function (k) {
        Logger.log("  '" + k + "': " + variants[k]);
      });
  });
}

function testJenisKelaminRaw() {
  var repo = new SheetRepository_("members");
  var all = repo.getAll();
  var sample = all[0];
  Logger.log("Raw jenis_kelamin: '" + sample.jenis_kelamin + "'");
  Logger.log("Type: " + typeof sample.jenis_kelamin);
  Logger.log("Length: " + String(sample.jenis_kelamin).length);
}

function testPhase6Export() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var t1 = Date.now();
  var result = getMembersForExport_(ctx, {});
  var elapsed = Date.now() - t1;

  Logger.log("Total members: " + result.data.length);
  Logger.log("Time: " + elapsed + "ms");

  var json = JSON.stringify(result);
  Logger.log(
    "Payload: " +
      json.length +
      " bytes (" +
      Math.round(json.length / 1024) +
      " KB)",
  );

  if (result.data.length > 0) {
    Logger.log("Sample fields: " + Object.keys(result.data[0]).join(", "));
    Logger.log("Sample data: " + JSON.stringify(result.data[0]));
  }
}

function auditInvalidationCoverage() {
  Logger.log("=== AUDIT INVALIDATION COVERAGE ===");

  var files = [
    "Members.js",
    "Attendance.js",
    "Meetings.js",
    "Monitoring.js",
    "Pending.js",
    "Groups.js",
    "Announcements.js",
    "Settings.js",
    "Users.js",
    "Profile_Management.js",
    "Auth.js",
  ];

  // Fungsi yang HARUS punya invalidateDashboardCache_
  var requiredFunctions = {
    createMember_: true,
    updateMember_: true,
    deactivateMember_: true,
    updateMyProfile_: true,
    saveAttendance_: true,
    bulkSaveAttendance_: true,
    deleteAttendance_: true,
    deleteAttendanceByMeeting_: true,
    deleteAttendanceByMember_: true,
    createMeeting_: true,
    updateMeeting_: true,
    deleteMeeting_: true,
    createMonitoring_: true,
    updateMonitoring_: true,
    approvePendingMember_: true,
    rejectPendingMember_: true,
    submitPublicRegistration_: true,
    saveGroup_: true,
    createAnnouncement_: true,
    updateAnnouncement_: true,
    updateSettings_: true,
    createUser_: true,
    updateUser_: true,
    updateUserRole_: true,
    uploadPhoto_: true,
    deletePhoto_: true,
  };

  // Fungsi yang TIDAK perlu invalidateDashboardCache_
  var excludedFunctions = {
    changeMyPassword_: true,
    resetUserPassword_: true,
    login_: true,
    logout_: true,
    getGeneralDashboard_: true,
    getPNKBDashboard_: true,
    getAbsensiDashboard_: true,
    getMyDashboard_: true,
  };

  Logger.log("Fungsi yang WAJIB punya invalidateDashboardCache_:");
  Object.keys(requiredFunctions).forEach(function (fn) {
    var source = "";
    try {
      source = eval(fn).toString();
    } catch (e) {
      Logger.log("  ⚠️ " + fn + " — tidak bisa di-eval");
      return;
    }
    var hasInvalidate = source.indexOf("invalidateDashboardCache_") !== -1;
    var hasMyInvalidate = source.indexOf("invalidateMyDashboardCache_") !== -1;
    Logger.log(
      "  " +
        (hasInvalidate ? "✅" : "❌") +
        " " +
        fn +
        (hasMyInvalidate ? " (+ my)" : ""),
    );
  });

  Logger.log("");
  Logger.log("Fungsi yang TIDAK perlu invalidateDashboardCache_:");
  Object.keys(excludedFunctions).forEach(function (fn) {
    var source = "";
    try {
      source = eval(fn).toString();
    } catch (e) {
      return;
    }
    var hasInvalidate = source.indexOf("invalidateDashboardCache_") !== -1;
    if (hasInvalidate) {
      Logger.log("  ⚠️ " + fn + " — punya invalidate (tidak perlu, hapus)");
    } else {
      Logger.log("  ✅ " + fn + " — tidak punya (benar)");
    }
  });
}

function testPhase5MyDashboard() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  // Clear cache
  invalidateMyDashboardCache_(ctx.user.member_id);

  var r = getMyDashboard_(ctx);
  Logger.log("My Dashboard:");
  Logger.log("  Attendance: " + r.data.attendance.length + " (max 50)");
  Logger.log("  Monitoring: " + r.data.monitoring.length + " (max 20)");
  Logger.log("  Upcoming: " + r.data.upcoming.length + " (max 5)");

  var json = JSON.stringify(r);
  Logger.log(
    "  Payload: " +
      json.length +
      " bytes (" +
      Math.round(json.length / 1024) +
      " KB)",
  );
}

function testPhase5Invalidation() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  // Warm cache
  getGeneralDashboard_(ctx);

  var cache = CacheService.getScriptCache();
  var key = "dash:general:" + ctx.user.role;
  Logger.log("Before invalidate: " + (cache.get(key) ? "ADA" : "KOSONG"));

  invalidateDashboardCache_();

  Logger.log("After invalidate: " + (cache.get(key) ? "ADA" : "KOSONG"));
}

function testPhase5Debug() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var cacheKey = "dash:general:" + ctx.user.role;
  Logger.log("Cache key: " + cacheKey);

  // Clear dulu
  CacheService.getScriptCache().remove(cacheKey);
  Logger.log(
    "After clear: " +
      (CacheService.getScriptCache().get(cacheKey) ? "ADA" : "KOSONG"),
  );

  // Test put manual
  CacheService.getScriptCache().put(
    cacheKey,
    JSON.stringify({ test: true }),
    60,
  );
  Logger.log(
    "After manual put: " +
      (CacheService.getScriptCache().get(cacheKey) ? "ADA" : "KOSONG"),
  );

  // Hapus, lalu panggil getGeneralDashboard_
  CacheService.getScriptCache().remove(cacheKey);
  Logger.log("");
  Logger.log("Calling getGeneralDashboard_...");

  var result = getGeneralDashboard_(ctx);
  Logger.log("Result success: " + result.success);

  // Cek cache
  var cached = CacheService.getScriptCache().get(cacheKey);
  Logger.log(
    "After getGeneralDashboard_: " +
      (cached ? "ADA (" + cached.length + " bytes)" : "KOSONG"),
  );

  // Cek isi result — berapa ukuran JSON-nya?
  var str = JSON.stringify(result);
  Logger.log("Result JSON size: " + str.length + " bytes");
  Logger.log(
    "Result data size: " + JSON.stringify(result.data).length + " bytes",
  );

  // Panggil lagi, ukur
  var t = Date.now();
  getGeneralDashboard_(ctx);
  Logger.log("Second call: " + (Date.now() - t) + "ms");
}

function testPhase5Debug2() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var cacheKey = "dash:general:" + ctx.user.role;

  // Clear
  CacheService.getScriptCache().remove(cacheKey);

  // Build manual
  var result = _buildGeneralDashboard_(ctx);
  Logger.log("Build result size: " + JSON.stringify(result).length);

  // Put manual TANPA try/catch
  var str = JSON.stringify(result);
  Logger.log("Str length: " + str.length);
  Logger.log("Will put...");

  CacheService.getScriptCache().put(cacheKey, str, 60);

  Logger.log("Put done");

  var cached = CacheService.getScriptCache().get(cacheKey);
  Logger.log(
    "After put: " + (cached ? "ADA (" + cached.length + " bytes)" : "KOSONG"),
  );
}

function testPhase5GeneralDashboard() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  invalidateDashboardCache_();

  var t1 = Date.now();
  var r1 = getGeneralDashboard_(ctx);
  var coldTime = Date.now() - t1;

  var t2 = Date.now();
  var r2 = getGeneralDashboard_(ctx);
  var warmTime = Date.now() - t2;

  Logger.log("Cold: " + coldTime + "ms");
  Logger.log("Warm: " + warmTime + "ms");
  Logger.log("Speedup: " + Math.round(coldTime / warmTime) + "x");
}

function testPhase5InvalidationFlow() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  // Warm cache
  getGeneralDashboard_(ctx);
  var cacheKey = "dash:general:" + ctx.user.role;
  Logger.log(
    "Setelah getGeneralDashboard_: " +
      (CacheService.getScriptCache().get(cacheKey) ? "ADA" : "KOSONG"),
  );

  // Trigger write — update member (dummy)
  var members = getMembers_(ctx, { limit: 1 });
  if (!members.data.length) {
    Logger.log("Tidak ada member");
    return;
  }
  var memberId = members.data[0].member_id;

  var updateResult = updateMember_(ctx, {
    member_id: memberId,
    nama_panggilan: "Test Invalidation " + Date.now(),
  });
  Logger.log("updateMember_: " + (updateResult.success ? "OK" : "FAIL"));

  // Cek cache setelah write
  Logger.log(
    "Setelah updateMember_: " +
      (CacheService.getScriptCache().get(cacheKey)
        ? "ADA (❌ harusnya KOSONG)"
        : "KOSONG (✅ invalidated)"),
  );
}

function testPhase4BackwardCompat() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var members = getMembers_(ctx, { limit: 1 });
  if (!members.data.length) {
    Logger.log("No members");
    return;
  }
  var memberId = members.data[0].member_id;

  // Test monitoring (default = array, limit 20)
  var mon = getMonitoring_(ctx, { member_id: memberId });
  Logger.log("Monitoring default:");
  Logger.log("  Is array: " + Array.isArray(mon.data));
  Logger.log(
    "  Length: " + (Array.isArray(mon.data) ? mon.data.length : "N/A"),
  );

  // Test monitoring paged
  var monPaged = getMonitoring_(ctx, { member_id: memberId, paged: true });
  Logger.log("Monitoring paged:");
  Logger.log("  Has items: " + !!monPaged.data.items);
  Logger.log("  Total: " + monPaged.data.total);
  Logger.log("  Items: " + monPaged.data.items.length);
  Logger.log("  Has more: " + monPaged.data.has_more);

  // Test announcements
  var ann = getAnnouncements_(ctx, {});
  Logger.log("");
  Logger.log("Announcements default:");
  Logger.log("  Is array: " + Array.isArray(ann.data));
  Logger.log(
    "  Length: " + (Array.isArray(ann.data) ? ann.data.length : "N/A"),
  );

  var annPaged = getAnnouncements_(ctx, { paged: true });
  Logger.log("Announcements paged:");
  Logger.log("  Total: " + annPaged.data.total);
  Logger.log("  Items: " + annPaged.data.items.length);

  // Test pending
  var pending = getPendingMembers_(ctx, {});
  Logger.log("");
  Logger.log("Pending default:");
  Logger.log("  Is array: " + Array.isArray(pending.data));
  Logger.log(
    "  Length: " + (Array.isArray(pending.data) ? pending.data.length : "N/A"),
  );
}

function testPhase4Payload() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var members = getMembers_(ctx, { limit: 1 });
  var memberId = members.data[0].member_id;

  // Payload monitoring default
  var mon = getMonitoring_(ctx, { member_id: memberId });
  var monSize = JSON.stringify(mon).length;
  Logger.log(
    "Monitoring payload (limit 20): " +
      monSize +
      " bytes (" +
      Math.round(monSize / 1024) +
      " KB)",
  );

  // Payload announcements default
  var ann = getAnnouncements_(ctx, {});
  var annSize = JSON.stringify(ann).length;
  Logger.log(
    "Announcements payload (limit 50): " +
      annSize +
      " bytes (" +
      Math.round(annSize / 1024) +
      " KB)",
  );

  // Payload pending
  var pending = getPendingMembers_(ctx, {});
  var pendSize = JSON.stringify(pending).length;
  Logger.log(
    "Pending payload (limit 50): " +
      pendSize +
      " bytes (" +
      Math.round(pendSize / 1024) +
      " KB)",
  );
}

function testPhase4PayloadReal() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var membersRepo = new SheetRepository_("members");
  var monitoringRepo = new SheetRepository_("monitoring");
  var allMonitoring = monitoringRepo.getAll();

  // Cari member yang punya banyak monitoring
  var byMember = {};
  allMonitoring.forEach(function (m) {
    byMember[m.member_id] = (byMember[m.member_id] || 0) + 1;
  });

  var topMember = null;
  var topCount = 0;
  Object.keys(byMember).forEach(function (mid) {
    if (byMember[mid] > topCount) {
      topCount = byMember[mid];
      topMember = mid;
    }
  });

  Logger.log(
    "Member dengan monitoring terbanyak: " +
      topMember +
      " (" +
      topCount +
      " record)",
  );

  // Test 1: Default (limit 20)
  var t1 = Date.now();
  var monDefault = getMonitoring_(ctx, { member_id: topMember });
  var t1Time = Date.now() - t1;
  var monDefaultSize = JSON.stringify(monDefault).length;

  Logger.log("");
  Logger.log("Default (limit 20):");
  Logger.log("  Time: " + t1Time + "ms");
  Logger.log("  Items: " + monDefault.data.length);
  Logger.log(
    "  Payload: " +
      monDefaultSize +
      " bytes (" +
      Math.round(monDefaultSize / 1024) +
      " KB)",
  );

  // Test 2: Paged (limit 20, paged true)
  var t2 = Date.now();
  var monPaged = getMonitoring_(ctx, { member_id: topMember, paged: true });
  var t2Time = Date.now() - t2;
  var monPagedSize = JSON.stringify(monPaged).length;

  Logger.log("");
  Logger.log("Paged (limit 20):");
  Logger.log("  Time: " + t2Time + "ms");
  Logger.log("  Items: " + monPaged.data.items.length);
  Logger.log("  Total: " + monPaged.data.total);
  Logger.log("  Has more: " + monPaged.data.has_more);
  Logger.log(
    "  Payload: " +
      monPagedSize +
      " bytes (" +
      Math.round(monPagedSize / 1024) +
      " KB)",
  );

  // Test 3: Paged limit 50
  var t3 = Date.now();
  var mon50 = getMonitoring_(ctx, {
    member_id: topMember,
    paged: true,
    limit: 50,
  });
  var t3Time = Date.now() - t3;
  var mon50Size = JSON.stringify(mon50).length;

  Logger.log("");
  Logger.log("Paged (limit 50):");
  Logger.log("  Time: " + t3Time + "ms");
  Logger.log("  Items: " + mon50.data.items.length);
  Logger.log(
    "  Payload: " +
      mon50Size +
      " bytes (" +
      Math.round(mon50Size / 1024) +
      " KB)",
  );

  // Simulasi: kalau tanpa limit (seperti sebelum FASE 4)
  var allRows = monitoringRepo.findByField("member_id", topMember);
  var allSize = JSON.stringify(allRows).length;
  Logger.log("");
  Logger.log(
    "SEBELUM FASE 4 (semua): " +
      allRows.length +
      " record, " +
      allSize +
      " bytes (" +
      Math.round(allSize / 1024) +
      " KB)",
  );

  Logger.log("");
  Logger.log("IMPROVEMENT:");
  Logger.log(
    "  Payload: " +
      Math.round((1 - monDefaultSize / allSize) * 100) +
      "% lebih kecil",
  );
  Logger.log("  Record: " + allRows.length + " → " + monDefault.data.length);
}

function generateDummyMonitoring() {
  var repo = new SheetRepository_("monitoring");
  var membersRepo = new SheetRepository_("members");
  var members = membersRepo.getAll().slice(0, 50); // 50 member pertama
  var now = nowIso_();

  var statuses = ["AKTIF", "PERLU_PERHATIAN", "KURANG_AKTIF", "TIDAK_AKTIF"];
  var jenisList = ["UMUM", "IBADAH", "SOSIAL", "AKADEMIK"];
  var total = 0;

  members.forEach(function (m) {
    var records = [];
    var count = 10 + Math.floor(Math.random() * 20); // 10-30 per member

    for (var i = 0; i < count; i++) {
      var daysAgo = Math.floor(Math.random() * 365);
      var tanggal = new Date(Date.now() - daysAgo * 86400000);
      records.push({
        monitoring_id: "MON" + Utilities.getUuid().slice(0, 8).toUpperCase(),
        member_id: m.member_id,
        tanggal: formatDate(tanggal),
        jenis: jenisList[Math.floor(Math.random() * jenisList.length)],
        status: statuses[Math.floor(Math.random() * statuses.length)],
        catatan: "Catatan dummy untuk testing " + i,
        tindak_lanjut: i % 3 === 0 ? "Tindak lanjut dummy" : "",
        created_by: "USR001",
        created_at: now,
        updated_at: now,
      });
    }

    repo.insertMany(records);
    total += records.length;
  });

  Logger.log(
    "Generated: " +
      total +
      " monitoring records for " +
      members.length +
      " members",
  );
}

function generateHeavyMonitoring() {
  var membersRepo = new SheetRepository_("members");
  var monitoringRepo = new SheetRepository_("monitoring");
  var members = membersRepo.getAll();

  if (members.length === 0) {
    Logger.log("Tidak ada member");
    return;
  }

  var targetMember = members[0];
  var now = nowIso_();
  var records = [];

  for (var i = 0; i < 200; i++) {
    var daysAgo = Math.floor(Math.random() * 365);
    var tanggal = new Date(Date.now() - daysAgo * 86400000);
    records.push({
      monitoring_id: "MON" + Utilities.getUuid().slice(0, 8).toUpperCase(),
      member_id: targetMember.member_id,
      tanggal: formatDate(tanggal),
      jenis: "UMUM",
      status: "AKTIF",
      catatan: "Catatan dummy berat untuk testing " + i,
      tindak_lanjut: "",
      created_by: "USR001",
      created_at: now,
      updated_at: now,
    });
  }

  monitoringRepo.insertMany(records);
  Logger.log(
    "Generated 200 monitoring untuk " +
      targetMember.member_id +
      " (" +
      targetMember.nama_lengkap +
      ")",
  );
}

function cleanupDummyMonitoring() {
  var repo = new SheetRepository_("monitoring");
  var all = repo.getAll();
  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var catatanColIndex = headers.indexOf("catatan");
  var lastRow = sheet.getLastRow();

  if (lastRow < 2) {
    Logger.log("Sheet kosong");
    return;
  }

  var values = sheet
    .getRange(2, catatanColIndex + 1, lastRow - 1, 1)
    .getValues();
  var rowsToDelete = [];

  for (var i = 0; i < values.length; i++) {
    var catatan = String(values[i][0] || "");
    if (catatan.indexOf("Catatan dummy") === 0) {
      rowsToDelete.push(i + 2);
    }
  }

  if (rowsToDelete.length === 0) {
    Logger.log("Tidak ada dummy monitoring");
    return;
  }

  // ⚠️ Sisakan minimal 1 baris data non-dummy
  var totalRows = lastRow - 1; // exclude header
  var maxDeletable = totalRows - 1; // sisakan 1 baris
  if (rowsToDelete.length > maxDeletable) {
    Logger.log(
      "⚠️ Terlalu banyak dummy (" +
        rowsToDelete.length +
        "). Sisakan " +
        maxDeletable +
        " baris untuk dihapus.",
    );
    rowsToDelete = rowsToDelete.slice(0, maxDeletable);
  }

  if (rowsToDelete.length === 0) {
    Logger.log(
      "Sheet hanya berisi dummy. Biarkan 1 baris dummy sebagai placeholder.",
    );
    Logger.log(
      "Atau hapus manual di Google Sheets, lalu jalankan setupSpreadsheet() untuk reset.",
    );
    return;
  }

  // Group consecutive rows
  rowsToDelete.sort(function (a, b) {
    return a - b;
  });
  var groups = [];
  var start = rowsToDelete[0];
  var prev = rowsToDelete[0];
  for (var j = 1; j < rowsToDelete.length; j++) {
    var curr = rowsToDelete[j];
    if (curr === prev + 1) {
      prev = curr;
    } else {
      groups.push({ start: start, count: prev - start + 1 });
      start = curr;
      prev = curr;
    }
  }
  groups.push({ start: start, count: prev - start + 1 });

  // Delete descending
  groups.sort(function (a, b) {
    return b.start - a.start;
  });
  groups.forEach(function (g) {
    sheet.deleteRows(g.start, g.count);
  });

  repo._invalidateCache();
  Logger.log("Deleted: " + rowsToDelete.length + " dummy monitoring records");
}

function deleteLastDummyMonitoring() {
  var repo = new SheetRepository_("monitoring");
  var all = repo.getAll();

  var dummy = all.filter(function (m) {
    return String(m.catatan || "").indexOf("Catatan dummy") === 0;
  });

  if (dummy.length === 0) {
    Logger.log("Tidak ada dummy tersisa");
    return;
  }

  // Hapus via clearContent (bukan deleteRows)
  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var idColIndex = headers.indexOf("monitoring_id");
  var lastRow = sheet.getLastRow();
  var idValues = sheet.getRange(2, idColIndex + 1, lastRow - 1, 1).getValues();

  var dummyIds = {};
  dummy.forEach(function (d) {
    dummyIds[d.monitoring_id] = true;
  });

  var cleared = 0;
  for (var i = 0; i < idValues.length; i++) {
    if (dummyIds[idValues[i][0]]) {
      sheet.getRange(i + 2, 1, 1, headers.length).clearContent();
      cleared++;
    }
  }

  repo._invalidateCache();
  Logger.log(
    "Cleared: " +
      cleared +
      " dummy rows (pakai clearContent, bukan deleteRows)",
  );
}

function testPhase3AttendancePage() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var meetings = getMeetings_(ctx, {});
  if (!meetings.data.length) {
    Logger.log("No meetings");
    return;
  }

  var meetingId = meetings.data[0].meeting_id;

  var t1 = Date.now();
  getMeetings_(ctx, {});
  getAttendanceMembers_(ctx, {});
  getAttendance_(ctx, { meeting_id: meetingId });
  var oldTime = Date.now() - t1;

  var t2 = Date.now();
  var pageResult = getAttendancePage_(ctx, { meeting_id: meetingId });
  var newTime = Date.now() - t2;

  Logger.log("Old flow (3 requests): " + oldTime + "ms");
  Logger.log("New flow (1 request): " + newTime + "ms");
  Logger.log("Improvement: " + Math.round((1 - newTime / oldTime) * 100) + "%");

  if (pageResult.success) {
    Logger.log("Meeting: " + pageResult.data.meeting.acara);
    Logger.log("Members: " + pageResult.data.members.length);
    Logger.log("Attendance rows: " + pageResult.data.attendance.length);

    var json = JSON.stringify(pageResult);
    Logger.log(
      "Payload: " +
        json.length +
        " bytes (" +
        Math.round(json.length / 1024) +
        " KB)",
    );
  } else {
    Logger.log("Error: " + pageResult.message);
  }
}

function testPhase3Fair() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var meetings = getMeetings_(ctx, {});
  if (!meetings.data.length) return;
  var meetingId = meetings.data[0].meeting_id;

  // Clear semua cache dulu
  var cache = CacheService.getScriptCache();
  cache.remove("sheet_meetings");
  cache.remove("sheet_members");
  cache.remove(ATTENDANCE_CACHE_PREFIX + meetingId);

  // Old flow (cold)
  var t1 = Date.now();
  getMeetings_(ctx, {});
  var tMeetings = Date.now() - t1;

  var t2 = Date.now();
  getAttendanceMembers_(ctx, {});
  var tMembers = Date.now() - t2;

  var t3 = Date.now();
  getAttendance_(ctx, { meeting_id: meetingId });
  var tAttendance = Date.now() - t3;

  var totalOld = tMeetings + tMembers + tAttendance;
  Logger.log("Old flow (cold):");
  Logger.log("  getMeetings: " + tMeetings + "ms");
  Logger.log("  getAttendanceMembers: " + tMembers + "ms");
  Logger.log("  getAttendance: " + tAttendance + "ms");
  Logger.log("  Total: " + totalOld + "ms");

  // Clear lagi
  cache.remove("sheet_meetings");
  cache.remove("sheet_members");
  cache.remove(ATTENDANCE_CACHE_PREFIX + meetingId);

  // New flow (cold)
  var t4 = Date.now();
  var pageResult = getAttendancePage_(ctx, { meeting_id: meetingId });
  var newTime = Date.now() - t4;

  Logger.log("");
  Logger.log("New flow (cold): " + newTime + "ms");
  Logger.log(
    "Improvement: " + Math.round((1 - newTime / totalOld) * 100) + "%",
  );
}

function testPhase2AttendanceMembersSize() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var ctx = validateSession_(loginResult.data.token);

  var t1 = Date.now();
  var result = getAttendanceMembers_(ctx, {});
  var elapsed = Date.now() - t1;

  Logger.log("Total members: " + result.data.length);
  Logger.log("Time: " + elapsed + "ms");

  var json = JSON.stringify(result);
  Logger.log(
    "Payload size: " +
      json.length +
      " bytes (" +
      Math.round(json.length / 1024) +
      " KB)",
  );

  Logger.log("Sample fields: " + Object.keys(result.data[0] || {}).join(", "));
}

function testPhase2MembersSize() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var ctx = validateSession_(loginResult.data.token);

  var t1 = Date.now();
  var result = getMembers_(ctx, { limit: 50 });
  var elapsed = Date.now() - t1;

  Logger.log("Total returned: " + result.data.length);
  Logger.log("Time: " + elapsed + "ms");

  var json = JSON.stringify(result);
  Logger.log(
    "Payload size: " +
      json.length +
      " bytes (" +
      Math.round(json.length / 1024) +
      " KB)",
  );

  Logger.log("Sample fields: " + Object.keys(result.data[0] || {}).join(", "));
}

function testPhase2MembersPaged() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var ctx = validateSession_(loginResult.data.token);

  var t1 = Date.now();
  var result = getMembersPaged_(ctx, {
    limit: 30,
    offset: 0,
    kategori: "REMAJA",
  });
  var elapsed = Date.now() - t1;

  Logger.log("Time: " + elapsed + "ms");
  Logger.log("Total: " + result.data.total);
  Logger.log("Items: " + result.data.items.length);
  Logger.log("Has more: " + result.data.has_more);
  Logger.log(
    "Sample names: " +
      result.data.items
        .slice(0, 3)
        .map(function (m) {
          return m.nama_lengkap;
        })
        .join(", "),
  );
}

function testPhase2MemberDetail() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var ctx = validateSession_(loginResult.data.token);

  var members = getMembers_(ctx, { limit: 1 });
  if (!members.data.length) {
    Logger.log("No members");
    return;
  }

  var result = getMemberDetail_(ctx, { member_id: members.data[0].member_id });
  Logger.log("Fields count: " + Object.keys(result.data).length);
  Logger.log("Fields: " + Object.keys(result.data).join(", "));
}

function testPhase1Login() {
  var result = login_({ username: "albetsurya", password: "albetsurya123" });
  Logger.log("Login: " + JSON.stringify(result.success));
  if (!result.success) return;

  var token = result.data.token;
  var cache = CacheService.getScriptCache();
  var cached = cache.get(SESSION_CACHE_PREFIX + token);
  Logger.log("Cached session: " + (cached ? "✅ OK" : "❌ MISS"));
  Logger.log("Payload: " + cached);

  var userIdx = cache.get(USER_SESSION_PREFIX + result.data.user.user_id);
  Logger.log("User index: " + userIdx);
}

function testPhase1Validate() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var token = loginResult.data.token;

  CacheService.getScriptCache().remove(SESSION_CACHE_PREFIX + token);

  var t1 = Date.now();
  var ctx1 = validateSession_(token);
  Logger.log("First call (cache miss): " + (Date.now() - t1) + "ms");

  var t2 = Date.now();
  var ctx2 = validateSession_(token);
  Logger.log("Second call (cache hit): " + (Date.now() - t2) + "ms");

  Logger.log("User: " + (ctx2 ? ctx2.user.nama : "null"));
}

function testPhase1Logout() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  var token = loginResult.data.token;
  var ctx = { token: token, user: loginResult.data.user };

  logout_(ctx);

  var cache = CacheService.getScriptCache();
  var cached = cache.get(SESSION_CACHE_PREFIX + token);
  Logger.log("After logout: " + (cached ? "❌ MASIH ADA" : "✅ Terhapus"));
}

function testPhase1InvalidateRole() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  if (!loginResult.success) {
    Logger.log("❌ Login gagal: " + loginResult.message);
    return;
  }

  var token = loginResult.data.token;
  var userId = loginResult.data.user.user_id;

  var adminCtx = {
    user: { user_id: userId, role: ROLES.SUPER_ADMIN, nama: "Albet" },
    token: token,
  };

  var membersRepo = new SheetRepository_("members");
  var usersRepo = new SheetRepository_("users");
  var allMembers = membersRepo.getAll();

  var eligibleMember = null;
  for (var i = 0; i < allMembers.length; i++) {
    var m = allMembers[i];
    var existing = usersRepo.find(function (u) {
      return u.member_id === m.member_id && toBool_(u.status_aktif);
    });
    if (existing.length === 0) {
      eligibleMember = m;
      break;
    }
  }

  if (!eligibleMember) {
    Logger.log("❌ Tidak ada member tanpa akun. Buat member dulu.");
    return;
  }

  Logger.log(
    "Pakai member: " +
      eligibleMember.member_id +
      " - " +
      eligibleMember.nama_lengkap,
  );

  var username = "testrole_" + Date.now();
  var createResult = createUser_(adminCtx, {
    username: username,
    password: "test123456",
    role: ROLES.ADMIN,
    nama: "Test Role",
    member_id: eligibleMember.member_id,
  });

  if (!createResult.success) {
    Logger.log("❌ Create gagal: " + createResult.message);
    return;
  }

  Logger.log("✅ User dibuat: " + createResult.data.user_id);

  var newUserId = createResult.data.user_id;

  var newLogin = login_({ username: username, password: "test123456" });
  if (!newLogin.success) {
    Logger.log("❌ Login user baru gagal");
    return;
  }

  var newToken = newLogin.data.token;
  var cache = CacheService.getScriptCache();

  Logger.log(
    "Before update: " +
      (cache.get(SESSION_CACHE_PREFIX + newToken) ? "✅ ADA" : "❌ KOSONG"),
  );

  var updateResult = updateUserRole_(adminCtx, {
    user_id: newUserId,
    role: ROLES.TIM_ABSENSI,
  });

  Logger.log(
    "Update result: " +
      (updateResult.success ? "✅ OK" : "❌ " + updateResult.message),
  );

  Logger.log(
    "After update: " +
      (cache.get(SESSION_CACHE_PREFIX + newToken)
        ? "❌ MASIH ADA"
        : "✅ TERHAPUS"),
  );
}

function testPhase1InvalidateStatus() {
  var loginResult = login_({
    username: "albetsurya",
    password: "albetsurya123",
  });
  if (!loginResult.success) {
    Logger.log("❌ Login gagal: " + loginResult.message);
    return;
  }

  var token = loginResult.data.token;
  var userId = loginResult.data.user.user_id;

  var adminCtx = {
    user: { user_id: userId, role: ROLES.SUPER_ADMIN, nama: "Albet" },
    token: token,
  };

  var membersRepo = new SheetRepository_("members");
  var usersRepo = new SheetRepository_("users");
  var allMembers = membersRepo.getAll();

  var eligibleMember = null;
  for (var i = 0; i < allMembers.length; i++) {
    var m = allMembers[i];
    var existing = usersRepo.find(function (u) {
      return u.member_id === m.member_id && toBool_(u.status_aktif);
    });
    if (existing.length === 0) {
      eligibleMember = m;
      break;
    }
  }

  if (!eligibleMember) {
    Logger.log("❌ Tidak ada member tanpa akun.");
    return;
  }

  Logger.log(
    "Pakai member: " +
      eligibleMember.member_id +
      " - " +
      eligibleMember.nama_lengkap,
  );

  var username = "teststatus_" + Date.now();
  var createResult = createUser_(adminCtx, {
    username: username,
    password: "test123456",
    role: ROLES.ADMIN,
    nama: "Test Status",
    member_id: eligibleMember.member_id,
  });

  if (!createResult.success) {
    Logger.log("❌ Create gagal: " + createResult.message);
    return;
  }

  Logger.log("✅ User dibuat: " + createResult.data.user_id);

  var newUserId = createResult.data.user_id;

  var newLogin = login_({ username: username, password: "test123456" });
  if (!newLogin.success) {
    Logger.log("❌ Login user baru gagal");
    return;
  }

  var newToken = newLogin.data.token;
  var cache = CacheService.getScriptCache();

  Logger.log(
    "Before: " +
      (cache.get(SESSION_CACHE_PREFIX + newToken) ? "✅ ADA" : "❌ KOSONG"),
  );

  var updateResult = updateUser_(adminCtx, {
    user_id: newUserId,
    status_aktif: false,
  });

  Logger.log(
    "Update result: " +
      (updateResult.success ? "✅ OK" : "❌ " + updateResult.message),
  );

  Logger.log(
    "After: " +
      (cache.get(SESSION_CACHE_PREFIX + newToken)
        ? "❌ MASIH ADA"
        : "✅ TERHAPUS"),
  );

  var ctxCheck = validateSession_(newToken);
  Logger.log(
    "validateSession returns: " + (ctxCheck ? "❌ masih valid" : "✅ null"),
  );
}

function testPhase1ChangePassword() {
  var loginA = login_({ username: "albetsurya", password: "albetsurya123" });
  var loginB = login_({ username: "albetsurya", password: "albetsurya123" });

  if (!loginA.success || !loginB.success) {
    Logger.log("❌ Login gagal: A=" + loginA.success + " B=" + loginB.success);
    return;
  }

  var tokenA = loginA.data.token;
  var tokenB = loginB.data.token;
  var userId = loginA.data.user.user_id;

  var cache = CacheService.getScriptCache();
  Logger.log(
    "Token A cached: " +
      (cache.get(SESSION_CACHE_PREFIX + tokenA) ? "✅" : "❌"),
  );
  Logger.log(
    "Token B cached: " +
      (cache.get(SESSION_CACHE_PREFIX + tokenB) ? "✅" : "❌"),
  );

  var ctxA = { token: tokenA, user: { user_id: userId } };

  var result = changeMyPassword_(ctxA, {
    old_password: "albetsurya123",
    new_password: "albetsurya456",
  });

  Logger.log(
    "changeMyPassword result: " +
      (result.success ? "✅ OK" : "❌ " + result.message),
  );

  Logger.log("After change (from A):");
  Logger.log(
    "  Token A: " +
      (cache.get(SESSION_CACHE_PREFIX + tokenA)
        ? "✅ MASIH ADA"
        : "❌ TERHAPUS"),
  );
  Logger.log(
    "  Token B: " +
      (cache.get(SESSION_CACHE_PREFIX + tokenB)
        ? "❌ MASIH ADA"
        : "✅ TERHAPUS"),
  );

  var resultBack = changeMyPassword_(ctxA, {
    old_password: "albetsurya456",
    new_password: "albetsurya123",
  });
  Logger.log(
    "Restore result: " +
      (resultBack.success ? "✅ OK" : "❌ " + resultBack.message),
  );
}

function cleanupPhase1TestUsers() {
  var usersRepo = new SheetRepository_("users");
  var membersRepo = new SheetRepository_("members");
  var allUsers = usersRepo.getAll();

  var toDelete = allUsers.filter(function (u) {
    return (
      String(u.username || "").indexOf("testrole_") === 0 ||
      String(u.username || "").indexOf("teststatus_") === 0
    );
  });

  Logger.log("User test yang akan dihapus: " + toDelete.length);

  toDelete.forEach(function (u) {
    usersRepo.deleteById("user_id", u.user_id);
    Logger.log("  🗑️  " + u.user_id + " | @" + u.username + " | " + u.nama);
  });

  Logger.log("✅ Cleanup selesai");
}

// Di Apps Script editor, jalankan manual:
function testAttendanceCache() {
  var t1 = Date.now();
  getAttendanceByMeeting_("MTG-XXX"); // ganti dengan meeting_id real
  Logger.log("First call: " + (Date.now() - t1) + "ms");

  var t2 = Date.now();
  getAttendanceByMeeting_("MTG-XXX");
  Logger.log("Cached call: " + (Date.now() - t2) + "ms"); // harus <100ms
}

function testInvalidate() {
  getAttendanceByMeeting_("MTG-XXX"); // cache
  invalidateAttendanceCache_("MTG-XXX");
  var cached = CacheService.getScriptCache().get("att_MTG-XXX");
  Logger.log("After invalidate: " + (cached === null ? "OK" : "FAIL"));
}

function testGetAllRecent() {
  // DEBUG: cek spreadsheet apa yang dipakai
  var ss = getSpreadsheet_();
  Logger.log("Spreadsheet name: " + ss.getName());
  Logger.log("Spreadsheet ID: " + ss.getId());
  Logger.log(
    "All sheets: " +
      ss
        .getSheets()
        .map(function (s) {
          return s.getName();
        })
        .join(", "),
  );

  var sheet = ss.getSheetByName("attendance");
  Logger.log("attendance sheet exists: " + (sheet !== null));
  if (sheet) {
    Logger.log("attendance lastRow: " + sheet.getLastRow());
    Logger.log("attendance lastCol: " + sheet.getLastColumn());
  }

  var repo = new SheetRepository_("attendance");

  var t1 = Date.now();
  var all = repo.getAll();
  Logger.log(
    "getAll(): " + all.length + " rows in " + (Date.now() - t1) + "ms",
  );

  var t2 = Date.now();
  var recent = repo.getAllRecent(3000);
  Logger.log(
    "getAllRecent(3000): " +
      recent.length +
      " rows in " +
      (Date.now() - t2) +
      "ms",
  );
}

function testBatchDelete() {
  // 1. Buat meeting dummy dengan banyak attendance
  // 2. Ukur waktu delete
  var t1 = Date.now();
  var result = deleteAttendanceByMeeting_(null, { meeting_id: "MTG-XXX" });
  Logger.log(
    "Deleted " + result.data.deleted + " rows in " + (Date.now() - t1) + "ms",
  );
}

function testBatchDeleteReal() {
  var meetingId = "MTGFD913B20";

  // 1. Cek dulu berapa row
  var before = new SheetRepository_("attendance").findByField(
    "meeting_id",
    meetingId,
  );
  Logger.log("Before: " + before.length + " rows");

  // 2. Ukur waktu delete — pakai ctx dummy
  var dummyCtx = { user: { user_id: "TEST", role: "SUPER_ADMIN" } };
  var t1 = Date.now();
  var result = deleteAttendanceByMeeting_(dummyCtx, { meeting_id: meetingId });
  var elapsed = Date.now() - t1;

  Logger.log("Deleted: " + result.data.deleted + " rows in " + elapsed + "ms");

  // 3. Cek sisa
  var after = new SheetRepository_("attendance").findByField(
    "meeting_id",
    meetingId,
  );
  Logger.log("After: " + after.length + " rows");
}

function generateDummyMembers() {
  var TOTAL = 1000;
  var membersRepo = new SheetRepository_("members");
  var now = nowIso_();

  var namaDepanL = [
    "Ahmad",
    "Budi",
    "Choirul",
    "Dedi",
    "Eko",
    "Fajar",
    "Gunawan",
    "Hadi",
    "Irfan",
    "Joko",
    "Krisna",
    "Lukman",
    "Muhammad",
    "Nur",
    "Omar",
    "Putra",
    "Rahmat",
    "Slamet",
    "Taufik",
    "Umar",
  ];
  var namaDepanP = [
    "Aisyah",
    "Bella",
    "Citra",
    "Dewi",
    "Eka",
    "Fitri",
    "Gita",
    "Hana",
    "Indah",
    "Jihan",
    "Kartika",
    "Lina",
    "Maya",
    "Nadia",
    "Oktaviani",
    "Putri",
    "Rina",
    "Sari",
    "Tina",
    "Umi",
  ];
  var namaBelakang = [
    "Wijaya",
    "Santoso",
    "Kusuma",
    "Pratama",
    "Hidayat",
    "Nugroho",
    "Setiawan",
    "Firmansyah",
    "Ramadhan",
    "Maulana",
    "Saputra",
    "Hartono",
    "Susanto",
    "Gunadi",
    "Wibowo",
    "Rahmawati",
    "Anggraini",
    "Puspita",
    "Lestari",
    "Safitri",
  ];

  var kelompokList = ["GRP_A", "GRP_B", "GRP_C", "GRP_D", "GRP_E"];
  var desaList = [
    "Desa Sukamaju",
    "Desa Mekarsari",
    "Desa Banyuwangi",
    "Desa Wonorejo",
    "Desa Sidoarjo",
  ];
  var daerahList = [
    "Kecamatan A",
    "Kecamatan B",
    "Kecamatan C",
    "Kecamatan D",
    "Kecamatan E",
  ];
  var pekerjaanList = [
    "Karyawan",
    "Wiraswasta",
    "Mahasiswa",
    "Pelajar",
    "Guru",
    "Petani",
    "Buruh",
    "PNS",
  ];
  var hobiList = [
    "Membaca",
    "Olahraga",
    "Traveling",
    "Memasak",
    "Berkebun",
    "Menulis",
    "Musik",
    "Fotografi",
  ];
  var jenjangList = ["SD", "SMP", "SMA", "SMK", "S1", "S2"];
  var sekolahList = [
    "SDN 1",
    "SMPN 2",
    "SMAN 3",
    "SMKN 4",
    "Universitas A",
    "Universitas B",
  ];
  var jurusanList = [
    "IPA",
    "IPS",
    "Teknik Informatika",
    "Ekonomi",
    "Hukum",
    "Kedokteran",
    "Psikologi",
  ];
  var statusPembinaanList = [
    "AKTIF",
    "AKTIF",
    "AKTIF",
    "AKTIF",
    "PERLU_PERHATIAN",
    "KURANG_AKTIF",
  ];

  var genderOptions = ["L", "P"];

  var batchSize = 100;
  var batches = Math.ceil(TOTAL / batchSize);
  var inserted = 0;

  for (var b = 0; b < batches; b++) {
    var batch = [];
    var batchCount = Math.min(batchSize, TOTAL - inserted);

    for (var i = 0; i < batchCount; i++) {
      var idx = inserted + i + 1;
      var gender = genderOptions[Math.floor(Math.random() * 2)];
      var namaDepan =
        gender === "L"
          ? namaDepanL[Math.floor(Math.random() * namaDepanL.length)]
          : namaDepanP[Math.floor(Math.random() * namaDepanP.length)];
      var namaBelakang =
        namaBelakang[Math.floor(Math.random() * namaBelakang.length)];
      var namaLengkap = "DUMMY_" + namaDepan + " " + namaBelakang + " " + idx;

      var usia = 5 + Math.floor(Math.random() * 60);
      var tahunLahir = new Date().getFullYear() - usia;
      var bulanLahir = 1 + Math.floor(Math.random() * 12);
      var tanggalLahir = 1 + Math.floor(Math.random() * 28);
      var tanggalLahirStr =
        tahunLahir +
        "-" +
        (bulanLahir < 10 ? "0" : "") +
        bulanLahir +
        "-" +
        (tanggalLahir < 10 ? "0" : "") +
        tanggalLahir;

      var jenjang = jenjangList[Math.floor(Math.random() * jenjangList.length)];

      batch.push({
        member_id: generateMemberId(),
        nama_lengkap: namaLengkap,
        nama_panggilan: namaDepan,
        jenis_kelamin: gender,
        tempat_lahir: "Kota " + String.fromCharCode(65 + (idx % 26)),
        tanggal_lahir: tanggalLahirStr,
        foto_url: "",
        no_wa: "62812" + String(100000000 + idx).slice(-8),
        alamat_rumah:
          "Jl. Dummy No. " +
          idx +
          ", RT " +
          (1 + (idx % 10)) +
          "/RW " +
          (1 + (idx % 5)),
        desa: desaList[idx % desaList.length],
        daerah: daerahList[idx % daerahList.length],
        kelompok: kelompokList[idx % kelompokList.length],
        is_muballigh: idx % 20 === 0,
        is_kerja: usia >= 17 && idx % 3 === 0,
        is_nikah: usia >= 22 && idx % 4 === 0,
        tinggi_badan:
          usia >= 10 ? String(150 + Math.floor(Math.random() * 30)) : "",
        berat_badan:
          usia >= 10 ? String(40 + Math.floor(Math.random() * 40)) : "",
        hobi: hobiList[idx % hobiList.length],
        pekerjaan:
          usia >= 17 ? pekerjaanList[idx % pekerjaanList.length] : "Pelajar",
        status_pembinaan: statusPembinaanList[idx % statusPembinaanList.length],
        status_aktif: true,
        tanggal_masuk: formatDate(now),
        tanggal_keluar: "",
        jenjang_pendidikan: jenjang,
        sekolah: sekolahList[idx % sekolahList.length],
        jurusan: usia >= 17 ? jurusanList[idx % jurusanList.length] : "",
        tahun_mulai_pendidikan: String(tahunLahir + 6),
        tahun_selesai_pendidikan: String(tahunLahir + 6 + 6),
        created_at: now,
        updated_at: now,
      });
    }

    membersRepo.insertMany(batch);
    inserted += batchCount;
    Logger.log(
      "Batch " +
        (b + 1) +
        "/" +
        batches +
        " selesai (" +
        inserted +
        "/" +
        TOTAL +
        ")",
    );
  }

  Logger.log("=== SELESAI ===");
  Logger.log("Total dummy member: " + inserted);

  return { total: inserted };
}

function deleteDummyMembers() {
  var membersRepo = new SheetRepository_("members");
  var sheet = membersRepo._sheet();
  var headers = membersRepo.def.headers;
  var lastRow = sheet.getLastRow();

  if (lastRow < 2) {
    Logger.log("Sheet kosong");
    return { deleted: 0 };
  }

  var nameColIndex = headers.indexOf("nama_lengkap");
  var values = sheet.getRange(2, nameColIndex + 1, lastRow - 1, 1).getValues();

  var rowsToDelete = [];
  for (var i = 0; i < values.length; i++) {
    var nama = String(values[i][0] || "");
    if (nama.indexOf("DUMMY_") === 0) {
      rowsToDelete.push(i + 2);
    }
  }

  if (rowsToDelete.length === 0) {
    Logger.log("Tidak ada dummy member");
    return { deleted: 0 };
  }

  rowsToDelete.sort(function (a, b) {
    return a - b;
  });

  var groups = [];
  var start = rowsToDelete[0];
  var prev = rowsToDelete[0];
  for (var j = 1; j < rowsToDelete.length; j++) {
    var curr = rowsToDelete[j];
    if (curr === prev + 1) {
      prev = curr;
    } else {
      groups.push({ start: start, count: prev - start + 1 });
      start = curr;
      prev = curr;
    }
  }
  groups.push({ start: start, count: prev - start + 1 });

  groups.sort(function (a, b) {
    return b.start - a.start;
  });
  groups.forEach(function (g) {
    sheet.deleteRows(g.start, g.count);
  });

  membersRepo._invalidateCache();

  Logger.log("=== SELESAI ===");
  Logger.log("Dummy member dihapus: " + rowsToDelete.length);

  return { deleted: rowsToDelete.length };
}

function countDummyMembers() {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();
  var dummy = all.filter(function (m) {
    return String(m.nama_lengkap || "").indexOf("DUMMY_") === 0;
  });

  Logger.log("Total member: " + all.length);
  Logger.log("Dummy member: " + dummy.length);
  Logger.log("Member asli: " + (all.length - dummy.length));

  return {
    total: all.length,
    dummy: dummy.length,
    real: all.length - dummy.length,
  };
}

function testGetMembers() {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();
  Logger.log("Total members: " + all.length);
  Logger.log(
    "Dummy members: " +
      all.filter(function (m) {
        return String(m.nama_lengkap || "").indexOf("DUMMY_") === 0;
      }).length,
  );

  var active = all.filter(function (m) {
    return toBool_(m.status_aktif);
  });
  Logger.log("Active members: " + active.length);
}

function debugQuotaIncrement() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var key = "aiquota:" + ctx.user.user_id + ":" + formatDate(nowIso_());
  PropertiesService.getScriptProperties().setProperty(key, "0");

  Logger.log("=== CEK SOURCE CODE YANG DI-LOAD ===");
  [
    "omnirouteChat_",
    "omnirouteChatMember_",
    "geminiChat_",
    "groqChatWithModel_",
    "groqChatMemberWithModel_",
    "handleAiChat_",
  ].forEach(function (fn) {
    try {
      var src = eval(fn).toString();
      var hasInc = src.indexOf("incrementAiQuota_") !== -1;
      var hasLog = src.indexOf("logAiUsage_") !== -1;
      Logger.log(
        "  " +
          fn +
          ": len=" +
          src.length +
          ", hasInc=" +
          hasInc +
          ", hasLog=" +
          hasLog,
      );
    } catch (e) {
      Logger.log("  " + fn + ": ERROR " + e.message);
    }
  });

  Logger.log("");
  Logger.log("=== CEK PROVIDER AKTIF ===");
  Logger.log("  Stored: " + getStoredProvider_());
  try {
    Logger.log("  Active: " + getActiveProvider());
  } catch (e) {
    Logger.log("  Active: ERROR " + e.message);
  }

  Logger.log("");
  Logger.log("=== PANGGIL omnirouteChat_ LANGSUNG ===");
  try {
    var r1 = omnirouteChat_({ message: "Halo", history: [] }, "Test", ctx);
    Logger.log("  Success: " + r1.success);
    Logger.log("  Kuota setelah omnirouteChat_: " + getAiQuotaUsed_(ctx));
  } catch (e) {
    Logger.log("  ERROR: " + e.message);
  }

  Logger.log("");
  Logger.log("=== PANGGIL geminiChat_ LANGSUNG ===");
  PropertiesService.getScriptProperties().setProperty(key, "0");
  try {
    var r2 = geminiChat_({ message: "Halo", history: [] }, "Test", ctx);
    Logger.log("  Success: " + r2.success);
    Logger.log("  Kuota setelah geminiChat_: " + getAiQuotaUsed_(ctx));
  } catch (e) {
    Logger.log("  ERROR: " + e.message);
  }

  Logger.log("");
  Logger.log("=== PANGGIL handleAiChat_ (END-TO-END) ===");
  PropertiesService.getScriptProperties().setProperty(key, "0");
  try {
    var r3 = handleAiChat_({ message: "Halo", history: [] }, ctx);
    Logger.log("  Success: " + r3.success);
    Logger.log("  Kuota setelah handleAiChat_: " + getAiQuotaUsed_(ctx));
  } catch (e) {
    Logger.log("  ERROR: " + e.message);
  }
}

function testQuotaExhaustion() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var key = "aiquota:" + ctx.user.user_id + ":" + formatDate(nowIso_());
  PropertiesService.getScriptProperties().setProperty(key, "99");

  Logger.log("Kuota dipakai: " + getAiQuotaUsed_(ctx) + "/100");

  // Chat ke-100 — harus LOLOS
  var r1 = handleAiChat_({ message: "Test 100", history: [] }, ctx);
  Logger.log(
    "Panggilan ke-100: " +
      (r1.success ? "✅ LOLOS" : "❌ DIBLOKIR: " + r1.message),
  );

  // Chat ke-101 — harus DIBLOKIR
  var r2 = handleAiChat_({ message: "Test 101", history: [] }, ctx);
  Logger.log(
    "Panggilan ke-101: " +
      (r2.success ? "❌ LOLOS (bug!)" : "✅ DIBLOKIR: " + r2.message),
  );

  // Reset
  PropertiesService.getScriptProperties().setProperty(key, "0");
  Logger.log("Counter direset ke 0");
}

function testSessionsSize() {
  Logger.log("=== UKURAN SHEET SESSIONS ===");
  var sheet = new SheetRepository_("sessions")._sheet();
  var lastRow = sheet.getLastRow();
  Logger.log("Total baris sessions: " + (lastRow - 1) + " (exclude header)");

  var repo = new SheetRepository_("sessions");
  var all = repo.getAll();
  var now = Date.now();
  var expired = all.filter(function (s) {
    return new Date(s.expires_at).getTime() < now;
  }).length;
  Logger.log("Expired: " + expired);
  Logger.log("Active: " + (all.length - expired));

  // Trigger cleanup manual — ukur waktunya
  var t1 = Date.now();
  maybeCleanupExpiredSessions_();
  var elapsed = Date.now() - t1;
  Logger.log("Cleanup time: " + elapsed + "ms");

  var afterRow = sheet.getLastRow();
  Logger.log("Setelah cleanup: " + (afterRow - 1) + " baris");
}
function testFlushAudit() {
  Logger.log("=== AUDIT FLUSH ===");
  var fns = ["submitPublicRegistration_"];
  fns.forEach(function (fn) {
    var src = eval(fn).toString();
    var count = (src.match(/SpreadsheetApp\.flush\(\)/g) || []).length;
    Logger.log(fn + ": " + count + " kali flush");
  });
}
function testDeleteAttendancePerf() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var attendanceRepo = new SheetRepository_("attendance");
  var all = attendanceRepo.getAll();
  Logger.log("Total attendance: " + all.length);

  // Cari meeting dengan paling banyak absensi
  var byMeeting = {};
  all.forEach(function (a) {
    byMeeting[a.meeting_id] = (byMeeting[a.meeting_id] || 0) + 1;
  });

  var topMeeting = null;
  var topCount = 0;
  Object.keys(byMeeting).forEach(function (k) {
    if (byMeeting[k] > topCount) {
      topCount = byMeeting[k];
      topMeeting = k;
    }
  });

  Logger.log(
    "Meeting dengan absensi terbanyak: " +
      topMeeting +
      " (" +
      topCount +
      " baris)",
  );
  Logger.log("");
  Logger.log(
    "Untuk mengukur waktu delete, jalankan testBatchDeleteReal dengan meeting_id = " +
      topMeeting,
  );
  Logger.log("(Perlu backup manual sebelum delete)");
}
function testDriveFolderSize() {
  Logger.log("=== UKURAN FOLDER DRIVE ===");
  try {
    var folder = getPhotoFolder_();
    Logger.log("Folder: " + folder.getName());

    var t1 = Date.now();
    var files = folder.getFiles();
    var count = 0;
    while (files.hasNext()) {
      files.next();
      count++;
    }
    Logger.log("Jumlah file: " + count);
    Logger.log("Waktu iterasi: " + (Date.now() - t1) + "ms");

    if (count > 200) {
      Logger.log(
        "⚠️ Folder besar — uploadPhoto_ yang scan folder memang lambat",
      );
    } else if (count > 50) {
      Logger.log("Folder sedang — dampak moderat");
    } else {
      Logger.log("Folder kecil — dampak minimal");
    }
  } catch (e) {
    Logger.log("Error: " + e.message);
  }
}

function testKategoriBaru() {
  Logger.log("=== TEST KATEGORI (PAUD/TK = CABERAWIT) ===");
  Logger.log("");

  function tglLahirDariUsia(usia) {
    var now = new Date();
    var tahunLahir = now.getFullYear() - usia;
    return tahunLahir + "-06-15";
  }

  var cases = [
    // ============ BALITA (usia < 6 th, jenjang KOSONG) ============
    ["Bayi 0 th (tanpa jenjang)", 0, false, "", "BALITA"],
    ["Bayi 2 th (tanpa jenjang)", 2, false, "", "BALITA"],
    ["Balita 4 th (tanpa jenjang)", 4, false, "", "BALITA"],
    ["Balita 5 th (tanpa jenjang)", 5, false, "", "BALITA"],

    // ============ CABERAWIT (PAUD/TK selalu CABERAWIT) ============
    ["PAUD 3 th", 3, false, "PAUD", "CABERAWIT"],
    ["PAUD 4 th", 4, false, "PAUD", "CABERAWIT"],
    ["TK 4 th", 4, false, "TK", "CABERAWIT"],
    ["TK 5 th", 5, false, "TK", "CABERAWIT"],
    ["TK 6 th", 6, false, "TK", "CABERAWIT"],
    ["PAUD 7 th", 7, false, "PAUD", "CABERAWIT"],

    // ============ CABERAWIT (SD atau usia 6–12) ============
    ["SD 7 th", 7, false, "SD", "CABERAWIT"],
    ["SD 8 th", 8, false, "SD", "CABERAWIT"],
    ["Caberawit 6 th (usia, tanpa jenjang)", 6, false, "", "CABERAWIT"],
    ["Caberawit 10 th (usia, tanpa jenjang)", 10, false, "", "CABERAWIT"],
    ["Caberawit 12 th (usia, tanpa jenjang)", 12, false, "", "CABERAWIT"],

    // ============ PRA_REMAJA ============
    ["SMP 14 th", 14, false, "SMP", "PRA_REMAJA"],
    ["Pra Remaja 13 th (usia, tanpa jenjang)", 13, false, "", "PRA_REMAJA"],
    ["Pra Remaja 15 th (usia, tanpa jenjang)", 15, false, "", "PRA_REMAJA"],

    // ============ REMAJA ============
    ["SMA 17 th", 17, false, "SMA", "REMAJA"],
    ["SMK 17 th", 17, false, "SMK", "REMAJA"],
    ["Remaja 16 th (usia, tanpa jenjang)", 16, false, "", "REMAJA"],
    ["Remaja 18 th (usia, tanpa jenjang)", 18, false, "", "REMAJA"],

    // ============ PRA_NIKAH ============
    ["Pra Nikah 19 th", 19, false, "", "PRA_NIKAH"],
    ["Pra Nikah 25 th", 25, false, "", "PRA_NIKAH"],

    // ============ DEWASA ============
    ["Dewasa 25 th (nikah)", 25, true, "", "DEWASA"],
    ["Dewasa 35 th (nikah)", 35, true, "", "DEWASA"],
    ["Dewasa 50 th (nikah)", 50, true, "", "DEWASA"],

    // ============ ISTIMEWA ============
    ["Istimewa 60 th (belum nikah)", 60, false, "", "ISTIMEWA"],
    ["Istimewa 65 th (nikah)", 65, true, "", "ISTIMEWA"],
    ["Istimewa 70 th (belum nikah)", 70, false, "", "ISTIMEWA"],
  ];

  var lulus = 0;
  var gagal = 0;

  cases.forEach(function (c) {
    var member = {
      nama_lengkap: c[0],
      tanggal_lahir: tglLahirDariUsia(c[1]),
      is_nikah: c[2],
      jenjang_pendidikan: c[3],
    };
    var hasil = getMemberCategory(member);
    var expected = c[4];
    var ok = hasil === expected;

    Logger.log(
      (ok ? "✅" : "❌") +
        " " +
        c[0] +
        " → " +
        hasil +
        (ok ? "" : " (expected: " + expected + ")"),
    );

    if (ok) lulus++;
    else gagal++;
  });

  Logger.log("");
  Logger.log("═══════════════════════════════════════");
  Logger.log("Lulus: " + lulus + " / " + cases.length);
  Logger.log("Gagal: " + gagal);
  Logger.log("═══════════════════════════════════════");

  if (gagal === 0) {
    Logger.log("🎉 SEMUA TEST LULUS");
  } else {
    Logger.log("⚠️ Ada " + gagal + " test gagal");
  }
}

function testBulkPreview() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var result = previewBulkMeetings_(ctx, {
    tahun: 2026,
    bulan: 10,
    hari: ["Minggu", "Selasa", "Kamis"],
    jam: "Isya di tempat",
    group_id: "GRPC88D1BA3",
    acara: "Sambung Kelompok",
    kategori_target: ["CABERAWIT", "BALITA"],
  });

  Logger.log("Success: " + result.success);
  Logger.log("Total new: " + result.data.total_new);
  Logger.log("Member target: " + result.data.member_target.length);
  Logger.log("Total WA: " + result.data.total_wa);
}

function testBulkPreviewSetelahFix() {
  Logger.log("=== TEST BULK PREVIEW SETELAH FIX ===");
  Logger.log("");

  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var result = previewBulkMeetings_(ctx, {
    tahun: 2026,
    bulan: 10,
    hari: ["Minggu", "Selasa", "Kamis"],
    jam: "Isya di tempat",
    group_id: "GRPC88D1BA3",
    acara: "Sambung Kelompok",
    kategori_target: ["CABERAWIT", "BALITA"],
  });

  Logger.log("Success: " + result.success);
  if (!result.success) {
    Logger.log("Error: " + result.message);
    return;
  }

  Logger.log("Total new: " + result.data.total_new);
  Logger.log("Member target: " + result.data.member_target.length);
  Logger.log("Total WA: " + result.data.total_wa);
  Logger.log("");

  if (result.data.member_target.length > 0) {
    Logger.log("Target member:");
    result.data.member_target.forEach(function (m) {
      Logger.log(
        "  - " + m.nama_lengkap + " | " + m.no_wa + " | " + m.kategori,
      );
    });
  } else {
    Logger.log("⚠️ Member target MASIH 0 setelah fix.");
    Logger.log("→ Masalahnya bukan di group filter, tapi di data member.");
  }
}

function verifyRollback() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var result = previewBulkMeetings_(ctx, {
    tahun: 2026,
    bulan: 12,
    hari: ["Minggu", "Kamis"],
    jam: "Isya di tempat",
    group_id: "GRPC88D1BA3",
    acara: "Sambung Kelompok",
  });

  Logger.log("Success: " + result.success);
  Logger.log("Total new: " + result.data.total_new);
  Logger.log("Total existing: " + result.data.total_existing);
  Logger.log(
    "Has member_target? " +
      (result.data.member_target !== undefined
        ? "❌ MASIH ADA"
        : "✅ sudah dihapus"),
  );
}
function testAnnouncementTemplateCRUD() {
  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  Logger.log("=== TEST TEMPLATE CRUD ===");
  Logger.log("");

  // 1. Create
  var created = createAnnouncementTemplate_(ctx, {
    nama_template: "Test Template",
    kode: "TEST_TPL",
    isi_template: "Assalamu'alaikum {{nama_kelompok}}\n\nAcara: {{acara}}",
  });
  Logger.log("Create: " + created.success);
  if (!created.success) {
    Logger.log("  Error: " + created.message);
    return;
  }
  var templateId = created.data.template_id;
  Logger.log("  ID: " + templateId);
  Logger.log("");

  // 2. Detail
  var detail = getAnnouncementTemplateDetail_(ctx, {
    template_id: templateId,
  });
  Logger.log("Detail: " + detail.success);
  Logger.log("  Nama: " + detail.data.nama_template);
  Logger.log("");

  // 3. Update
  var updated = updateAnnouncementTemplate_(ctx, {
    template_id: templateId,
    nama_template: "Test Template Updated",
  });
  Logger.log("Update: " + updated.success);
  Logger.log("");

  // 4. List semua
  var list = getAllAnnouncementTemplates_(ctx, { include_inactive: "true" });
  Logger.log("List total: " + list.data.length + " template");
  Logger.log("");

  // 5. Delete
  var deleted = deleteAnnouncementTemplate_(ctx, {
    template_id: templateId,
  });
  Logger.log("Delete: " + deleted.success);
  Logger.log("");

  Logger.log("=== SELESAI ===");
}
