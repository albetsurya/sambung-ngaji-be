function testBatch3() {
  Logger.log("=== TEST BATCH 3 ===");

  var loginResult = login_({ username: "superadmin", password: "ganti123" });
  if (!loginResult.success) {
    Logger.log("❌ Login gagal: " + loginResult.message);
    return;
  }
  var ctx = validateSession_(loginResult.data.token);

  Logger.log("1. Test getMembers_:");
  var membersResult = getMembers_(ctx, {});
  if (membersResult.success) {
    Logger.log("   ✅ OK — " + membersResult.data.length + " members");
  } else {
    Logger.log("   ❌ " + membersResult.message);
  }

  Logger.log("");
  Logger.log("2. Test getMembersPaged_:");
  var pagedResult = getMembersPaged_(ctx, { limit: 5, offset: 0 });
  if (pagedResult.success) {
    Logger.log(
      "   ✅ OK — " +
        pagedResult.data.items.length +
        " of " +
        pagedResult.data.total,
    );
  } else {
    Logger.log("   ❌ " + pagedResult.message);
  }

  Logger.log("");
  Logger.log("3. Test getAttendance_ (by meeting):");
  var meetingsRepo = new SheetRepository_("meetings");
  var meetings = meetingsRepo.getAll();
  if (meetings.length > 0) {
    var attResult = getAttendance_(ctx, { meeting_id: meetings[0].meeting_id });
    if (attResult.success) {
      Logger.log("   ✅ OK — " + attResult.data.length + " attendance");
    } else {
      Logger.log("   ❌ " + attResult.message);
    }
  } else {
    Logger.log("   ⚠️ Tidak ada meeting untuk test");
  }

  Logger.log("");
  Logger.log("4. Test getMonitoring_:");
  if (membersResult.success && membersResult.data.length > 0) {
    var monResult = getMonitoring_(ctx, {
      member_id: membersResult.data[0].member_id,
    });
    if (monResult.success) {
      Logger.log("   ✅ OK — " + monResult.data.length + " entries");
    } else {
      Logger.log("   ❌ " + monResult.message);
    }
  }

  Logger.log("");
  Logger.log("✅ Batch 3 selesai");
}

function testBatch4() {
  Logger.log("=== TEST BATCH 4 ===");

  var loginResult = login_({ username: "superadmin", password: "ganti123" });
  if (!loginResult.success) {
    Logger.log("❌ Login gagal");
    return;
  }
  var ctx = validateSession_(loginResult.data.token);

  Logger.log("1. Test createMeeting_:");
  var createResult = createMeeting_(ctx, {
    tanggal: "2026-09-20",
    jam: "Isya",
    acara: "Test Meeting",
    kategori_target: ["PRA_NIKAH"],
  });
  if (createResult.success) {
    Logger.log("   ✅ Meeting created: " + createResult.data.meeting_id);

    Logger.log("");
    Logger.log("2. Test getMeetings_:");
    var listResult = getMeetings_(ctx, {});
    if (listResult.success) {
      Logger.log("   ✅ OK — " + listResult.data.length + " meetings");
    }

    Logger.log("");
    Logger.log("3. Test updateMeeting_:");
    var updateResult = updateMeeting_(ctx, {
      meeting_id: createResult.data.meeting_id,
      acara: "Test Meeting (Updated)",
    });
    if (updateResult.success) {
      Logger.log("   ✅ Updated: " + updateResult.data.acara);
    }
  } else {
    Logger.log("   ❌ " + createResult.message);
  }

  Logger.log("");
  Logger.log("4. Test saveGroup_:");
  var groupResult = saveGroup_(ctx, {
    group_code: "TEST",
    group_name: "Test Group",
  });
  if (groupResult.success) {
    Logger.log("   ✅ Group created: " + groupResult.data.group_id);
  }

  Logger.log("");
  Logger.log("5. Test getGroups_:");
  var groupsList = getGroups_(ctx, {});
  if (groupsList.success) {
    Logger.log("   ✅ OK — " + groupsList.data.length + " groups");
  }

  Logger.log("");
  Logger.log("6. Test getAnnouncementTemplates_:");
  var tplResult = getAnnouncementTemplates_(ctx, {});
  if (tplResult.success) {
    Logger.log("   ✅ OK — " + tplResult.data.length + " templates");
  }

  Logger.log("");
  Logger.log("✅ Batch 4 selesai");
}

function testBatch5() {
  Logger.log("=== TEST BATCH 5 ===");

  var loginResult = login_({ username: "superadmin", password: "ganti123" });
  if (!loginResult.success) {
    Logger.log("❌ Login gagal: " + loginResult.message);
    return;
  }
  var ctx = validateSession_(loginResult.data.token);
  Logger.log("1. Login: ✅");

  Logger.log("");
  Logger.log("2. Test getUsers_:");
  var usersResult = getUsers_(ctx, {});
  if (usersResult.success) {
    Logger.log("   ✅ " + usersResult.data.length + " users");
  }

  Logger.log("");
  Logger.log("3. Test getSettings_:");
  var settingsResult = getSettings_(ctx, {});
  if (settingsResult.success) {
    Logger.log("   ✅ Settings OK");
    Logger.log(
      "   jadwal_rutin: " + JSON.stringify(settingsResult.data.jadwal_rutin),
    );
  }

  Logger.log("");
  Logger.log("4. Test getAuditLogs_:");
  var auditResult = getAuditLogs_(ctx, { limit: 5 });
  if (auditResult.success) {
    Logger.log("   ✅ " + auditResult.data.length + " logs");
  }

  Logger.log("");
  Logger.log("5. Test submitPublicRegistration_:");
  var regResult = submitPublicRegistration_(null, {
    nama_lengkap: "Test Registrasi " + Date.now(),
    jenis_kelamin: "L",
    no_wa: "62812" + String(Date.now()).slice(-8),
    _client_ip: "test-ip-" + Date.now(),
  });
  if (regResult.success) {
    Logger.log("   ✅ Submission: " + regResult.data.submission_id);
  } else {
    Logger.log("   ⚠️ " + regResult.message);
  }

  Logger.log("");
  Logger.log("✅ Batch 5 selesai");
}

function testAfterCleanup() {
  Logger.log("=== TEST SETELAH HAPUS Kode.js ===");

  Logger.log("1. Test nowIso_: " + nowIso_());
  Logger.log("2. Test formatDate: " + formatDate("2026-09-13"));

  Logger.log("");
  Logger.log("3. Test login:");
  var loginResult = login_({ username: "superadmin", password: "ganti123" });
  if (loginResult.success) {
    Logger.log("   ✅ Login OK");

    var ctx = validateSession_(loginResult.data.token);

    Logger.log("");
    Logger.log("4. Test getDashboard_:");
    var dashResult = getDashboard_(ctx, {});
    if (dashResult.success) {
      Logger.log("   ✅ Dashboard OK");
    }

    Logger.log("");
    Logger.log("5. Test getMembers_:");
    var membersResult = getMembers_(ctx, {});
    if (membersResult.success) {
      Logger.log("   ✅ Members OK — " + membersResult.data.length);
    }
  } else {
    Logger.log("   ❌ Login gagal: " + loginResult.message);
  }

  Logger.log("");
  Logger.log("✅ Cleanup test selesai");
}

function diagnoseTimezone() {
  Logger.log("=== DIAGNOSE TIMEZONE ===");
  Logger.log("");

  Logger.log("Script timezone: " + Session.getScriptTimeZone());
  Logger.log("new Date().toISOString(): " + new Date().toISOString());
  Logger.log("new Date().getDate(): " + new Date().getDate());
  Logger.log("new Date().getMonth()+1: " + (new Date().getMonth() + 1));
  Logger.log("");

  // Cek Date object dari sheet
  var sheet = new SheetRepository_("meetings")._sheet();
  var values = sheet.getRange(2, 2, 3, 1).getValues(); // kolom B (tanggal), 3 baris

  Logger.log("Date object dari sheet:");
  values.forEach(function (row, i) {
    var v = row[0];
    Logger.log("Row " + (i + 2) + ":");
    Logger.log("  Raw value: " + v);
    Logger.log("  Type: " + (v instanceof Date ? "Date" : typeof v));

    if (v instanceof Date) {
      Logger.log("  toISOString(): " + v.toISOString());
      Logger.log("  getDate(): " + v.getDate());
      Logger.log("  getMonth()+1: " + (v.getMonth() + 1));
      Logger.log("  getFullYear(): " + v.getFullYear());
      Logger.log(
        "  Utilities.formatDate(WIB): " +
          Utilities.formatDate(v, "Asia/Jakarta", "yyyy-MM-dd"),
      );
      Logger.log(
        "  Utilities.formatDate(script): " +
          Utilities.formatDate(v, Session.getScriptTimeZone(), "yyyy-MM-dd"),
      );
    }
    Logger.log("");
  });

  Logger.log("=== YANG DICARI ===");
  Logger.log(
    "getDate() vs Utilities.formatDate(WIB) — kalau beda, kita tahu biang keroknya.",
  );
}

function testTanggalFixFinal() {
  Logger.log("=== TEST TANGGAL FIX FINAL ===");
  Logger.log("");

  var login = login_({ username: "albetsurya", password: "albetsurya123" });
  var ctx = validateSession_(login.data.token);

  var result = getMeetings_(ctx, {});
  if (!result.success) {
    Logger.log("❌ Gagal: " + result.message);
    return;
  }

  Logger.log("Expected mapping (dari sheet display):");
  Logger.log("  MTGACDA6421 → 2026-09-15 (Selasa)");
  Logger.log("  MTG78094BF1 → 2026-09-14 (Senin)");
  Logger.log("  MTGBE1CB9BF → 2026-09-13 (Minggu)");
  Logger.log("");
  Logger.log("Actual dari getMeetings_:");

  result.data.slice(0, 3).forEach(function (m) {
    var ok = /^\d{4}-\d{2}-\d{2}$/.test(m.tanggal);
    Logger.log(
      "  " +
        (ok ? "✅" : "❌") +
        " " +
        m.meeting_id +
        " → tanggal: " +
        m.tanggal +
        " | hari: " +
        m.hari,
    );
  });
}

function testFormatDateQuick() {
  Logger.log("=== TEST FORMATDATE QUICK ===");
  Logger.log("");

  var sheet = new SheetRepository_("meetings")._sheet();
  var values = sheet.getRange(2, 2, 3, 1).getValues();

  values.forEach(function (row, i) {
    var v = row[0];
    Logger.log("Row " + (i + 2) + ":");
    Logger.log("  Raw: " + v);
    Logger.log("  formatDate(raw): " + formatDate(v));

    // Cek implementasi formatDate
    var expected = Utilities.formatDate(v, "Asia/Jakarta", "yyyy-MM-dd");
    var actual = formatDate(v);
    Logger.log(
      "  " +
        (actual === expected ? "✅" : "❌") +
        " expected: " +
        expected +
        " vs actual: " +
        actual,
    );
    Logger.log("");
  });

  Logger.log("=== CEK IMPLEMENTASI ===");
  Logger.log("formatDate source (50 char pertama):");
  Logger.log(formatDate.toString().slice(0, 200));
}

function diagnoseFormatDateDuplicate() {
  Logger.log("=== DIAGNOSE FORMATDATE DUPLIKASI ===");
  Logger.log("");

  // 1. Cek source aktual
  var source = formatDate.toString();
  Logger.log("1. Panjang source: " + source.length + " char");
  Logger.log("");
  Logger.log("Isi source LENGKAP:");
  Logger.log("───────────────");
  Logger.log(source);
  Logger.log("───────────────");
  Logger.log("");

  // 2. Cek apakah ada toISOString (bug)
  var hasToISO = source.indexOf("toISOString") !== -1;
  Logger.log("2. Analisis:");
  Logger.log("   toISOString: " + (hasToISO ? "❌ ADA (bug)" : "✅ TIDAK ADA"));
  Logger.log(
    "   Utilities.formatDate: " +
      (source.indexOf("Utilities.formatDate") !== -1 ? "✅" : "❌"),
  );
  Logger.log(
    "   +7 jam fallback: " +
      (source.indexOf("+ 7") !== -1 || source.indexOf("+7") !== -1
        ? "✅"
        : "❌"),
  );
  Logger.log("");

  // 3. Test langsung dengan Date object
  Logger.log("3. Test langsung:");
  var testDate = new Date("1998-01-01T00:00:00+07:00"); // WIB
  Logger.log("   Input: " + testDate);
  Logger.log("   toISOString: " + testDate.toISOString());
  Logger.log("   getDate() [runtime]: " + testDate.getDate());
  Logger.log("   getUTCDate() [runtime]: " + testDate.getUTCDate());
  Logger.log("   formatDate(input): " + formatDate(testDate));
  Logger.log("   Expected: 1998-01-01");
  Logger.log(
    "   " + (formatDate(testDate) === "1998-01-01" ? "✅ BENAR" : "❌ SALAH"),
  );
}

function cekDuplikasiFormatDate() {
  Logger.log("=== CARI DUPLIKASI formatDate ===");
  Logger.log("");
  Logger.log("Buka editor GAS:");
  Logger.log("  1. Tekan Ctrl+Shift+F (Find di semua file)");
  Logger.log("  2. Cari: function formatDate");
  Logger.log("  3. Hitung berapa hasil");
  Logger.log("");
  Logger.log("Kalau lebih dari 1 → ADA DUPLIKASI, hapus yang salah.");
}
