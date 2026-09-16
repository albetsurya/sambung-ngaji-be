function testSetupSpreadsheet() {
  Logger.log("=== TEST SETUP SPREADSHEET ===");

  var ss = getSpreadsheet_();
  var expectedSheets = Object.keys(SHEETS);

  Logger.log("Expected sheets: " + expectedSheets.length);
  Logger.log("");

  var missing = [];
  var ok = [];

  expectedSheets.forEach(function (key) {
    var def = SHEETS[key];
    var sheet = ss.getSheetByName(def.name);
    if (!sheet) {
      missing.push(def.name);
      Logger.log("❌ Sheet missing: " + def.name);
    } else {
      ok.push(def.name);
      var lastCol = sheet.getLastColumn();
      var headerRow = sheet
        .getRange(1, 1, 1, def.headers.length)
        .getValues()[0];
      var hasHeader = headerRow.join("") !== "";
      Logger.log(
        "✅ " +
          def.name +
          " (kolom: " +
          lastCol +
          ", header: " +
          (hasHeader ? "ok" : "kosong") +
          ")",
      );
    }
  });

  Logger.log("");
  Logger.log("Total OK: " + ok.length);
  Logger.log("Total missing: " + missing.length);

  if (missing.length > 0) {
    Logger.log("");
    Logger.log(
      "⚠️ Jalankan setupSpreadsheet() untuk membuat sheet yang missing.",
    );
  }

  return { ok: ok, missing: missing };
}

function testSheetHeaders() {
  Logger.log("=== TEST SHEET HEADERS ===");

  var ss = getSpreadsheet_();

  Object.keys(SHEETS).forEach(function (key) {
    var def = SHEETS[key];
    var sheet = ss.getSheetByName(def.name);
    if (!sheet) {
      Logger.log("❌ " + def.name + " — sheet tidak ada");
      return;
    }

    var headerRow = sheet.getRange(1, 1, 1, def.headers.length).getValues()[0];

    var match = true;
    for (var i = 0; i < def.headers.length; i++) {
      if (headerRow[i] !== def.headers[i]) {
        match = false;
        break;
      }
    }

    if (match) {
      Logger.log(
        "✅ " +
          def.name +
          " — headers match (" +
          def.headers.length +
          " kolom)",
      );
    } else {
      Logger.log("❌ " + def.name + " — headers tidak match");
      Logger.log("   Expected: " + def.headers.join(", "));
      Logger.log("   Actual:   " + headerRow.join(", "));
    }
  });
}

function testRepositoryCRUD() {
  Logger.log("=== TEST REPOSITORY CRUD ===");

  var repo = new SheetRepository_("settings");

  var testKey = "__test_repo_" + Date.now();
  var testValue = JSON.stringify({ test: true, ts: Date.now() });

  Logger.log("1. INSERT");
  var inserted = repo.insert({
    key: testKey,
    value: testValue,
    updated_at: nowIso_(),
  });
  Logger.log("   ✅ Inserted: " + inserted.key);

  Logger.log("");
  Logger.log("2. FIND BY ID");
  var found = repo.findById("key", testKey);
  if (found && found.key === testKey) {
    Logger.log("   ✅ Found: " + found.key);
  } else {
    Logger.log("   ❌ Not found");
  }

  Logger.log("");
  Logger.log("3. UPDATE");
  var updated = repo.updateById("key", testKey, {
    value: JSON.stringify({ test: "updated" }),
    updated_at: nowIso_(),
  });
  if (updated && updated.value.indexOf("updated") !== -1) {
    Logger.log("   ✅ Updated");
  } else {
    Logger.log("   ❌ Update failed");
  }

  Logger.log("");
  Logger.log("4. DELETE");
  var deleted = repo.deleteById("key", testKey);
  if (deleted) {
    Logger.log("   ✅ Deleted");
  } else {
    Logger.log("   ❌ Delete failed");
  }

  Logger.log("");
  Logger.log("5. VERIFY DELETED");
  var afterDelete = repo.findById("key", testKey);
  if (!afterDelete) {
    Logger.log("   ✅ Confirmed deleted");
  } else {
    Logger.log("   ❌ Still exists");
  }
}

function testUserLogin() {
  Logger.log("=== TEST USER LOGIN ===");

  var result = login_({
    username: "superadmin",
    password: "ganti123",
  });

  if (result.success) {
    Logger.log("✅ Login OK");
    Logger.log("   Token: " + result.data.token.substring(0, 20) + "...");
    Logger.log("   User: " + JSON.stringify(result.data.user, null, 2));
    return result.data.token;
  } else {
    Logger.log("❌ Login failed: " + result.message);
    return null;
  }
}

function testUserLoginWrongPassword() {
  Logger.log("=== TEST LOGIN — WRONG PASSWORD ===");

  var result = login_({
    username: "superadmin",
    password: "salah",
  });

  if (!result.success) {
    Logger.log("✅ Correctly rejected: " + result.message);
  } else {
    Logger.log("❌ Should have failed!");
  }
}

function testValidateSession() {
  Logger.log("=== TEST VALIDATE SESSION ===");

  var loginResult = login_({
    username: "superadmin",
    password: "ganti123",
  });

  if (!loginResult.success) {
    Logger.log("❌ Login failed");
    return;
  }

  var token = loginResult.data.token;
  var ctx = validateSession_(token);

  if (ctx && ctx.user) {
    Logger.log("✅ Session valid");
    Logger.log("   User: " + ctx.user.nama + " (" + ctx.user.role + ")");
  } else {
    Logger.log("❌ Session invalid");
  }

  Logger.log("");
  Logger.log("Test invalid token:");
  var badCtx = validateSession_("invalid-token-xxx");
  if (!badCtx) {
    Logger.log("✅ Invalid token correctly rejected");
  } else {
    Logger.log("❌ Invalid token should be rejected");
  }
}

function testPublicRegistration() {
  Logger.log("=== TEST PUBLIC REGISTRATION ===");

  var timestamp = Date.now();
  var testWa = "62" + String(timestamp).slice(-10);

  var result = submitPublicRegistration_(null, {
    nama_lengkap: "Test Jamaah " + timestamp,
    nama_panggilan: "Test",
    jenis_kelamin: "L",
    no_wa: testWa,
    tempat_lahir: "Surabaya",
    tanggal_lahir: "2000-01-15",
    alamat_rumah: "Jl. Test No. 1",
    desa: "Test Desa",
    daerah: "Test Daerah",
    pekerjaan: "Karyawan",
    hobi: "Membaca",
    is_nikah: false,
    jenjang_pendidikan: "S1",
    sekolah: "ITS",
    jurusan: "Teknik Informatika",
    tahun_mulai_pendidikan: "2018",
    tahun_selesai_pendidikan: "2022",
    _client_ip: "test-ip-" + timestamp,
  });

  if (result.success) {
    Logger.log("✅ Registration submitted");
    Logger.log("   Submission ID: " + result.data.submission_id);
    Logger.log("   Nama: " + result.data.nama_lengkap);
    return result.data.submission_id;
  } else {
    Logger.log("❌ Registration failed: " + result.message);
    return null;
  }
}

function testDebugDuplicate() {
  Logger.log("=== DEBUG DUPLICATE ===");

  var timestamp = Date.now();
  var testWa = "62" + String(timestamp).slice(-10);

  Logger.log("Test WA: " + testWa);

  Logger.log("");
  Logger.log("1. Cek cache status:");
  var cached = CacheService.getScriptCache().get("sheet_pending_members");
  Logger.log(
    "   Cache ada: " + (cached ? "ya (" + cached.length + " bytes)" : "tidak"),
  );

  Logger.log("");
  Logger.log("2. Submit pertama:");
  var r1 = submitPublicRegistration_(null, {
    nama_lengkap: "Debug User 1",
    jenis_kelamin: "L",
    no_wa: testWa,
    _client_ip: "debug-ip-1",
  });
  Logger.log("   Result: " + (r1.success ? "✅ OK" : "❌ " + r1.message));
  if (r1.success) Logger.log("   Submission: " + r1.data.submission_id);

  Logger.log("");
  Logger.log("3. Cek cache setelah insert:");
  var cachedAfter = CacheService.getScriptCache().get("sheet_pending_members");
  Logger.log(
    "   Cache ada: " +
      (cachedAfter ? "ya (" + cachedAfter.length + " bytes)" : "tidak"),
  );

  Logger.log("");
  Logger.log("4. Cek isi sheet pending_members:");
  var repo = new SheetRepository_("pending_members");
  var all = repo.getAll();
  Logger.log("   Total rows: " + all.length);

  all.forEach(function (p, i) {
    Logger.log(
      "   [" +
        i +
        "] no_wa=" +
        p.no_wa +
        " (tipe: " +
        typeof p.no_wa +
        "), status=" +
        p.status,
    );
  });

  Logger.log("");
  Logger.log("5. Cek normalisasi:");
  var normalized = normalizePhoneNumber(testWa);
  Logger.log("   Input:     '" + testWa + "' (tipe: " + typeof testWa + ")");
  Logger.log(
    "   Normalized: '" + normalized + "' (tipe: " + typeof normalized + ")",
  );

  Logger.log("");
  Logger.log("6. Simulasi find:");
  var found = repo.find(function (p) {
    var match =
      String(p.no_wa) === String(normalized) &&
      p.status === PENDING_STATUS.PENDING;
    Logger.log(
      "   Row: p.no_wa='" +
        p.no_wa +
        "' vs normalized='" +
        normalized +
        "' → " +
        match,
    );
    return match;
  });
  Logger.log("   Found: " + found.length);

  Logger.log("");
  Logger.log("7. Submit kedua dengan no WA sama:");
  var r2 = submitPublicRegistration_(null, {
    nama_lengkap: "Debug User 2",
    jenis_kelamin: "P",
    no_wa: testWa,
    _client_ip: "debug-ip-2",
  });
  Logger.log(
    "   Result: " +
      (r2.success
        ? "❌ BERHASIL (HARUSNYA GAGAL!)"
        : "✅ Rejected: " + r2.message),
  );

  Logger.log("");
  Logger.log("8. Cleanup:");
  var testRows = repo.find(function (p) {
    return (
      String(p.no_wa).indexOf("62") === 0 &&
      p.nama_lengkap.indexOf("Debug User") === 0
    );
  });
  testRows.forEach(function (p) {
    repo.deleteById("submission_id", p.submission_id);
  });
  Logger.log("   Deleted " + testRows.length + " test rows");
}

function testPublicRegistrationValidation() {
  Logger.log("=== TEST PUBLIC REGISTRATION — VALIDATION ===");

  Logger.log("Test 1: Nama kosong");
  var r1 = submitPublicRegistration_(null, {
    nama_lengkap: "",
    jenis_kelamin: "L",
    no_wa: "081234567890",
  });
  Logger.log(r1.success ? "❌ Should fail" : "✅ Rejected: " + r1.message);

  Logger.log("");
  Logger.log("Test 2: Nama terlalu pendek");
  var r2 = submitPublicRegistration_(null, {
    nama_lengkap: "AB",
    jenis_kelamin: "L",
    no_wa: "081234567890",
  });
  Logger.log(r2.success ? "❌ Should fail" : "✅ Rejected: " + r2.message);

  Logger.log("");
  Logger.log("Test 3: Jenis kelamin invalid");
  var r3 = submitPublicRegistration_(null, {
    nama_lengkap: "Test User",
    jenis_kelamin: "X",
    no_wa: "081234567890",
  });
  Logger.log(r3.success ? "❌ Should fail" : "✅ Rejected: " + r3.message);

  Logger.log("");
  Logger.log("Test 4: No WA kosong");
  var r4 = submitPublicRegistration_(null, {
    nama_lengkap: "Test User",
    jenis_kelamin: "L",
    no_wa: "",
  });
  Logger.log(r4.success ? "❌ Should fail" : "✅ Rejected: " + r4.message);

  Logger.log("");
  Logger.log("Test 5: No WA terlalu pendek");
  var r5 = submitPublicRegistration_(null, {
    nama_lengkap: "Test User",
    jenis_kelamin: "L",
    no_wa: "08123",
  });
  Logger.log(r5.success ? "❌ Should fail" : "✅ Rejected: " + r5.message);
}

function testPublicRegistrationDuplicate() {
  Logger.log("=== TEST PUBLIC REGISTRATION — DUPLICATE ===");

  var timestamp = Date.now();
  var testWa = "62" + String(timestamp).slice(-10);

  Logger.log("1. Submit pertama (should succeed):");
  var r1 = submitPublicRegistration_(null, {
    nama_lengkap: "Dup Test 1",
    jenis_kelamin: "L",
    no_wa: testWa,
    _client_ip: "test-ip-dup-1",
  });
  Logger.log(
    r1.success ? "✅ OK: " + r1.data.submission_id : "❌ Failed: " + r1.message,
  );

  Logger.log("");
  Logger.log("2. Submit kedua dengan no WA sama (should fail):");
  var r2 = submitPublicRegistration_(null, {
    nama_lengkap: "Dup Test 2",
    jenis_kelamin: "P",
    no_wa: testWa,
    _client_ip: "test-ip-dup-2",
  });
  Logger.log(
    r2.success ? "❌ Should fail but succeeded" : "✅ Rejected: " + r2.message,
  );
}

function testVerifyNoWaType() {
  Logger.log("=== VERIFY no_wa TYPE IN SHEET ===");

  var sheetsToCheck = ["members", "pending_members"];

  sheetsToCheck.forEach(function (name) {
    var repo = new SheetRepository_(name);
    var sheet = repo._sheet();
    var headers = repo.def.headers;
    var colIndex = headers.indexOf("no_wa") + 1;

    if (colIndex === 0) {
      Logger.log("❌ " + name + " tidak punya kolom no_wa");
      return;
    }

    var lastRow = sheet.getLastRow();
    if (lastRow < 2) {
      Logger.log("ℹ️ " + name + " kosong");
      return;
    }

    var values = sheet.getRange(2, colIndex, lastRow - 1, 1).getValues();

    var stringCount = 0;
    var numberCount = 0;

    values.forEach(function (row) {
      var val = row[0];
      if (val === "" || val === null || val === undefined) return;
      if (typeof val === "string") stringCount++;
      else if (typeof val === "number") numberCount++;
    });

    Logger.log("📊 " + name + ".no_wa:");
    Logger.log("   String: " + stringCount);
    Logger.log("   Number: " + numberCount);

    if (numberCount > 0) {
      Logger.log(
        "   ⚠️ Ada " + numberCount + " baris yang tersimpan sebagai NUMBER",
      );
      Logger.log("   → Jalankan migratePhoneNumbersToText() untuk fix");
    } else {
      Logger.log("   ✅ Semua string");
    }
  });
}

function testPendingMemberFlow() {
  Logger.log("=== TEST PENDING MEMBER FLOW ===");

  var submissionId = testPublicRegistration();
  if (!submissionId) {
    Logger.log("❌ Setup failed");
    return;
  }

  Logger.log("");
  Logger.log("1. Get pending members (admin)");
  var ctx = {
    user: {
      user_id: "USR001",
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
    },
  };

  var listResult = getPendingMembers_(ctx, { status: PENDING_STATUS.PENDING });
  if (listResult.success) {
    Logger.log("   ✅ Found " + listResult.data.length + " pending");
  } else {
    Logger.log("   ❌ Failed: " + listResult.message);
  }

  Logger.log("");
  Logger.log("2. Get detail");
  var detailResult = getPendingMemberDetail_(ctx, {
    submission_id: submissionId,
  });
  if (detailResult.success) {
    Logger.log("   ✅ Detail: " + detailResult.data.nama_lengkap);
  } else {
    Logger.log("   ❌ Failed: " + detailResult.message);
  }

  Logger.log("");
  Logger.log("3. Approve");
  var approveResult = approvePendingMember_(ctx, {
    submission_id: submissionId,
    kelompok: "",
    create_user: false,
  });
  if (approveResult.success) {
    Logger.log("   ✅ Approved");
    Logger.log("   Member ID: " + approveResult.data.member_id);
  } else {
    Logger.log("   ❌ Failed: " + approveResult.message);
  }

  Logger.log("");
  Logger.log("4. Verify member created");
  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", approveResult.data.member_id);
  if (member) {
    Logger.log("   ✅ Member exists: " + member.nama_lengkap);
  } else {
    Logger.log("   ❌ Member not found");
  }

  Logger.log("");
  Logger.log("5. Verify pending status updated");
  var updated = pendingRepo_findById(submissionId);
  if (updated && updated.status === PENDING_STATUS.APPROVED) {
    Logger.log("   ✅ Status: " + updated.status);
  } else {
    Logger.log("   ❌ Status not updated");
  }
}

function pendingRepo_findById(submissionId) {
  var repo = new SheetRepository_("pending_members");
  return repo.findById("submission_id", submissionId);
}

function testPendingMemberReject() {
  Logger.log("=== TEST PENDING MEMBER REJECT ===");

  var submissionId = testPublicRegistration();
  if (!submissionId) return;

  var ctx = {
    user: {
      user_id: "USR001",
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
    },
  };

  Logger.log("Reject with reason:");
  var result = rejectPendingMember_(ctx, {
    submission_id: submissionId,
    reason: "Data tidak lengkap",
  });

  if (result.success) {
    Logger.log("   ✅ Rejected");
    var updated = pendingRepo_findById(submissionId);
    Logger.log("   Status: " + updated.status);
    Logger.log("   Reason: " + updated.rejection_reason);
  } else {
    Logger.log("   ❌ Failed: " + result.message);
  }
}

function testApproveWithCreateUser() {
  Logger.log("=== TEST APPROVE WITH CREATE USER ===");

  var submissionId = testPublicRegistration();
  if (!submissionId) return;

  var ctx = {
    user: {
      user_id: "USR001",
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
    },
  };

  var pending = pendingRepo_findById(submissionId);
  var username = "member_" + Date.now();
  var password = "test123456";

  var result = approvePendingMember_(ctx, {
    submission_id: submissionId,
    kelompok: "",
    create_user: true,
    username: username,
    password: password,
  });

  if (result.success) {
    Logger.log("   ✅ Approved");
    Logger.log("   Member ID: " + result.data.member_id);
    Logger.log("   Created user: " + JSON.stringify(result.data.created_user));

    if (result.data.created_user) {
      Logger.log("");
      Logger.log("Test login with new user:");
      var loginResult = login_({
        username: username,
        password: password,
      });
      if (loginResult.success) {
        Logger.log("   ✅ Login OK");
        Logger.log("   Role: " + loginResult.data.user.role);
        Logger.log("   Member ID: " + loginResult.data.user.member_id);
      } else {
        Logger.log("   ❌ Login failed: " + loginResult.message);
      }
    }
  } else {
    Logger.log("   ❌ Failed: " + result.message);
  }
}

function testMemberSelfProfile() {
  Logger.log("=== TEST MEMBER SELF PROFILE ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];

  if (!firstMember) {
    Logger.log("❌ Tidak ada member. Buat dulu via testPendingMemberFlow()");
    return;
  }

  var ctx = {
    user: {
      user_id: "TEST_MEMBER",
      nama: firstMember.nama_lengkap,
      role: ROLES.MEMBER,
      member_id: firstMember.member_id,
    },
  };

  Logger.log("Member ID: " + firstMember.member_id);
  Logger.log("");

  var result = getMyProfile_(ctx, {});
  if (result.success) {
    Logger.log("✅ Profile OK");
    Logger.log("   Nama: " + result.data.nama_lengkap);
    Logger.log("   Kategori: " + result.data.kategori);
    Logger.log("   Usia: " + result.data.usia);
    Logger.log("");

    Logger.log("   Fields yang dikembalikan:");
    Object.keys(result.data).forEach(function (k) {
      Logger.log("   - " + k);
    });

    Logger.log("");
    Logger.log("   Cek field sensitif TIDAK ada:");
    var forbidden = ["is_muballigh", "created_at", "updated_at"];
    forbidden.forEach(function (f) {
      if (result.data.hasOwnProperty(f)) {
        Logger.log("   ❌ " + f + " — seharusnya tidak ada!");
      } else {
        Logger.log("   ✅ " + f + " — tidak ada");
      }
    });
  } else {
    Logger.log("❌ Failed: " + result.message);
  }
}

function testMemberSelfAttendance() {
  Logger.log("=== TEST MEMBER SELF ATTENDANCE ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];
  if (!firstMember) {
    Logger.log("❌ Tidak ada member");
    return;
  }

  var ctx = {
    user: {
      user_id: "TEST_MEMBER",
      nama: firstMember.nama_lengkap,
      role: ROLES.MEMBER,
      member_id: firstMember.member_id,
    },
  };

  var result = getMyAttendance_(ctx, {});
  if (result.success) {
    Logger.log("✅ Attendance OK");
    Logger.log("   Total: " + result.data.length);
    if (result.data.length > 0) {
      Logger.log("   Contoh:");
      Logger.log("   " + JSON.stringify(result.data[0], null, 2));
    }
  } else {
    Logger.log("❌ Failed: " + result.message);
  }
}

function testMemberSelfMonitoring() {
  Logger.log("=== TEST MEMBER SELF MONITORING ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];
  if (!firstMember) {
    Logger.log("❌ Tidak ada member");
    return;
  }

  var ctx = {
    user: {
      user_id: "TEST_MEMBER",
      nama: firstMember.nama_lengkap,
      role: ROLES.MEMBER,
      member_id: firstMember.member_id,
    },
  };

  var result = getMyMonitoring_(ctx, {});
  if (result.success) {
    Logger.log("✅ Monitoring OK");
    Logger.log("   Total: " + result.data.length);
  } else {
    Logger.log("❌ Failed: " + result.message);
  }
}

function testMemberSelfUpcomingMeetings() {
  Logger.log("=== TEST MEMBER SELF UPCOMING MEETINGS ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];
  if (!firstMember) {
    Logger.log("❌ Tidak ada member");
    return;
  }

  var ctx = {
    user: {
      user_id: "TEST_MEMBER",
      nama: firstMember.nama_lengkap,
      role: ROLES.MEMBER,
      member_id: firstMember.member_id,
    },
  };

  var result = getUpcomingMeetings_(ctx, { limit: 5 });
  if (result.success) {
    Logger.log("✅ Upcoming OK");
    Logger.log("   Total: " + result.data.length);
    result.data.forEach(function (m, i) {
      Logger.log(
        "   " + (i + 1) + ". " + m.hari + " " + m.tanggal + " — " + m.acara,
      );
    });
  } else {
    Logger.log("❌ Failed: " + result.message);
  }
}

function testMemberSelfUpdateProfile() {
  Logger.log("=== TEST MEMBER SELF UPDATE PROFILE ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];
  if (!firstMember) {
    Logger.log("❌ Tidak ada member");
    return;
  }

  var ctx = {
    user: {
      user_id: "TEST_MEMBER",
      nama: firstMember.nama_lengkap,
      role: ROLES.MEMBER,
      member_id: firstMember.member_id,
    },
  };

  Logger.log("Before update:");
  Logger.log("   nama_panggilan: " + firstMember.nama_panggilan);
  Logger.log("   hobi: " + firstMember.hobi);
  Logger.log("");

  var testHobi = "Testing " + new Date().toISOString();

  var result = updateMyProfile_(ctx, {
    nama_panggilan: "Test Update",
    hobi: testHobi,
    no_wa: "081234567890",
  });

  if (result.success) {
    Logger.log("✅ Update OK");
    Logger.log("   nama_panggilan: " + result.data.nama_panggilan);
    Logger.log("   hobi: " + result.data.hobi);
    Logger.log("   no_wa: " + result.data.no_wa);
  } else {
    Logger.log("❌ Failed: " + result.message);
  }

  Logger.log("");
  Logger.log("Test forbidden field (nama_lengkap):");
  var result2 = updateMyProfile_(ctx, {
    nama_lengkap: "HACKED",
  });
  Logger.log(
    "   Result: " + (result2.success ? "❌ Should fail" : "✅ Rejected"),
  );

  if (result2.success) {
    Logger.log("   Cek apakah nama_lengkap berubah:");
    var after = membersRepo.findById("member_id", firstMember.member_id);
    Logger.log("   Nama sekarang: " + after.nama_lengkap);
  }
}

function testMemberSelfAccessWithoutLink() {
  Logger.log("=== TEST MEMBER ACCESS WITHOUT LINK ===");

  var ctx = {
    user: {
      user_id: "TEST",
      nama: "Test",
      role: ROLES.MEMBER,
      member_id: "",
    },
  };

  var result = getMyProfile_(ctx, {});
  if (!result.success) {
    Logger.log("✅ Correctly rejected: " + result.message);
  } else {
    Logger.log("❌ Should have failed!");
  }
}

function testAiQuotaCheck() {
  Logger.log("=== TEST AI QUOTA CHECK ===");

  var ctx = {
    user: {
      user_id: "TEST_QUOTA_USER",
      nama: "Test Quota",
      role: ROLES.MEMBER,
    },
  };

  var usageRepo = new SheetRepository_("ai_usage");
  var today = formatDate(nowIso_());
  var testUserId = "TEST_QUOTA_USER";

  Logger.log("1. Cek quota awal (harus kosong):");
  var existing = usageRepo.find(function (u) {
    return u.user_id === testUserId && String(u.timestamp).indexOf(today) === 0;
  });
  Logger.log("   Existing usage: " + existing.length);

  Logger.log("");
  Logger.log("2. Insert 10 usage dummy:");

  var toInsert = [];
  for (var i = 0; i < 10; i++) {
    toInsert.push({
      usage_id: generateUsageId(),
      user_id: testUserId,
      user_nama: "Test Quota",
      role: ROLES.MEMBER,
      provider: "omniroute",
      input_tokens: 100,
      output_tokens: 50,
      total_tokens: 150,
      timestamp: nowIso_(),
    });
  }
  usageRepo.insertMany(toInsert);
  Logger.log("   ✅ Inserted 10 usage");

  Logger.log("");
  Logger.log("3. Cek quota (harus limit tercapai):");
  var quotaCheck = checkAiQuota_(ctx);
  if (quotaCheck && !quotaCheck.success) {
    Logger.log("   ✅ Quota limit reached: " + quotaCheck.message);
  } else {
    Logger.log("   ❌ Should have reached limit");
  }

  Logger.log("");
  Logger.log("4. Cleanup test data:");
  var testRows = usageRepo.find(function (u) {
    return u.user_id === testUserId;
  });
  testRows.forEach(function (r) {
    usageRepo.deleteById("usage_id", r.usage_id);
  });
  Logger.log("   ✅ Cleaned " + testRows.length + " rows");
}

function testAiUsageLogging() {
  Logger.log("=== TEST AI USAGE LOGGING ===");

  var ctx = {
    user: {
      user_id: "TEST_LOG_USER",
      nama: "Test Log",
      role: ROLES.MEMBER,
    },
  };

  Logger.log("Before logging:");
  var repo = new SheetRepository_("ai_usage");
  var before = repo.find(function (u) {
    return u.user_id === "TEST_LOG_USER";
  });
  Logger.log("   Existing: " + before.length);

  Logger.log("");
  Logger.log("Log usage:");
  logAiUsage_(ctx, "omniroute", 150, 75);
  Logger.log("   ✅ Logged");

  Logger.log("");
  Logger.log("After logging:");
  var after = repo.find(function (u) {
    return u.user_id === "TEST_LOG_USER";
  });
  Logger.log("   Total: " + after.length);

  if (after.length > 0) {
    var latest = after[after.length - 1];
    Logger.log("   Latest:");
    Logger.log("   - provider: " + latest.provider);
    Logger.log("   - input_tokens: " + latest.input_tokens);
    Logger.log("   - output_tokens: " + latest.output_tokens);
    Logger.log("   - total_tokens: " + latest.total_tokens);
  }

  Logger.log("");
  Logger.log("Cleanup:");
  after.forEach(function (r) {
    repo.deleteById("usage_id", r.usage_id);
  });
  Logger.log("   ✅ Cleaned");
}

function testDeleteAttendance() {
  Logger.log("=== TEST DELETE ATTENDANCE ===");

  var attendanceRepo = new SheetRepository_("attendance");
  var firstAttendance = attendanceRepo.getAll()[0];

  if (!firstAttendance) {
    Logger.log("⚠️ Tidak ada attendance. Skip test.");
    return;
  }

  var ctx = {
    user: {
      user_id: "USR001",
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
    },
  };

  var attendanceId = firstAttendance.attendance_id;
  Logger.log("Test attendance ID: " + attendanceId);

  Logger.log("");
  Logger.log("1. Backup data");
  var backup = Object.assign({}, firstAttendance);
  Logger.log("   ✅ Backed up");

  Logger.log("");
  Logger.log("2. Delete");
  var result = deleteAttendance_(ctx, {
    meeting_id: firstAttendance.meeting_id,
    member_id: firstAttendance.member_id,
  });

  if (result.success && result.data.deleted === 1) {
    Logger.log("   ✅ Deleted");
  } else {
    Logger.log("   ❌ Failed: " + JSON.stringify(result));
    return;
  }

  Logger.log("");
  Logger.log("3. Verify deleted");
  var afterDelete = attendanceRepo.findById("attendance_id", attendanceId);
  if (!afterDelete) {
    Logger.log("   ✅ Confirmed deleted");
  } else {
    Logger.log("   ❌ Still exists");
  }

  Logger.log("");
  Logger.log("4. Restore backup");
  var restoreRow = {
    attendance_id: backup.attendance_id,
    meeting_id: backup.meeting_id,
    member_id: backup.member_id,
    status: backup.status,
    catatan: backup.catatan,
    created_by: backup.created_by,
    created_at: backup.created_at,
    updated_at: backup.updated_at,
  };
  attendanceRepo.insert(restoreRow);
  Logger.log("   ✅ Restored");
}

function testUserCreateMember() {
  Logger.log("=== TEST CREATE USER — ROLE MEMBER ===");

  var membersRepo = new SheetRepository_("members");
  var firstMember = membersRepo.getAll()[0];

  if (!firstMember) {
    Logger.log("❌ Tidak ada member");
    return;
  }

  var ctx = {
    user: {
      user_id: "USR001",
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
    },
  };

  var username = "testmember_" + Date.now();

  Logger.log("Test 1: Create MEMBER tanpa member_id (should fail)");
  var r1 = createUser_(ctx, {
    username: username + "_1",
    password: "test123456",
    role: ROLES.MEMBER,
    nama: "Test",
    member_id: "",
  });
  Logger.log(r1.success ? "❌ Should fail" : "✅ Rejected: " + r1.message);

  Logger.log("");
  Logger.log("Test 2: Create MEMBER dengan member_id (should succeed)");
  var r2 = createUser_(ctx, {
    username: username + "_2",
    password: "test123456",
    role: ROLES.MEMBER,
    nama: "Test Member",
    member_id: firstMember.member_id,
  });

  if (r2.success) {
    Logger.log("   ✅ Created: " + r2.data.username);
    Logger.log("   Role: " + r2.data.role);
    Logger.log("   Member ID: " + r2.data.member_id);

    Logger.log("");
    Logger.log("Test login dengan user baru:");
    var loginResult = login_({
      username: username + "_2",
      password: "test123456",
    });
    if (loginResult.success) {
      Logger.log("   ✅ Login OK");
      Logger.log("   Role: " + loginResult.data.user.role);
    } else {
      Logger.log("   ❌ Login failed");
    }
  } else {
    Logger.log("   ❌ Failed: " + r2.message);
  }
}

function testFullIntegration() {
  Logger.log("╔════════════════════════════════════════════════╗");
  Logger.log("║       FULL BACKEND INTEGRATION TEST            ║");
  Logger.log("╚════════════════════════════════════════════════╝");
  Logger.log("");

  Logger.log("▶️  STEP 1: Setup & Config");
  testSetupSpreadsheet();
  Logger.log("");

  Logger.log("▶️  STEP 2: Sheet Headers");
  testSheetHeaders();
  Logger.log("");

  Logger.log("▶️  STEP 3: Repository CRUD");
  testRepositoryCRUD();
  Logger.log("");

  Logger.log("▶️  STEP 4: Auth");
  testUserLogin();
  testUserLoginWrongPassword();
  testValidateSession();
  Logger.log("");

  Logger.log("▶️  STEP 5: Public Registration");
  testPublicRegistration();
  testPublicRegistrationValidation();
  testPublicRegistrationDuplicate();
  Logger.log("");

  Logger.log("▶️  STEP 6: Pending Member Flow");
  testPendingMemberFlow();
  testPendingMemberReject();
  testApproveWithCreateUser();
  Logger.log("");

  Logger.log("▶️  STEP 7: Member Self-Service");
  testMemberSelfProfile();
  testMemberSelfAttendance();
  testMemberSelfMonitoring();
  testMemberSelfUpcomingMeetings();
  testMemberSelfUpdateProfile();
  testMemberSelfAccessWithoutLink();
  Logger.log("");

  Logger.log("▶️  STEP 8: AI Quota & Usage");
  testAiQuotaCheck();
  testAiUsageLogging();
  Logger.log("");

  Logger.log("▶️  STEP 9: Delete Attendance");
  testDeleteAttendance();
  Logger.log("");

  Logger.log("▶️  STEP 10: User Management");
  testUserCreateMember();
  Logger.log("");

  Logger.log("════════════════════════════════════════════════");
  Logger.log("✅ FULL INTEGRATION TEST SELESAI");
  Logger.log("════════════════════════════════════════════════");
}

function testParseParams() {
  Logger.log("=== TEST parseParams_ ===");

  // Simulasi request dari frontend
  var mockEvent = {
    parameter: {},
    postData: {
      type: "text/plain;charset=utf-8",
      contents: JSON.stringify({
        action: "login",
        username: "superadmin",
        password: "ganti123",
      }),
    },
  };

  var params = parseParams_(mockEvent);

  Logger.log("Input postData.type: " + mockEvent.postData.type);
  Logger.log("Input postData.contents: " + mockEvent.postData.contents);
  Logger.log("");
  Logger.log("Output params: " + JSON.stringify(params, null, 2));
  Logger.log("");
  Logger.log(
    "action ada? " + (params.action ? "✅ " + params.action : "❌ KOSONG"),
  );
}

function testSendWA() {
  var res = sendWhatsApp_("085791978786", "Test dari Pengajian ✅");
  Logger.log(res);
}

function testDeleteMeeting() {
  var fakeCtx = { user: { user_id: "test", role: "ADMIN" }, token: "test" };
  var result = deleteMeeting_(fakeCtx, {
    meeting_id: "MTG06746AB6", // ganti dengan meeting_id yang ada
  });
  Logger.log(result);
}
