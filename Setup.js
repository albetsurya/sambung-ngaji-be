function keepAlive() {
  return ok_({ ping: new Date().toISOString() });
}

function setupKeepAliveTrigger() {
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "keepAlive") {
      ScriptApp.deleteTrigger(t);
    }
  });

  ScriptApp.newTrigger("keepAlive").timeBased().everyMinutes(5).create();

  Logger.log("Keep-alive trigger created. Akan ping setiap 5 menit.");
}

/* -------------------------------------------------------------------------- */
/*                    DAILY AI QUOTA CLEANUP TRIGGER                          */
/* -------------------------------------------------------------------------- */

/**
 * Setup trigger harian untuk membersihkan key kuota AI dari hari-hari sebelumnya.
 * Jalankan SEKALI dari editor GAS. Trigger akan otomatis jalan setiap jam 2 pagi.
 *
 * Aman dijalankan ulang: akan hapus trigger lama dengan nama sama sebelum buat baru.
 */
function setupDailyAiQuotaCleanupTrigger() {
  var existing = ScriptApp.getProjectTriggers();
  existing.forEach(function (t) {
    if (t.getHandlerFunction() === "cleanupOldAiQuota_") {
      ScriptApp.deleteTrigger(t);
      Logger.log("Menghapus trigger lama: cleanupOldAiQuota_");
    }
  });

  ScriptApp.newTrigger("cleanupOldAiQuota_")
    .timeBased()
    .atHour(2)
    .everyDays(1)
    .create();

  Logger.log("✅ Trigger harian dibuat: cleanupOldAiQuota_ setiap jam 2 pagi.");

  // Verifikasi
  var allTriggers = ScriptApp.getProjectTriggers();
  Logger.log("");
  Logger.log("=== SEMUA TRIGGER AKTIF ===");
  allTriggers.forEach(function (t) {
    Logger.log("  - " + t.getHandlerFunction() + " (" + t.getEventType() + ")");
  });
}

/**
 * Cek daftar trigger yang sedang aktif (untuk verifikasi).
 */
function listAllTriggers() {
  var triggers = ScriptApp.getProjectTriggers();
  Logger.log("=== TRIGGERS ===");
  Logger.log("Total: " + triggers.length);
  Logger.log("");
  triggers.forEach(function (t, i) {
    Logger.log(
      i +
        1 +
        ". " +
        t.getHandlerFunction() +
        " | " +
        t.getEventType() +
        " | " +
        t.getTriggerSource(),
    );
  });
  return triggers.length;
}

/**
 * Test manual: jalankan cleanup sekali sekarang tanpa menunggu jam 2 pagi.
 * Berguna untuk verifikasi bahwa fungsi bekerja.
 */
function testRunCleanupNow() {
  Logger.log("=== TEST CLEANUP AI QUOTA (MANUAL) ===");
  Logger.log("");

  var props = PropertiesService.getScriptProperties();
  var keysBefore = props.getKeys().filter(function (k) {
    return k.indexOf("aiquota:") === 0;
  });
  Logger.log("Key sebelum cleanup: " + keysBefore.length);
  keysBefore.forEach(function (k) {
    Logger.log("  - " + k + " = " + props.getProperty(k));
  });

  Logger.log("");
  var removed = cleanupOldAiQuota_();

  Logger.log("");
  var keysAfter = props.getKeys().filter(function (k) {
    return k.indexOf("aiquota:") === 0;
  });
  Logger.log("Key setelah cleanup: " + keysAfter.length);
  keysAfter.forEach(function (k) {
    Logger.log("  - " + k + " = " + props.getProperty(k));
  });

  Logger.log("");
  Logger.log("Total dihapus: " + removed);
  return removed;
}

/* -------------------------------------------------------------------------- */
/*                          SPREADSHEET SETUP                                 */
/* -------------------------------------------------------------------------- */

function setupSpreadsheet() {
  var ss = getSpreadsheet_();
  Object.keys(SHEETS).forEach(function (key) {
    var def = SHEETS[key];
    var sheet = ss.getSheetByName(def.name);
    if (!sheet) {
      sheet = ss.insertSheet(def.name);
    }
    var firstRow = sheet.getRange(1, 1, 1, def.headers.length).getValues()[0];
    var hasHeader = firstRow.join("") !== "";
    if (!hasHeader) {
      sheet.getRange(1, 1, 1, def.headers.length).setValues([def.headers]);
      sheet.setFrozenRows(1);
      sheet.getRange(1, 1, 1, def.headers.length).setFontWeight("bold");
    }
  });

  var sheet1 = ss.getSheetByName("Sheet1");
  if (sheet1 && sheet1.getLastRow() === 0 && ss.getSheets().length > 1) {
    ss.deleteSheet(sheet1);
  }

  seedInitialData_();
  Logger.log("Setup selesai. Total sheet: " + ss.getSheets().length);
}

function setEnvironmentConfig(spreadsheetId, folderId, archiveFolderId) {
  if (!spreadsheetId) {
    return { success: false, message: "spreadsheetId wajib diisi" };
  }
  if (!folderId) {
    return { success: false, message: "folderId wajib diisi" };
  }
  if (!archiveFolderId) {
    return { success: false, message: "archiveFolderId wajib diisi" };
  }

  var props = PropertiesService.getScriptProperties();
  props.setProperty(PROP_KEY_SPREADSHEET_ID, spreadsheetId);
  props.setProperty(PROP_KEY_DRIVE_FOLDER_ID, folderId);
  props.setProperty(PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID, archiveFolderId);

  Logger.log("=== ENVIRONMENT CONFIG ===");
  Logger.log("SPREADSHEET_ID: " + spreadsheetId);
  Logger.log("DRIVE_FOLDER_ID: " + folderId);
  Logger.log("DRIVE_ARCHIVE_FOLDER_ID: " + archiveFolderId);

  return {
    success: true,
    spreadsheetId: spreadsheetId,
    folderId: folderId,
    archiveFolderId: archiveFolderId,
  };
}

function setupProd() {
  setEnvironmentConfig(
    "1DatV0OTTpvwJ1dfquA-ywmIeEq3UJQaRMU-1yt-IhYI",
    "15riNBnDt-IIepfVCcgXX7pWBRR68J7Ax",
    "17QoKWtgJsH64q07wwiKHewKuY7r8k8V-",
  );
}

function verifyProd() {
  showEnvironmentConfig();
  validateEnvironmentConfig();
}

function showEnvironmentConfig() {
  var props = PropertiesService.getScriptProperties();
  var config = {
    SPREADSHEET_ID: props.getProperty(PROP_KEY_SPREADSHEET_ID) || "(kosong)",
    DRIVE_FOLDER_ID: props.getProperty(PROP_KEY_DRIVE_FOLDER_ID) || "(kosong)",
    DRIVE_ARCHIVE_FOLDER_ID:
      props.getProperty(PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID) || "(kosong)",
  };

  Logger.log("=== ENVIRONMENT CONFIG ===");
  Object.keys(config).forEach(function (key) {
    Logger.log(key + ": " + config[key]);
  });

  return config;
}

function clearEnvironmentConfig() {
  var props = PropertiesService.getScriptProperties();
  props.deleteProperty(PROP_KEY_SPREADSHEET_ID);
  props.deleteProperty(PROP_KEY_DRIVE_FOLDER_ID);
  props.deleteProperty(PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID);
  Logger.log("Environment config dihapus");
  return { success: true };
}

function validateEnvironmentConfig() {
  var errors = [];

  var spreadsheetId = getSpreadsheetId_();
  if (!spreadsheetId) {
    errors.push("SPREADSHEET_ID belum diatur");
  } else {
    try {
      var ss = SpreadsheetApp.openById(spreadsheetId);
      Logger.log("Spreadsheet OK: " + ss.getName());
    } catch (e) {
      errors.push("SPREADSHEET_ID tidak valid: " + e.message);
    }
  }

  var folderId = getDriveFolderId_();
  if (!folderId) {
    errors.push("DRIVE_FOLDER_ID belum diatur");
  } else {
    try {
      var folder = DriveApp.getFolderById(folderId);
      Logger.log("Photo folder OK: " + folder.getName());
    } catch (e) {
      errors.push("DRIVE_FOLDER_ID tidak valid: " + e.message);
    }
  }

  var archiveFolderId = getDriveArchiveFolderId_();
  if (!archiveFolderId) {
    errors.push("DRIVE_ARCHIVE_FOLDER_ID belum diatur");
  } else {
    try {
      var archiveFolder = DriveApp.getFolderById(archiveFolderId);
      Logger.log("Archive folder OK: " + archiveFolder.getName());
    } catch (e) {
      errors.push("DRIVE_ARCHIVE_FOLDER_ID tidak valid: " + e.message);
    }
  }

  if (errors.length > 0) {
    Logger.log("=== ERROR ===");
    errors.forEach(function (e) {
      Logger.log("  - " + e);
    });
    return { success: false, errors: errors };
  }

  Logger.log("=== SEMUA OK ===");
  return { success: true };
}

function seedInitialData_() {
  var usersRepo = new SheetRepository_("users");
  if (usersRepo.getAll().length === 0) {
    var now = nowIso_();
    usersRepo.insert({
      user_id: "USR001",
      username: "superadmin",
      password_hash: hashPassword_("ganti123"),
      nama: "Super Admin",
      role: ROLES.SUPER_ADMIN,
      member_id: "",
      status_aktif: true,
      created_at: now,
      updated_at: now,
      last_login_at: "",
    });
    Logger.log("User default: superadmin / ganti123");
  }

  var tplRepo = new SheetRepository_("announcement_templates");
  if (tplRepo.getAll().length === 0) {
    var now2 = nowIso_();
    var defaultTemplate =
      "\u25CF\u25C9\u2740 *UNDANGAN SAMBUNG*\n*KELOMPOK* \u2740 \u25C9\u25CF\u2022\u25E6\n\n" +
      "Assalamu'alaikum wr. wb\n\n" +
      "Diberitahukan kepada seluruh jama'ah {{nama_kelompok}}, bahwa :\n\n" +
      "HARI : {{hari}}, {{tanggal}}\n\nJAM : {{jam}}\n\nACARA : {{acara}}\n\n" +
      "MATERI :\n{{materi}}\n\nNB :\n{{catatan}}\n\n" +
      "Alkhamdulillahi jazakumullohu khoiro\nWassalamualaikum wr. wb\n\n" +
      "ttd\n{{penandatangan}}";

    var templates = [
      ["TPL001", "Undangan Sambung Kelompok", "UNDANGAN_SAMBUNG"],
      ["TPL002", "Pengumuman Pengajian", "PENGUMUMAN_PENGAJIAN"],
      ["TPL003", "Pengingat Pengajian", "PENGINGAT_PENGAJIAN"],
      ["TPL004", "Perubahan Jadwal", "PERUBAHAN_JADWAL"],
      ["TPL005", "Pengumuman Umum", "PENGUMUMAN_UMUM"],
    ];
    templates.forEach(function (t) {
      tplRepo.insert({
        template_id: t[0],
        nama_template: t[1],
        kode: t[2],
        isi_template: defaultTemplate,
        status_aktif: true,
        created_at: now2,
        updated_at: now2,
      });
    });
  }

  var settingsRepo = new SheetRepository_("settings");
  if (settingsRepo.getAll().length === 0) {
    settingsRepo.insert({
      key: "jadwal_rutin",
      value: JSON.stringify(["Minggu", "Selasa", "Kamis"]),
      updated_at: nowIso_(),
    });
  }
}

function setupAllTriggers() {
  Logger.log("=== SETUP ALL TRIGGERS ===");
  Logger.log("");

  Logger.log("1. Hapus trigger lama (kalau ada)...");
  var existing = ScriptApp.getProjectTriggers();
  existing.forEach(function (t) {
    if (
      t.getHandlerFunction() === "keepAlive" ||
      t.getHandlerFunction() === "cleanupOldAiQuota_"
    ) {
      ScriptApp.deleteTrigger(t);
      Logger.log("   Hapus: " + t.getHandlerFunction());
    }
  });
  Logger.log("");

  Logger.log("2. Buat keepAlive trigger (setiap 5 menit)...");
  ScriptApp.newTrigger("keepAlive").timeBased().everyMinutes(5).create();
  Logger.log("   ✅ keepAlive dibuat");
  Logger.log("");

  Logger.log("3. Buat cleanupOldAiQuota_ trigger (setiap hari jam 2 pagi)...");
  ScriptApp.newTrigger("cleanupOldAiQuota_")
    .timeBased()
    .atHour(2)
    .everyDays(1)
    .create();
  Logger.log("   ✅ cleanupOldAiQuota_ dibuat");
  Logger.log("");

  Logger.log("=== VERIFIKASI ===");
  listAllTriggers();

  Logger.log("");
  Logger.log("Yang diharapkan: 2 trigger aktif.");
}

function listAllTriggers() {
  var triggers = ScriptApp.getProjectTriggers();
  Logger.log("=== TRIGGERS ===");
  Logger.log("Total: " + triggers.length);
  Logger.log("");
  triggers.forEach(function (t, i) {
    Logger.log(
      i +
        1 +
        ". " +
        t.getHandlerFunction() +
        " | " +
        t.getEventType() +
        " | " +
        t.getTriggerSource(),
    );
  });
  return triggers.length;
}

/**
 * KeepAlive agresif (setiap 1 menit).
 *
 * Berguna kalau cold start GAS sering terjadi.
 * Trade-off:
 * - Consume quota trigger ~24 menit/hari (masih aman untuk akun gratis)
 * - Log lebih banyak
 *
 * Kalau ingin lebih hemat, ganti .everyMinutes(1) menjadi .everyMinutes(5).
 */
function setupAggressiveKeepAlive() {
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "keepAlive") {
      ScriptApp.deleteTrigger(t);
    }
  });

  ScriptApp.newTrigger("keepAlive").timeBased().everyMinutes(1).create();

  Logger.log("✅ KeepAlive: setiap 1 menit");
  listAllTriggers();
}

/**
 * KeepAlive standar (setiap 5 menit).
 * Sama seperti setupKeepAliveTrigger lama, tapi dengan nama lebih jelas.
 */
function setupStandardKeepAlive() {
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "keepAlive") {
      ScriptApp.deleteTrigger(t);
    }
  });

  ScriptApp.newTrigger("keepAlive").timeBased().everyMinutes(5).create();

  Logger.log("✅ KeepAlive: setiap 5 menit");
  listAllTriggers();
}

/* -------------------------------------------------------------------------- */
/*                          WA QUEUE TRIGGER                                  */
/* -------------------------------------------------------------------------- */

/**
 * Setup trigger processWaQueue_ setiap 15 menit.
 * Jalankan SEKALI dari editor GAS.
 */
function setupWaQueueTrigger() {
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "processWaQueue_") {
      ScriptApp.deleteTrigger(t);
    }
  });

  ScriptApp.newTrigger("processWaQueue_")
    .timeBased()
    .everyMinutes(WA_QUEUE_TRIGGER_MINUTES)
    .create();

  Logger.log(
    "✅ Trigger WA Queue dibuat: setiap " + WA_QUEUE_TRIGGER_MINUTES + " menit",
  );
  listAllTriggers();
}

/**
 * Setup trigger cleanup WA queue harian (jam 2 pagi).
 */
function setupWaQueueCleanupTrigger() {
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "cleanupOldWaQueue_") {
      ScriptApp.deleteTrigger(t);
    }
  });

  ScriptApp.newTrigger("cleanupOldWaQueue_")
    .timeBased()
    .atHour(2)
    .everyDays(1)
    .create();

  Logger.log("✅ Trigger cleanup WA queue dibuat: setiap hari jam 2 pagi");
  listAllTriggers();
}

/**
 * Setup semua trigger WA queue sekaligus.
 */
function setupAllWaQueueTriggers() {
  setupWaQueueTrigger();
  setupWaQueueCleanupTrigger();
}
