function migrateAuditLogsAddUserName() {
  Logger.log("=== MIGRASI AUDIT LOGS — TAMBAH user_nama ===");

  var logsRepo = new SheetRepository_("audit_logs");
  var usersRepo = new SheetRepository_("users");
  var sheet = logsRepo._sheet();
  var headers = logsRepo.def.headers;

  var userNameColIdx = headers.indexOf("user_nama");
  if (userNameColIdx === -1) {
    Logger.log(
      "❌ Kolom user_nama tidak ada di schema. Update Config.js dulu.",
    );
    return;
  }

  var lastRow = sheet.getLastRow();
  if (lastRow < 2) {
    Logger.log("Sheet kosong.");
    return;
  }

  // Build users map
  var users = usersRepo.getAll();
  var usersById = {};
  users.forEach(function (u) {
    usersById[u.user_id] = u.nama || "";
  });

  // Baca kolom user_id dan user_nama
  var userIdColIdx = headers.indexOf("user_id");
  var range = sheet.getRange(2, 1, lastRow - 1, headers.length);
  var values = range.getValues();

  var updated = 0;
  for (var i = 0; i < values.length; i++) {
    var userId = values[i][userIdColIdx];
    var currentNama = values[i][userNameColIdx];

    if (!currentNama && userId && usersById[userId]) {
      values[i][userNameColIdx] = usersById[userId];
      updated++;
    }
  }

  if (updated > 0) {
    range.setValues(values);
    logsRepo._invalidateCache();
    Logger.log("✅ " + updated + " log di-update dengan user_nama");
  } else {
    Logger.log("Tidak ada log yang perlu di-update");
  }
}

function migratePhotoUrls() {
  Logger.log("=== MIGRATE FOTO URLS ===");

  var repo = new SheetRepository_("members");
  var members = repo.getAll();

  var migrated = 0;
  var skipped = 0;

  members.forEach(function (m) {
    if (!m.foto_url) {
      skipped++;
      return;
    }

    var idMatch = String(m.foto_url).match(/[?&]id=([^&]+)/);
    if (!idMatch) {
      skipped++;
      Logger.log("⚠️ Skip (no id): " + m.nama_lengkap + " → " + m.foto_url);
      return;
    }

    var fileId = idMatch[1];

    if (m.foto_url.indexOf("thumbnail?") !== -1) {
      skipped++;
      Logger.log("⏭️ Skip (already new): " + m.nama_lengkap);
      return;
    }

    var newUrl =
      "https://drive.google.com/thumbnail?id=" + fileId + "&sz=w1000";

    repo.updateById("member_id", m.member_id, {
      foto_url: newUrl,
      updated_at: nowIso_(),
    });

    migrated++;
    Logger.log("✅ " + m.nama_lengkap);
    Logger.log("   OLD: " + m.foto_url);
    Logger.log("   NEW: " + newUrl);
  });

  Logger.log("");
  Logger.log("═══════════════════════════════════════");
  Logger.log("Migrated: " + migrated);
  Logger.log("Skipped:  " + skipped);
  Logger.log("Total:    " + members.length);
  Logger.log("═══════════════════════════════════════");
}

function migratePhoneNumbersToText() {
  Logger.log("=== MIGRATE PHONE NUMBERS TO TEXT ===");

  var sheetsToFix = [
    { key: "members", field: "no_wa" },
    { key: "pending_members", field: "no_wa" },
  ];

  sheetsToFix.forEach(function (cfg) {
    var repo = new SheetRepository_(cfg.key);
    var sheet = repo._sheet();
    var headers = repo.def.headers;
    var colIndex = headers.indexOf(cfg.field) + 1;

    if (colIndex === 0) {
      Logger.log("❌ Kolom " + cfg.field + " tidak ada di " + cfg.key);
      return;
    }

    var lastRow = sheet.getLastRow();
    if (lastRow < 2) {
      Logger.log("ℹ️ Sheet " + cfg.key + " kosong");
      return;
    }

    var range = sheet.getRange(2, colIndex, lastRow - 1, 1);
    var values = range.getValues();

    var newValues = values.map(function (row) {
      var val = row[0];
      if (val === null || val === undefined || val === "") return [""];
      return [String(val)];
    });

    range.setNumberFormat("@");
    range.setValues(newValues);

    repo._invalidateCache();

    Logger.log(
      "✅ " +
        cfg.key +
        "." +
        cfg.field +
        " — " +
        newValues.length +
        " rows migrated",
    );
  });

  Logger.log("");
  Logger.log("Selesai. Silakan verifikasi dengan testVerifyNoWaType().");
}
