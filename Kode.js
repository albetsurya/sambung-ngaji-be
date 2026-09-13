// ============================================================
// KEEP-ALIVE: biar Apps Script container tetap warm
// ============================================================
function keepAlive() {
  // Fungsi minimal — hanya untuk trigger container warm
  return ok_({ ping: new Date().toISOString() });
}

function setupKeepAliveTrigger() {
  // Hapus trigger lama
  var triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    if (t.getHandlerFunction() === "keepAlive") {
      ScriptApp.deleteTrigger(t);
    }
  });

  // Buat trigger baru: setiap 5 menit
  ScriptApp.newTrigger("keepAlive").timeBased().everyMinutes(5).create();

  Logger.log("Keep-alive trigger created. Akan ping setiap 5 menit.");
}

function getConfig_(key, fallback) {
  try {
    var props = PropertiesService.getScriptProperties();
    var value = props.getProperty(key);
    if (value !== null && value !== undefined && value !== "") {
      return value;
    }
  } catch (e) {
    Logger.log("Error baca property " + key + ": " + e);
  }
  return fallback || "";
}

function getSpreadsheetId_() {
  return getConfig_(PROP_KEY_SPREADSHEET_ID, SPREADSHEET_ID);
}

function getDriveFolderId_() {
  return getConfig_(PROP_KEY_DRIVE_FOLDER_ID, DRIVE_FOLDER_ID);
}

function getDriveArchiveFolderId_() {
  return getConfig_(PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID, DRIVE_ARCHIVE_FOLDER_ID);
}

function nowIso_() {
  return new Date().toISOString();
}

function ok_(data, message) {
  return {
    success: true,
    data: data === undefined ? null : data,
    message: message || "",
  };
}

function fail_(message, data) {
  return {
    success: false,
    data: data === undefined ? null : data,
    message: message || "Error",
  };
}

function jsonOutput_(obj) {
  return ContentService.createTextOutput(JSON.stringify(obj)).setMimeType(
    ContentService.MimeType.JSON,
  );
}

function newId_(prefix) {
  var t = Utilities.getUuid().replace(/-/g, "").slice(0, 8).toUpperCase();
  return prefix + t;
}

function generateMemberId() {
  return newId_("MBR");
}
function generateAttendanceId() {
  return newId_("ATD");
}
function generateMeetingId() {
  return newId_("MTG");
}
function generateAnnouncementId() {
  return newId_("ANN");
}
function generateMonitoringId() {
  return newId_("MON");
}
function generateGroupId() {
  return newId_("GRP");
}
function generateUserId() {
  return newId_("USR");
}
function generateLogId() {
  return newId_("LOG");
}
function generateUsageId() {
  return newId_("USE");
}
function generateSubmissionId() {
  return newId_("SUB");
}

function hashPassword_(plain) {
  var digest = Utilities.computeDigest(
    Utilities.DigestAlgorithm.SHA_256,
    plain,
    Utilities.Charset.UTF_8,
  );
  return digest
    .map(function (b) {
      return ("0" + (b & 0xff).toString(16)).slice(-2);
    })
    .join("");
}

function verifyPassword_(plain, hash) {
  return hashPassword_(plain) === hash;
}

function normalizePhoneNumber(raw) {
  if (!raw) return "";
  var digits = String(raw).replace(/[^0-9]/g, "");
  if (digits.indexOf("0") === 0) digits = "62" + digits.slice(1);
  if (digits.indexOf("62") !== 0) digits = "62" + digits;
  return digits;
}

function parseIsoParts_(dateStr) {
  if (!dateStr) return null;
  if (dateStr instanceof Date) {
    var y = dateStr.getFullYear();
    var m = dateStr.getMonth() + 1;
    var d = dateStr.getDate();
    if (!y || !m || !d) return null;
    return { year: y, month: m, day: d };
  }
  var iso = String(dateStr).slice(0, 10);
  var parts = iso.split("-");
  var yy = Number(parts[0]);
  var mm = Number(parts[1]);
  var dd = Number(parts[2]);
  if (!yy || !mm || !dd) return null;
  return { year: yy, month: mm, day: dd };
}

function parseDate_(str) {
  var p = parseIsoParts_(str);
  if (!p) return null;
  return new Date(p.year, p.month - 1, p.day);
}

function formatDate(dateStr) {
  var p = parseIsoParts_(dateStr);
  if (!p) return "";
  var pad = function (n) {
    return String(n).length < 2 ? "0" + n : String(n);
  };
  return p.year + "-" + pad(p.month) + "-" + pad(p.day);
}

var HARI_ID = ["Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"];
var BULAN_ID = [
  "Januari",
  "Februari",
  "Maret",
  "April",
  "Mei",
  "Juni",
  "Juli",
  "Agustus",
  "September",
  "Oktober",
  "November",
  "Desember",
];

function getHariFromDate(dateStr) {
  var p = parseIsoParts_(dateStr);
  if (!p) return "";
  var d = new Date(p.year, p.month - 1, p.day);
  return HARI_ID[d.getDay()];
}

function formatDateShort(dateStr) {
  var p = parseIsoParts_(dateStr);
  if (!p) return "";
  return p.day + " " + BULAN_ID[p.month - 1].slice(0, 3) + " " + p.year;
}

function getMemberAge(tanggalLahir) {
  var p = parseIsoParts_(tanggalLahir);
  if (!p) return null;
  var now = new Date();
  var age = now.getFullYear() - p.year;
  var m = now.getMonth() + 1 - p.month;
  if (m < 0 || (m === 0 && now.getDate() < p.day)) age--;
  return age;
}

function getMemberCategory(member) {
  if (!member) return null;

  if (toBool_(member.is_nikah)) {
    var age = getMemberAge(member.tanggal_lahir);
    return age !== null && age >= 60
      ? MEMBER_CATEGORY.MANULA
      : MEMBER_CATEGORY.DEWASA;
  }

  var jenjang = String(member.jenjang_pendidikan || "").toUpperCase();

  if (jenjang === "TK" || jenjang === "SD") return MEMBER_CATEGORY.CABERAWIT;
  if (jenjang === "SMP") return MEMBER_CATEGORY.PRA_REMAJA;
  if (jenjang === "SMA" || jenjang === "SMK") return MEMBER_CATEGORY.REMAJA;

  var age2 = getMemberAge(member.tanggal_lahir);
  if (age2 !== null && age2 >= 60) return MEMBER_CATEGORY.MANULA;
  if (age2 !== null && age2 < 13) return MEMBER_CATEGORY.CABERAWIT;
  return MEMBER_CATEGORY.PRA_NIKAH;
}

function toBool_(v) {
  return v === true || v === "true" || v === "TRUE" || v === 1 || v === "1";
}

var CACHE_TTL_SECONDS = 600;

var CACHEABLE_SHEETS = {
  users: true,
  members: true,
  groups: true,
  meetings: true,
  // attendance: di-cache per-meeting, bukan per-sheet (lihat getAttendanceByMeeting_)
  monitoring: true,
  announcements: true,
  announcement_templates: true,
  settings: true,
};

function SheetRepository_(sheetKey) {
  this.key = sheetKey;
  this.def = SHEETS[sheetKey];
  if (!this.def) throw new Error("Sheet tidak dikenal: " + sheetKey);
}

SheetRepository_.prototype._sheet = function () {
  var ss = getSpreadsheet_();
  var sheet = ss.getSheetByName(this.def.name);
  if (!sheet)
    throw new Error(
      "Sheet belum dibuat: " +
        this.def.name +
        ". Jalankan setupSpreadsheet() dulu.",
    );
  return sheet;
};

SheetRepository_.prototype._rowsToObjects = function (values) {
  var headers = this.def.headers;
  var out = [];
  for (var i = 1; i < values.length; i++) {
    var row = values[i];
    if (row.join("") === "") continue;
    var obj = {};
    for (var c = 0; c < headers.length; c++) {
      obj[headers[c]] = row[c];
    }
    obj._row = i + 1;
    out.push(obj);
  }
  return out;
};

SheetRepository_.prototype.getAll = function () {
  if (CACHEABLE_SHEETS[this.key]) {
    var cached = CacheService.getScriptCache().get("sheet_" + this.key);
    if (cached) {
      try {
        return JSON.parse(cached);
      } catch (e) {}
    }
  }

  var sheet = this._sheet();
  var lastRow = sheet.getLastRow();
  var lastCol = this.def.headers.length;
  if (lastRow < 2) return [];

  var values = sheet.getRange(1, 1, lastRow, lastCol).getValues();
  var objects = this._rowsToObjects(values);

  if (CACHEABLE_SHEETS[this.key]) {
    try {
      var str = JSON.stringify(objects);
      if (str.length < 95000) {
        CacheService.getScriptCache().put(
          "sheet_" + this.key,
          str,
          CACHE_TTL_SECONDS,
        );
      }
    } catch (e) {}
  }
  return objects;
};

SheetRepository_.prototype._invalidateCache = function () {
  if (CACHEABLE_SHEETS[this.key]) {
    CacheService.getScriptCache().remove("sheet_" + this.key);
  }
};

SheetRepository_.prototype.findById = function (idField, idValue) {
  var all = this.getAll();
  for (var i = 0; i < all.length; i++) {
    if (all[i][idField] === idValue) return all[i];
  }
  return null;
};

SheetRepository_.prototype.find = function (filterFn) {
  return this.getAll().filter(filterFn);
};

SheetRepository_.prototype.insert = function (obj) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var row = headers.map(function (h) {
    if (!obj.hasOwnProperty(h)) return "";
    var val = obj[h];
    if (val === null || val === undefined) return "";
    if (typeof val === "number" || typeof val === "boolean") {
      return "'" + String(val);
    }
    return val;
  });
  sheet.appendRow(row);
  this._invalidateCache();
  return obj;
};

SheetRepository_.prototype.insertMany = function (objs) {
  if (!objs || !objs.length) return objs;
  var sheet = this._sheet();
  var headers = this.def.headers;
  var rows = objs.map(function (obj) {
    return headers.map(function (h) {
      if (!obj.hasOwnProperty(h)) return "";
      var val = obj[h];
      if (val === null || val === undefined) return "";
      if (typeof val === "number" || typeof val === "boolean") {
        return "'" + String(val);
      }
      return val;
    });
  });

  var startRow = sheet.getLastRow() + 1;
  sheet.getRange(startRow, 1, rows.length, headers.length).setValues(rows);
  this._invalidateCache();
  return objs;
};

SheetRepository_.prototype.updateById = function (idField, idValue, patch) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return null;
  var idColIndex = headers.indexOf(idField);
  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();

  for (var i = 0; i < idColValues.length; i++) {
    if (idColValues[i][0] === idValue) {
      var rowNum = i + 2;
      var currentRow = sheet
        .getRange(rowNum, 1, 1, headers.length)
        .getValues()[0];
      var current = {};
      headers.forEach(function (h, idx) {
        current[h] = currentRow[idx];
      });
      var updated = Object.assign({}, current, patch);
      var newRow = headers.map(function (h) {
        return updated.hasOwnProperty(h) ? updated[h] : "";
      });
      sheet.getRange(rowNum, 1, 1, headers.length).setValues([newRow]);
      this._invalidateCache();
      return updated;
    }
  }
  return null;
};

SheetRepository_.prototype.findByField = function (field, value) {
  return this.find(function (o) {
    return o[field] === value;
  });
};

SheetRepository_.prototype.deleteById = function (idField, idValue) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return false;
  var idColIndex = headers.indexOf(idField);
  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();

  for (var i = 0; i < idColValues.length; i++) {
    if (idColValues[i][0] === idValue) {
      sheet.deleteRow(i + 2);
      this._invalidateCache();
      return true;
    }
  }
  return false;
};

function buildIndexList_(rows, keyField) {
  var map = {};
  for (var i = 0; i < rows.length; i++) {
    var k = rows[i][keyField];
    if (!map[k]) map[k] = [];
    map[k].push(rows[i]);
  }
  return map;
}

function buildIndexOne_(rows, keyField) {
  var map = {};
  for (var i = 0; i < rows.length; i++) {
    map[rows[i][keyField]] = rows[i];
  }
  return map;
}

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

function getSpreadsheet_() {
  var id = getSpreadsheetId_();
  return id
    ? SpreadsheetApp.openById(id)
    : SpreadsheetApp.getActiveSpreadsheet();
}

function getPhotoFolder_() {
  var id = getDriveFolderId_();
  if (!id) {
    throw new Error("DRIVE_FOLDER_ID belum diatur di Script Properties");
  }
  return DriveApp.getFolderById(id);
}

function getPhotoArchiveFolder_() {
  var id = getDriveArchiveFolderId_();
  if (!id) {
    throw new Error(
      "DRIVE_ARCHIVE_FOLDER_ID belum diatur di Script Properties",
    );
  }
  return DriveApp.getFolderById(id);
}

function login_(params) {
  var username = String(params.username || "").trim();
  var password = String(params.password || "");
  if (!username || !password) return fail_("Username dan password wajib diisi");

  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.findById("username", username);
  if (!user) return fail_("Username atau password salah");
  if (!toBool_(user.status_aktif)) return fail_("Akun tidak aktif");
  if (!verifyPassword_(password, user.password_hash))
    return fail_("Username atau password salah");

  var token = Utilities.getUuid();
  var now = new Date();
  var expires = new Date(now.getTime() + SESSION_TTL_HOURS * 3600 * 1000);
  var sessionsRepo = new SheetRepository_("sessions");
  sessionsRepo.insert({
    token: token,
    user_id: user.user_id,
    created_at: now.toISOString(),
    expires_at: expires.toISOString(),
  });

  usersRepo.updateById("user_id", user.user_id, {
    last_login_at: now.toISOString(),
  });
  writeAuditLog_(user.user_id, "LOGIN", "USER", user.user_id);
  return ok_({ token: token, user: publicUser_(user) });
}

function logout_(ctx) {
  if (!ctx || !ctx.token) return ok_(null);
  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", ctx.token);
  if (session)
    sessionsRepo.updateById("token", ctx.token, { expires_at: nowIso_() });
  if (ctx.user)
    writeAuditLog_(ctx.user.user_id, "LOGOUT", "USER", ctx.user.user_id);
  return ok_(null);
}

function publicUser_(user) {
  return {
    user_id: user.user_id,
    username: user.username,
    nama: user.nama,
    role: user.role,
    member_id: user.member_id || "",
    jenis_kelamin: user.jenis_kelamin || "",
  };
}

function validateSession_(token) {
  if (!token) return null;
  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", token);
  if (!session) return null;
  if (new Date(session.expires_at).getTime() < Date.now()) return null;

  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.findById("user_id", session.user_id);
  if (!user || !toBool_(user.status_aktif)) return null;
  return { token: token, user: user };
}

function checkPermission_(user, action) {
  var perms = ROLE_PERMISSIONS[user.role];
  if (!perms) return false;
  if (perms.indexOf("*") !== -1) return true;
  return perms.indexOf(action) !== -1;
}

function writeAuditLog_(userId, action, targetType, targetId) {
  try {
    var logsRepo = new SheetRepository_("audit_logs");
    logsRepo.insert({
      log_id: generateLogId(),
      user_id: userId || "",
      action: action,
      target_type: targetType || "",
      target_id: targetId || "",
      timestamp: nowIso_(),
    });
  } catch (e) {
    Logger.log("Gagal menulis audit log: " + e);
  }
}

var PUBLIC_ACTIONS = {
  login: true,
  submitPublicRegistration: true,
};

var ACTION_HANDLERS = {
  login: function (ctx, p) {
    return login_(p);
  },
  logout: function (ctx, p) {
    return logout_(ctx);
  },
  validateSession: function (ctx, p) {
    return ok_({ user: publicUser_(ctx.user) });
  },

  aiChat: function (ctx, p) {
    var body = {
      message: p.message,
      history: p.history ? JSON.parse(p.history) : [],
      provider: p.provider,
    };
    return handleAiChat_(body, ctx);
  },

  getMemberUserStatus: function (ctx, p) {
    if (!p.member_id) return fail_("member_id wajib diisi");
    return ok_(getMemberUserStatus_(p.member_id));
  },

  getCurrentProvider: function (ctx, p) {
    return getCurrentProvider();
  },
  setAIProvider: function (ctx, p) {
    return setAIProvider(p.provider);
  },

  getDashboard: getDashboard_,
  getMyDashboard: getMyDashboard_,

  changeMyPassword: changeMyPassword_,
  resetUserPassword: resetUserPassword_,

  getAiUsageStats: getAiUsageStats_,

  getMembers: getMembers_,
  getMembersPaged: getMembersPaged_,
  getPNKBMembers: getPNKBMembers_,
  getPNKBMembersPaged: getPNKBMembersPaged_,
  getAttendanceMembers: getAttendanceMembers_,
  getMemberDetail: getMemberDetail_,
  createMember: createMember_,
  updateMember: updateMember_,
  deactivateMember: deactivateMember_,

  getGroups: getGroups_,
  saveGroup: saveGroup_,

  getMeetings: getMeetings_,
  createMeeting: createMeeting_,
  updateMeeting: updateMeeting_,

  getAttendance: getAttendance_,
  saveAttendance: saveAttendance_,
  bulkSaveAttendance: bulkSaveAttendance_,

  deleteAttendance: deleteAttendance_,
  deleteAttendanceByMeeting: deleteAttendanceByMeeting_, // ← tambah
  deleteAttendanceByMember: deleteAttendanceByMember_,

  getMonitoring: getMonitoring_,
  createMonitoring: createMonitoring_,
  updateMonitoring: updateMonitoring_,

  getAnnouncementTemplates: getAnnouncementTemplates_,
  generateAnnouncement: generateAnnouncement_,
  generateWeeklyAnnouncements: generateWeeklyAnnouncements_,
  createAnnouncement: createAnnouncement_,
  updateAnnouncement: updateAnnouncement_,
  getAnnouncements: getAnnouncements_,
  getAnnouncementRecipientSummary: getAnnouncementRecipientSummary_,

  uploadPhoto: uploadPhoto_,
  deletePhoto: deletePhoto_,

  getUsers: getUsers_,
  getUserDetail: getUserDetail_,
  createUser: createUser_,
  updateUser: updateUser_,
  updateUserRole: updateUserRole_,

  getSettings: getSettings_,
  updateSettings: updateSettings_,

  getAuditLogs: getAuditLogs_,

  getMyProfile: getMyProfile_,
  updateMyProfile: updateMyProfile_,
  getMyAttendance: getMyAttendance_,
  getMyMonitoring: getMyMonitoring_,
  getUpcomingMeetings: getUpcomingMeetings_,

  submitPublicRegistration: submitPublicRegistration_,
  getPendingMembers: getPendingMembers_,
  getPendingMemberDetail: getPendingMemberDetail_,
  approvePendingMember: approvePendingMember_,
  rejectPendingMember: rejectPendingMember_,
};

var SUPER_ADMIN_ONLY_ACTIONS = {
  getUsers: true,
  getUserDetail: true,
  createUser: true,
  updateUser: true,
  updateUserRole: true,
  getAuditLogs: true,
  updateSettings: true,
  resetUserPassword: true,
};

function doGet(e) {
  return handleRequest_(e);
}
function doPost(e) {
  return handleRequest_(e);
}

function handleRequest_(e) {
  var params = parseParams_(e);
  var action = params.action;
  if (!action) return jsonOutput_(fail_("Parameter action wajib diisi"));

  var handler = ACTION_HANDLERS[action];
  if (!handler) return jsonOutput_(fail_("Action tidak dikenal: " + action));

  // Auth check DULU (sebelum lock)
  var ctx = {};
  if (!PUBLIC_ACTIONS[action]) {
    ctx = validateSession_(params.token);
    if (!ctx)
      return jsonOutput_(
        fail_("Unauthorized: sesi tidak valid atau kadaluarsa"),
      );

    if (
      SUPER_ADMIN_ONLY_ACTIONS[action] &&
      ctx.user.role !== ROLES.SUPER_ADMIN
    ) {
      return jsonOutput_(fail_("Forbidden: hanya SUPER_ADMIN"));
    }
    if (!checkPermission_(ctx.user, action)) {
      return jsonOutput_(
        fail_(
          "Forbidden: role " +
            ctx.user.role +
            " tidak memiliki akses ke " +
            action,
        ),
      );
    }
  }

  // Lock HANYA untuk write actions
  var lock = null;
  var hasLock = true;

  if (shouldUseLock_(action)) {
    lock = LockService.getScriptLock();
    hasLock = lock.tryLock(30000);
    if (!hasLock) {
      return jsonOutput_(fail_("Server sedang sibuk. Coba lagi."));
    }
  }

  try {
    var result = handler(ctx, params);
    return jsonOutput_(result);
  } catch (err) {
    Logger.log("handleRequest_ error: " + err + "\n" + (err && err.stack));
    return jsonOutput_(fail_("Terjadi kesalahan server: " + err));
  } finally {
    if (lock && hasLock) {
      try {
        lock.releaseLock();
      } catch (e) {
        /* ignore */
      }
    }
  }
}

function parseParams_(e) {
  var params = {};

  // 1. Query params
  if (e && e.parameter) {
    Object.keys(e.parameter).forEach(function (k) {
      params[k] = e.parameter[k];
    });
  }

  // 2. POST body
  if (e && e.postData && e.postData.contents) {
    var contents = e.postData.contents;
    var type = String(e.postData.type || "").toLowerCase();

    // 2a. JSON
    if (contents.charAt(0) === "{") {
      try {
        var body = JSON.parse(contents);
        Object.keys(body).forEach(function (k) {
          params[k] = body[k];
        });
      } catch (err) {
        Logger.log("parseParams_ JSON error: " + err);
      }
    }
    // 2b. urlencoded
    else if (type.indexOf("application/x-www-form-urlencoded") === 0) {
      try {
        contents.split("&").forEach(function (pair) {
          var idx = pair.indexOf("=");
          if (idx > 0) {
            var key = decodeURIComponent(
              pair.slice(0, idx).replace(/\+/g, " "),
            );
            var val = decodeURIComponent(
              pair.slice(idx + 1).replace(/\+/g, " "),
            );
            params[key] = val;
          }
        });
      } catch (err) {
        Logger.log("parseParams_ urlencoded error: " + err);
      }
    }
  }

  // 3. ⭐ Fallback: kalau action masih kosong, coba dari e.parameter.action
  //    (kadang Apps Script lempar JSON ke parameter kalau body tidak terdeteksi)
  if (!params.action && e && e.parameter) {
    // Cek apakah ada parameter aneh yang isinya JSON
    Object.keys(e.parameter).forEach(function (k) {
      var v = String(e.parameter[k]);
      if (v.charAt(0) === "{") {
        try {
          var parsed = JSON.parse(v);
          Object.keys(parsed).forEach(function (pk) {
            if (!params[pk]) params[pk] = parsed[pk];
          });
        } catch (e) {}
      }
    });
  }

  return params;
}

function getDashboard_(ctx, params) {
  var role = ctx.user.role;
  if (role === ROLES.TIM_PNKB) return getPNKBDashboard_(ctx);
  if (role === ROLES.TIM_ABSENSI) return getAbsensiDashboard_(ctx);
  return getGeneralDashboard_(ctx);
}

function getMyDashboard_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  if (!member) return fail_("Data jamaah tidak ditemukan");

  var enriched = enrichMember_(member);

  var profileFields = [
    "member_id",
    "nama_lengkap",
    "nama_panggilan",
    "jenis_kelamin",
    "tempat_lahir",
    "tanggal_lahir",
    "foto_url",
    "kelompok",
    "desa",
    "daerah",
    "alamat_rumah",
    "no_wa",
    "pekerjaan",
    "hobi",
    "status_pembinaan",
    "status_aktif",
    "tanggal_masuk",
    "jenjang_pendidikan",
    "sekolah",
    "jurusan",
    "tahun_mulai_pendidikan",
    "tahun_selesai_pendidikan",
    "kategori",
    "usia",
  ];
  var profile = {};
  profileFields.forEach(function (f) {
    if (enriched.hasOwnProperty(f)) profile[f] = enriched[f];
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var attendanceRows = attendanceRepo.findByField(
    "member_id",
    ctx.user.member_id,
  );
  var meetingsRepo = new SheetRepository_("meetings");
  var allMeetings = meetingsRepo.getAll();
  var meetingsById = buildIndexOne_(meetingsRepo.getAll(), "meeting_id");

  var attendance = attendanceRows.map(function (a) {
    var m = meetingsById[a.meeting_id] || {};
    return {
      attendance_id: a.attendance_id,
      meeting_id: a.meeting_id,
      status: a.status,
      catatan: a.catatan || "",
      tanggal: m.tanggal || "",
      hari: m.hari || "",
      acara: m.acara || "",
      jam: m.jam || "",
      created_at: a.created_at,
    };
  });
  attendance.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  var monitoringRepo = new SheetRepository_("monitoring");
  var monitoring = monitoringRepo.findByField("member_id", ctx.user.member_id);
  monitoring.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  monitoring = monitoring.map(function (r) {
    var c = Object.assign({}, r);
    delete c._row;
    return c;
  });

  var today = formatDate(nowIso_());
  var memberCategory = getMemberCategory(member);
  var upcoming = allMeetings
    .filter(function (m) {
      if (formatDate(m.tanggal) < today) return false;
      var targets = parseKategoriTarget_(m.kategori_target);
      if (targets.length === 0) return true;
      return memberCategory && targets.indexOf(memberCategory) !== -1;
    })
    .sort(function (a, b) {
      return new Date(a.tanggal) - new Date(b.tanggal);
    })
    .slice(0, 5)
    .map(publicMeeting_);

  return ok_({
    profile: profile,
    attendance: attendance,
    monitoring: monitoring,
    upcoming: upcoming,
  });
}

function getGeneralDashboard_(ctx) {
  var membersRepo = new SheetRepository_("members");
  var members = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var perKategori = {};
  Object.keys(MEMBER_CATEGORY).forEach(function (k) {
    perKategori[k] = 0;
  });
  members.forEach(function (m) {
    var cat = getMemberCategory(m);
    if (cat) perKategori[cat] = (perKategori[cat] || 0) + 1;
  });

  var meetingsRepo = new SheetRepository_("meetings");
  var meetings = meetingsRepo.getAll();
  var today = formatDate(nowIso_());
  var upcoming = meetings
    .filter(function (m) {
      return formatDate(m.tanggal) >= today;
    })
    .sort(function (a, b) {
      return new Date(a.tanggal) - new Date(b.tanggal);
    })[0];

  // OPTIMASI: baca hanya N row terakhir, bukan 36.000
  var attendanceRepo = new SheetRepository_("attendance");
  var recentAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var meetingsById = buildIndexOne_(meetings, "meeting_id");
  var attendanceByMember = buildIndexList_(recentAttendance, "member_id");

  var result = {
    total_jamaah: members.length,
    total_jamaah_aktif: members.length,
    per_kategori: perKategori,
    rata_rata_kehadiran: computeOverallAttendanceRate_(recentAttendance),
    pengajian_terdekat: upcoming
      ? {
          meeting_id: upcoming.meeting_id,
          tanggal: formatDate(upcoming.tanggal),
          hari: upcoming.hari,
          acara: upcoming.acara,
          kategori_target: parseKategoriTarget_(upcoming.kategori_target),
        }
      : null,
    jamaah_perlu_perhatian: buildAttentionList_(
      members,
      attendanceByMember,
      meetingsById,
    ).slice(0, 10),
    data_belum_lengkap: countIncompleteData_(members),
  };

  if (ctx.user.role === ROLES.SUPER_ADMIN || ctx.user.role === ROLES.ADMIN) {
    var usersRepo = new SheetRepository_("users");
    result.user_aktif = usersRepo.find(function (u) {
      return toBool_(u.status_aktif);
    }).length;
  }
  return ok_(result);
}

function getPNKBDashboard_(ctx) {
  var membersRepo = new SheetRepository_("members");
  var allActive = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var pnkb = allActive.filter(function (m) {
    return getMemberCategory(m) === MEMBER_CATEGORY.PRA_NIKAH;
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var meetingsRepo = new SheetRepository_("meetings");
  var allAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var meetingsById = buildIndexOne_(meetingsRepo.getAll(), "meeting_id");
  var attendanceByMember = buildIndexList_(allAttendance, "member_id");

  var attention = buildAttentionList_(pnkb, attendanceByMember, meetingsById);

  var pnkbIds = {};
  pnkb.forEach(function (m) {
    pnkbIds[m.member_id] = true;
  });
  var pnkbAttendance = allAttendance.filter(function (a) {
    return pnkbIds[a.member_id];
  });
  var hadir = pnkbAttendance.filter(function (a) {
    return a.status === ATTENDANCE_STATUS.HADIR;
  }).length;
  var rate = pnkbAttendance.length
    ? Math.round((hadir / pnkbAttendance.length) * 100)
    : 0;

  return ok_({
    total: pnkb.length,
    aktif: pnkb.filter(function (m) {
      return m.status_pembinaan === MONITORING_STATUS.AKTIF;
    }).length,
    perlu_perhatian: attention.length,
    kehadiran: rate,
    data_belum_lengkap: countIncompleteData_(pnkb),
  });
}

function getAbsensiDashboard_(ctx, params) {
  var meetingsRepo = new SheetRepository_("meetings");
  var today = formatDate(nowIso_());
  var todayMeetings = meetingsRepo.find(function (m) {
    return formatDate(m.tanggal) === today;
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var allAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var attendanceByMeeting = buildIndexList_(allAttendance, "meeting_id");

  var membersRepo = new SheetRepository_("members");
  var allMembers = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var summary = { HADIR: 0, IJIN: 0, SAKIT: 0, TANPA_KETERANGAN: 0 };
  var totalJamaah = 0;
  var totalTarget = 0;
  var byCategory = {};

  todayMeetings.forEach(function (meeting) {
    var targets = parseKategoriTarget_(meeting.kategori_target);

    var eligibleMembers = allMembers;
    if (targets.length > 0) {
      eligibleMembers = allMembers.filter(function (m) {
        var cat = getMemberCategory(m);
        return cat && targets.indexOf(cat) !== -1;
      });
    }
    totalTarget += eligibleMembers.length;

    var rows = attendanceByMeeting[meeting.meeting_id] || [];
    totalJamaah += rows.length;
    rows.forEach(function (r) {
      if (summary.hasOwnProperty(r.status)) summary[r.status]++;
    });

    var key = targets.length === 0 ? "SEMUA" : targets.slice().sort().join("|");
    if (!byCategory[key]) {
      byCategory[key] = {
        kategori: targets,
        is_semua: targets.length === 0,
        hadir: 0,
        ijin: 0,
        sakit: 0,
        tanpa_keterangan: 0,
        total_absen: 0,
        total_target: 0,
        meeting_count: 0,
      };
    }
    byCategory[key].meeting_count++;
    byCategory[key].total_target += eligibleMembers.length;
    byCategory[key].total_absen += rows.length;
    rows.forEach(function (r) {
      if (r.status === ATTENDANCE_STATUS.HADIR) byCategory[key].hadir++;
      else if (r.status === ATTENDANCE_STATUS.IJIN) byCategory[key].ijin++;
      else if (r.status === ATTENDANCE_STATUS.SAKIT) byCategory[key].sakit++;
      else if (r.status === ATTENDANCE_STATUS.TANPA_KETERANGAN)
        byCategory[key].tanpa_keterangan++;
    });
  });

  var byCategoryList = Object.keys(byCategory).map(function (key) {
    var item = byCategory[key];
    item.persentase = item.total_target
      ? Math.round((item.hadir / item.total_target) * 100)
      : 0;
    return item;
  });

  byCategoryList.sort(function (a, b) {
    if (a.is_semua !== b.is_semua) return a.is_semua ? 1 : -1;
    return b.total_target - a.total_target;
  });

  return ok_({
    pengajian_hari_ini: todayMeetings.map(function (m) {
      return {
        meeting_id: m.meeting_id,
        tanggal: formatDate(m.tanggal),
        jam: m.jam,
        acara: m.acara,
        group_id: m.group_id,
        kategori_target: parseKategoriTarget_(m.kategori_target),
      };
    }),
    jumlah_jamaah: totalJamaah,
    total_target: totalTarget,
    hadir: summary.HADIR,
    ijin: summary.IJIN,
    sakit: summary.SAKIT,
    tanpa_keterangan: summary.TANPA_KETERANGAN,
    by_category: byCategoryList,
  });
}

function parseKategoriTarget_(raw) {
  if (!raw) return [];
  try {
    var parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
    return Array.isArray(parsed) ? parsed : [];
  } catch (e) {
    return [];
  }
}

function computeOverallAttendanceRate_(allAttendance) {
  if (!allAttendance || !allAttendance.length) return 0;
  var hadir = allAttendance.filter(function (a) {
    return a.status === ATTENDANCE_STATUS.HADIR;
  }).length;
  return Math.round((hadir / allAttendance.length) * 100);
}

function buildAttentionList_(members, attendanceByMember, meetingsById) {
  var sixMonthsAgo = new Date();
  sixMonthsAgo.setMonth(sixMonthsAgo.getMonth() - 6);

  var list = [];
  members.forEach(function (m) {
    var rows = (attendanceByMember[m.member_id] || []).slice();
    rows.sort(function (a, b) {
      var ma = meetingsById[a.meeting_id];
      var mb = meetingsById[b.meeting_id];
      return new Date(mb ? mb.tanggal : 0) - new Date(ma ? ma.tanggal : 0);
    });

    var reasons = [];
    if (rows.length >= 3) {
      var hadirCount = rows.filter(function (r) {
        return r.status === ATTENDANCE_STATUS.HADIR;
      }).length;
      var rate = Math.round((hadirCount / rows.length) * 100);
      if (rate < 60) reasons.push("Kehadiran " + rate + "%");

      var last3 = rows.slice(0, 3);
      if (
        last3.length === 3 &&
        last3.every(function (r) {
          return r.status !== ATTENDANCE_STATUS.HADIR;
        })
      ) {
        reasons.push("3x berturut-turut tidak hadir");
      }
    }
    if (!m.updated_at || new Date(m.updated_at) < sixMonthsAgo) {
      reasons.push("Data belum diperbarui > 6 bulan");
    }
    if (
      m.status_pembinaan === MONITORING_STATUS.PERLU_PERHATIAN ||
      m.status_pembinaan === MONITORING_STATUS.TIDAK_AKTIF
    ) {
      reasons.push("Status pembinaan: " + m.status_pembinaan);
    }

    if (reasons.length) {
      list.push({
        member_id: m.member_id,
        nama_lengkap: m.nama_lengkap,
        foto_url: m.foto_url,
        reasons: reasons,
      });
    }
  });
  return list;
}

function countIncompleteData_(members) {
  var requiredFields = ["tanggal_lahir", "kelompok", "desa", "alamat_rumah"];
  return members.filter(function (m) {
    return requiredFields.some(function (f) {
      return !m[f];
    });
  }).length;
}

function enrichMember_(member) {
  var out = Object.assign({}, member);
  out.kategori = getMemberCategory(member);
  out.usia = getMemberAge(member.tanggal_lahir);
  delete out._row;
  return out;
}

function filterMemberFieldsByRole_(member, role) {
  var allowed = FIELD_VISIBILITY.UMUM.concat(["kategori", "usia", "no_wa"]);
  if (role === "SUPER_ADMIN" || role === "ADMIN") {
    allowed = allowed.concat(
      FIELD_VISIBILITY.INTERNAL,
      FIELD_VISIBILITY.PNKB,
      FIELD_VISIBILITY.PENDIDIKAN,
    );
  } else if (role === "PENGAWAS") {
    allowed = allowed.concat(
      FIELD_VISIBILITY.INTERNAL,
      FIELD_VISIBILITY.PNKB,
      FIELD_VISIBILITY.PENDIDIKAN,
    );
  } else if (role === "TIM_PNKB") {
    allowed = allowed.concat([
      "pekerjaan",
      "is_muballigh",
      "is_kerja",
      "is_nikah",
      "tinggi_badan",
      "berat_badan",
      "hobi",
      "status_pembinaan",
      "alamat_rumah",
    ]);
  } else if (role === "TIM_ABSENSI") {
    allowed = allowed.concat(["kelompok"]);
  }
  var out = {};
  allowed.forEach(function (f) {
    if (member.hasOwnProperty(f)) out[f] = member[f];
  });
  return out;
}

function getMembers_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();

  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (m) {
      return toBool_(m.status_aktif);
    });
  }
  if (params.kelompok)
    all = all.filter(function (m) {
      return m.kelompok === params.kelompok;
    });
  if (params.jenis_kelamin)
    all = all.filter(function (m) {
      return m.jenis_kelamin === params.jenis_kelamin;
    });
  if (params.desa)
    all = all.filter(function (m) {
      return m.desa === params.desa;
    });
  if (params.search) {
    var q = String(params.search).toLowerCase();
    all = all.filter(function (m) {
      return (
        String(m.nama_lengkap).toLowerCase().indexOf(q) !== -1 ||
        String(m.nama_panggilan).toLowerCase().indexOf(q) !== -1
      );
    });
  }

  var enriched = all.map(enrichMember_);
  if (params.kategori) {
    enriched = enriched.filter(function (m) {
      return m.kategori === params.kategori;
    });
  }

  var role = ctx.user.role;
  var filtered = enriched.map(function (m) {
    return filterMemberFieldsByRole_(m, role);
  });
  return ok_(filtered);
}

function getMembersPaged_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();

  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (m) {
      return toBool_(m.status_aktif);
    });
  }
  if (params.kelompok)
    all = all.filter(function (m) {
      return m.kelompok === params.kelompok;
    });
  if (params.jenis_kelamin)
    all = all.filter(function (m) {
      return m.jenis_kelamin === params.jenis_kelamin;
    });
  if (params.desa)
    all = all.filter(function (m) {
      return m.desa === params.desa;
    });
  if (params.search) {
    var q = String(params.search).toLowerCase();
    all = all.filter(function (m) {
      return (
        String(m.nama_lengkap).toLowerCase().indexOf(q) !== -1 ||
        String(m.nama_panggilan).toLowerCase().indexOf(q) !== -1
      );
    });
  }

  var enriched = all.map(enrichMember_);
  if (params.kategori) {
    enriched = enriched.filter(function (m) {
      return m.kategori === params.kategori;
    });
  }

  enriched.sort(function (a, b) {
    return String(a.nama_lengkap).localeCompare(String(b.nama_lengkap));
  });

  var total = enriched.length;
  var limit = params.limit ? Number(params.limit) : 0;
  var offset = params.offset ? Number(params.offset) : 0;

  var paged = enriched;
  if (limit > 0) {
    paged = enriched.slice(offset, offset + limit);
  }

  var role = ctx.user.role;
  var filtered = paged.map(function (m) {
    return filterMemberFieldsByRole_(m, role);
  });

  return ok_({
    items: filtered,
    total: total,
    limit: limit,
    offset: offset,
    has_more: limit > 0 ? offset + paged.length < total : false,
  });
}

function getPNKBMembers_(ctx, params) {
  return getMembers_(
    ctx,
    Object.assign({}, params, { kategori: MEMBER_CATEGORY.PRA_NIKAH }),
  );
}

function getPNKBMembersPaged_(ctx, params) {
  return getMembersPaged_(
    ctx,
    Object.assign({}, params, { kategori: MEMBER_CATEGORY.PRA_NIKAH }),
  );
}

function getAttendanceMembers_(ctx, params) {
  return getMembers_(ctx, params);
}

function getMemberDetail_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", memberId);
  if (!member) return fail_("Jamaah tidak ditemukan");

  if (ctx.user.role === ROLES.TIM_PNKB) {
    var cat = getMemberCategory(member);
    if (cat !== MEMBER_CATEGORY.PRA_NIKAH)
      return fail_("Tidak memiliki akses ke data jamaah ini");
  }

  var enriched = enrichMember_(member);
  var filtered = filterMemberFieldsByRole_(enriched, ctx.user.role);
  return ok_(filtered);
}

function createMember_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var now = nowIso_();
  var memberId = generateMemberId();
  var member = {
    member_id: memberId,
    nama_lengkap: params.nama_lengkap || "",
    nama_panggilan: params.nama_panggilan || "",
    jenis_kelamin: params.jenis_kelamin || "",
    tempat_lahir: params.tempat_lahir || "",
    tanggal_lahir: params.tanggal_lahir || "",
    kelompok: params.kelompok || "",
    desa: params.desa || "",
    daerah: params.daerah || "",
    alamat_rumah: params.alamat_rumah || "",
    no_wa: params.no_wa ? normalizePhoneNumber(params.no_wa) : "",
    is_muballigh: toBool_(params.is_muballigh),
    is_kerja: toBool_(params.is_kerja),
    is_nikah: toBool_(params.is_nikah),
    tinggi_badan: params.tinggi_badan || "",
    berat_badan: params.berat_badan || "",
    hobi: params.hobi || "",
    pekerjaan: params.pekerjaan || "",
    foto_url: params.foto_url || "",
    status_pembinaan: params.status_pembinaan || MONITORING_STATUS.AKTIF,
    status_aktif: true,
    tanggal_masuk: params.tanggal_masuk || formatDate(now),
    tanggal_keluar: "",
    jenjang_pendidikan: params.jenjang_pendidikan || "",
    sekolah: params.sekolah || "",
    jurusan: params.jurusan || "",
    tahun_mulai_pendidikan: params.tahun_mulai_pendidikan || "",
    tahun_selesai_pendidikan: params.tahun_selesai_pendidikan || "",
    created_at: now,
    updated_at: now,
  };
  membersRepo.insert(member);
  writeAuditLog_(ctx.user.user_id, "CREATE_MEMBER", "MEMBER", memberId);
  return ok_(member);
}

var MEMBER_EDITABLE_FIELDS = [
  "nama_lengkap",
  "nama_panggilan",
  "jenis_kelamin",
  "tempat_lahir",
  "tanggal_lahir",
  "kelompok",
  "desa",
  "daerah",
  "alamat_rumah",
  "no_wa",
  "is_muballigh",
  "is_kerja",
  "is_nikah",
  "tinggi_badan",
  "berat_badan",
  "hobi",
  "pekerjaan",
  "foto_url",
  "status_pembinaan",
  "tanggal_masuk",
  "tanggal_keluar",
  "jenjang_pendidikan",
  "sekolah",
  "jurusan",
  "tahun_mulai_pendidikan",
  "tahun_selesai_pendidikan",
];

function updateMember_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");
  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", memberId);
  if (!existing) return fail_("Jamaah tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  MEMBER_EDITABLE_FIELDS.forEach(function (f) {
    if (params.hasOwnProperty(f)) {
      patch[f] =
        f === "no_wa" && params[f]
          ? normalizePhoneNumber(params[f])
          : params[f];
    }
  });
  var updated = membersRepo.updateById("member_id", memberId, patch);
  writeAuditLog_(ctx.user.user_id, "UPDATE_MEMBER", "MEMBER", memberId);
  return ok_(updated);
}

function deactivateMember_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");
  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", memberId);
  if (!existing) return fail_("Jamaah tidak ditemukan");
  var updated = membersRepo.updateById("member_id", memberId, {
    status_aktif: false,
    tanggal_keluar: formatDate(nowIso_()),
    updated_at: nowIso_(),
  });
  writeAuditLog_(ctx.user.user_id, "DEACTIVATE_MEMBER", "MEMBER", memberId);
  return ok_(updated);
}

function requireMemberLink_(ctx) {
  if (!ctx.user.member_id) {
    return fail_("Akun Anda belum terhubung ke data jamaah. Hubungi admin.");
  }
  return null;
}

var MEMBER_SELF_EDITABLE_FIELDS = [
  "nama_panggilan",
  "no_wa",
  "alamat_rumah",
  "desa",
  "daerah",
  "pekerjaan",
  "hobi",
  "foto_url",
  "tinggi_badan",
  "berat_badan",
  "is_kerja",
  "is_nikah",
  "jenjang_pendidikan",
  "sekolah",
  "jurusan",
  "tahun_mulai_pendidikan",
  "tahun_selesai_pendidikan",
];

function getMyProfile_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  if (!member) return fail_("Data jamaah tidak ditemukan");

  var enriched = enrichMember_(member);
  var allowed = [
    "member_id",
    "nama_lengkap",
    "nama_panggilan",
    "jenis_kelamin",
    "tempat_lahir",
    "tanggal_lahir",
    "foto_url",
    "kelompok",
    "desa",
    "daerah",
    "alamat_rumah",
    "no_wa",
    "pekerjaan",
    "hobi",
    "status_pembinaan",
    "status_aktif",
    "tanggal_masuk",
    "jenjang_pendidikan",
    "sekolah",
    "jurusan",
    "tahun_mulai_pendidikan",
    "tahun_selesai_pendidikan",
    "kategori",
    "usia",
  ];
  var out = {};
  allowed.forEach(function (f) {
    if (enriched.hasOwnProperty(f)) out[f] = enriched[f];
  });
  return ok_(out);
}

function updateMyProfile_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", ctx.user.member_id);
  if (!existing) return fail_("Data jamaah tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  MEMBER_SELF_EDITABLE_FIELDS.forEach(function (f) {
    if (params.hasOwnProperty(f)) {
      if (f === "no_wa" && params[f]) {
        patch[f] = normalizePhoneNumber(params[f]);
      } else {
        patch[f] = params[f];
      }
    }
  });

  var changedFields = Object.keys(patch).filter(function (k) {
    return k !== "updated_at";
  });
  if (changedFields.length === 0) {
    return fail_("Tidak ada perubahan.");
  }

  var updated = membersRepo.updateById("member_id", ctx.user.member_id, patch);

  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MY_PROFILE",
    "MEMBER",
    ctx.user.member_id,
  );

  return ok_(enrichMember_(updated));
}

function getMyAttendance_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var repo = new SheetRepository_("attendance");
  var rows = repo.findByField("member_id", ctx.user.member_id);

  var meetingsRepo = new SheetRepository_("meetings");
  var meetingsById = buildIndexOne_(meetingsRepo.getAll(), "meeting_id");

  var enriched = rows.map(function (a) {
    var m = meetingsById[a.meeting_id] || {};
    return {
      attendance_id: a.attendance_id,
      meeting_id: a.meeting_id,
      status: a.status,
      catatan: a.catatan || "",
      tanggal: m.tanggal || "",
      hari: m.hari || "",
      acara: m.acara || "",
      jam: m.jam || "",
      created_at: a.created_at,
    };
  });

  enriched.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  return ok_(enriched);
}

function getMyMonitoring_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var repo = new SheetRepository_("monitoring");
  var rows = repo.findByField("member_id", ctx.user.member_id);

  rows.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  return ok_(
    rows.map(function (r) {
      var c = Object.assign({}, r);
      delete c._row;
      return c;
    }),
  );
}

function getUpcomingMeetings_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  var memberCategory = member ? getMemberCategory(member) : null;

  var meetingsRepo = new SheetRepository_("meetings");
  var today = formatDate(nowIso_());
  var limit = params.limit ? Number(params.limit) : 10;

  var upcoming = meetingsRepo
    .getAll()
    .filter(function (m) {
      if (formatDate(m.tanggal) < today) return false;
      var targets = parseKategoriTarget_(m.kategori_target);
      if (targets.length === 0) return true;
      return memberCategory && targets.indexOf(memberCategory) !== -1;
    })
    .sort(function (a, b) {
      return parseDate_(a.tanggal) - parseDate_(b.tanggal);
    })
    .slice(0, limit);

  return ok_(upcoming.map(publicMeeting_));
}

var MAX_SUBMISSION_PER_IP_PER_DAY = 3;
var REGISTRATION_MIN_AGE = 0;
var REGISTRATION_MAX_AGE = 120;

function submitPublicRegistration_(ctx, params) {
  var namaLengkap = String(params.nama_lengkap || "").trim();
  var jenisKelamin = String(params.jenis_kelamin || "")
    .trim()
    .toUpperCase();
  var noWa = String(params.no_wa || "").trim();

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

  var clientIp = String(params._client_ip || "unknown").substring(0, 50);

  SpreadsheetApp.flush();

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
  for (var j = 0; j < allPending.length; j++) {
    var p = allPending[j];
    var status = String(p.status).trim().toUpperCase();
    var wa = String(p.no_wa).trim();
    if (wa === normalizedWa && status === PENDING_STATUS.PENDING) {
      existingPending = p;
      break;
    }
  }
  if (existingPending) {
    return fail_("Pendaftaran dengan nomor ini sedang menunggu verifikasi");
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
    status: PENDING_STATUS.PENDING,
    submitted_at: now,
    submitted_ip: clientIp,
    reviewed_by: "",
    reviewed_at: "",
    rejection_reason: "",
    created_member_id: "",
  };

  pendingRepo.insert(row);

  SpreadsheetApp.flush();

  return ok_({
    submission_id: submissionId,
    nama_lengkap: namaLengkap,
    submitted_at: now,
  });
}

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

  return ok_(
    all.map(function (p) {
      var c = Object.assign({}, p);
      delete c._row;
      return c;
    }),
  );
}

function getPendingMemberDetail_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var repo = new SheetRepository_("pending_members");
  var pending = repo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");

  var c = Object.assign({}, pending);
  delete c._row;
  return ok_(c);
}

function approvePendingMember_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var pendingRepo = new SheetRepository_("pending_members");
  var pending = pendingRepo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");
  if (pending.status !== PENDING_STATUS.PENDING) {
    return fail_("Pendaftaran sudah diproses");
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

  pendingRepo.updateById("submission_id", params.submission_id, {
    status: PENDING_STATUS.APPROVED,
    reviewed_by: ctx.user.user_id,
    reviewed_at: now,
    created_member_id: memberId,
  });

  var createUser = toBool_(params.create_user);
  var createdUser = null;
  if (createUser && pending.no_wa) {
    var username = String(params.username || pending.no_wa).trim();
    var password = String(params.password || "").trim();

    if (username && password.length >= 6) {
      var usersRepo = new SheetRepository_("users");
      var existingUser = usersRepo.findById("username", username);
      if (!existingUser) {
        var userId = generateUserId();
        usersRepo.insert({
          user_id: userId,
          username: username,
          password_hash: hashPassword_(password),
          nama: pending.nama_lengkap,
          role: ROLES.MEMBER,
          member_id: memberId,
          status_aktif: true,
          created_at: now,
          updated_at: now,
          last_login_at: "",
        });
        createdUser = { user_id: userId, username: username };
        writeAuditLog_(
          ctx.user.user_id,
          "CREATE_USER_FROM_PENDING",
          "USER",
          userId,
        );
      }
    }
  }

  writeAuditLog_(
    ctx.user.user_id,
    "APPROVE_PENDING",
    "PENDING",
    params.submission_id,
  );

  return ok_({
    member_id: memberId,
    submission_id: params.submission_id,
    created_user: createdUser,
  });
}

function rejectPendingMember_(ctx, params) {
  if (!params.submission_id) return fail_("submission_id wajib diisi");

  var pendingRepo = new SheetRepository_("pending_members");
  var pending = pendingRepo.findById("submission_id", params.submission_id);
  if (!pending) return fail_("Pendaftaran tidak ditemukan");
  if (pending.status !== PENDING_STATUS.PENDING) {
    return fail_("Pendaftaran sudah diproses");
  }

  var now = nowIso_();
  var reason = String(params.reason || "").trim();

  pendingRepo.updateById("submission_id", params.submission_id, {
    status: PENDING_STATUS.REJECTED,
    reviewed_by: ctx.user.user_id,
    reviewed_at: now,
    rejection_reason: reason,
  });

  writeAuditLog_(
    ctx.user.user_id,
    "REJECT_PENDING",
    "PENDING",
    params.submission_id,
  );

  return ok_({
    submission_id: params.submission_id,
    status: PENDING_STATUS.REJECTED,
  });
}

function getGroups_(ctx, params) {
  var repo = new SheetRepository_("groups");
  var all = repo.getAll();
  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (g) {
      return toBool_(g.status_aktif);
    });
  }
  return ok_(
    all.map(function (g) {
      var c = Object.assign({}, g);
      delete c._row;
      return c;
    }),
  );
}

function saveGroup_(ctx, params) {
  var repo = new SheetRepository_("groups");
  var now = nowIso_();

  if (params.group_id) {
    var existing = repo.findById("group_id", params.group_id);
    if (!existing) return fail_("Kelompok tidak ditemukan");
    var patch = { updated_at: now };
    [
      "group_code",
      "group_name",
      "pembina",
      "penandatangan",
      "jadwal",
      "status_aktif",
    ].forEach(function (f) {
      if (params.hasOwnProperty(f)) patch[f] = params[f];
    });
    var updated = repo.updateById("group_id", params.group_id, patch);
    writeAuditLog_(ctx.user.user_id, "UPDATE_GROUP", "GROUP", params.group_id);
    return ok_(updated);
  }

  var groupId = generateGroupId();
  var row = {
    group_id: groupId,
    group_code: params.group_code || "",
    group_name: params.group_name || "",
    pembina: params.pembina || "",
    penandatangan: params.penandatangan || "",
    jadwal: params.jadwal || "",
    status_aktif: true,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  writeAuditLog_(ctx.user.user_id, "CREATE_GROUP", "GROUP", groupId);
  return ok_(row);
}

function getMeetings_(ctx, params) {
  var repo = new SheetRepository_("meetings");
  var all = repo.getAll();
  if (params.group_id)
    all = all.filter(function (m) {
      return m.group_id === params.group_id;
    });
  if (params.from)
    all = all.filter(function (m) {
      return formatDate(m.tanggal) >= params.from;
    });
  if (params.to)
    all = all.filter(function (m) {
      return formatDate(m.tanggal) <= params.to;
    });
  all.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  return ok_(all.map(publicMeeting_));
}

function normalizeIso_(val) {
  if (!val) return "";
  if (val instanceof Date) return val.toISOString();
  return String(val);
}

function publicMeeting_(m) {
  if (!m) return null;
  var c = Object.assign({}, m);
  delete c._row;
  c.tanggal = formatDate(c.tanggal);
  c.kategori_target = parseKategoriTarget_(c.kategori_target);
  c.created_at = normalizeIso_(c.created_at);
  c.updated_at = normalizeIso_(c.updated_at);
  return c;
}

function createMeeting_(ctx, params) {
  if (!params.tanggal) return fail_("Tanggal wajib diisi");
  var repo = new SheetRepository_("meetings");
  var now = nowIso_();
  var meetingId = generateMeetingId();
  var kategoriTarget = "";
  if (params.kategori_target) {
    try {
      var kt =
        typeof params.kategori_target === "string"
          ? JSON.parse(params.kategori_target)
          : params.kategori_target;
      if (Array.isArray(kt)) kategoriTarget = JSON.stringify(kt);
    } catch (e) {
      kategoriTarget = "";
    }
  }
  var row = {
    meeting_id: meetingId,
    tanggal: formatDate(params.tanggal),
    hari: getHariFromDate(params.tanggal),
    jam: params.jam || "",
    group_id: params.group_id || "",
    acara: params.acara || "",
    materi: params.materi || "",
    status: params.status || "DIJADWALKAN",
    catatan: params.catatan || "",
    kategori_target: kategoriTarget,
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  writeAuditLog_(ctx.user.user_id, "CREATE_MEETING", "MEETING", meetingId);
  return ok_(publicMeeting_(row));
}

function updateMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");
  var repo = new SheetRepository_("meetings");
  var existing = repo.findById("meeting_id", params.meeting_id);
  if (!existing) return fail_("Meeting tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  ["jam", "group_id", "acara", "materi", "status", "catatan"].forEach(
    function (f) {
      if (params.hasOwnProperty(f)) patch[f] = params[f];
    },
  );
  if (params.hasOwnProperty("kategori_target")) {
    var kt = params.kategori_target;
    try {
      if (typeof kt === "string") kt = JSON.parse(kt);
      patch.kategori_target = Array.isArray(kt) ? JSON.stringify(kt) : "";
    } catch (e) {
      patch.kategori_target = "";
    }
  }
  if (params.tanggal) {
    patch.tanggal = formatDate(params.tanggal);
    patch.hari = getHariFromDate(params.tanggal);
  }
  var updated = repo.updateById("meeting_id", params.meeting_id, patch);
  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MEETING",
    "MEETING",
    params.meeting_id,
  );
  return ok_(publicMeeting_(updated));
}

function getAttendance_(ctx, params) {
  if (params.meeting_id) {
    var rows = getAttendanceByMeeting_(params.meeting_id);
    return ok_(rows);
  }
  if (params.member_id) {
    var repo = new SheetRepository_("attendance");
    var memberRows = repo.findByField("member_id", params.member_id);
    return ok_(
      memberRows.map(function (r) {
        var c = Object.assign({}, r);
        delete c._row;
        return c;
      }),
    );
  }
  return fail_("meeting_id atau member_id wajib diisi");
}

function saveAttendance_(ctx, params) {
  if (!params.meeting_id || !params.member_id || !params.status) {
    return fail_("meeting_id, member_id, dan status wajib diisi");
  }
  if (Object.keys(ATTENDANCE_STATUS).indexOf(params.status) === -1) {
    return fail_("Status absensi tidak valid");
  }
  var repo = new SheetRepository_("attendance");
  var existing = repo.find(function (a) {
    return (
      a.meeting_id === params.meeting_id && a.member_id === params.member_id
    );
  })[0];

  var now = nowIso_();
  if (existing) {
    var updated = repo.updateById("attendance_id", existing.attendance_id, {
      status: params.status,
      catatan: params.catatan || existing.catatan || "",
      updated_at: now,
    });
    invalidateAttendanceCache_(params.meeting_id);
    writeAuditLog_(
      ctx.user.user_id,
      "UPDATE_ATTENDANCE",
      "ATTENDANCE",
      existing.attendance_id,
    );
    return ok_(updated);
  }

  var attendanceId = generateAttendanceId();
  var row = {
    attendance_id: attendanceId,
    meeting_id: params.meeting_id,
    member_id: params.member_id,
    status: params.status,
    catatan: params.catatan || "",
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  invalidateAttendanceCache_(params.meeting_id);
  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_ATTENDANCE",
    "ATTENDANCE",
    attendanceId,
  );
  return ok_(row);
}

function bulkSaveAttendance_(ctx, params) {
  if (!params.meeting_id || !params.items) {
    return fail_("meeting_id dan items wajib diisi");
  }
  var items = params.items;
  if (typeof items === "string") {
    try {
      items = JSON.parse(items);
    } catch (e) {
      return fail_("items tidak valid JSON");
    }
  }
  if (!Array.isArray(items)) return fail_("items harus array");

  var repo = new SheetRepository_("attendance");
  var all = repo.getAll();
  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var now = nowIso_();

  var existingMap = {};
  all.forEach(function (a) {
    existingMap[a.meeting_id + "|" + a.member_id] = a;
  });

  var toInsert = [];
  var toUpdate = []; // {rowNum, patch}

  items.forEach(function (item) {
    if (!item.member_id || !item.status) return;
    if (Object.keys(ATTENDANCE_STATUS).indexOf(item.status) === -1) return;

    var key = params.meeting_id + "|" + item.member_id;
    var existing = existingMap[key];

    if (existing) {
      toUpdate.push({
        rowNum: existing._row,
        patch: {
          status: item.status,
          catatan: item.catatan || existing.catatan || "",
          updated_at: now,
        },
      });
    } else {
      toInsert.push({
        attendance_id: generateAttendanceId(),
        meeting_id: params.meeting_id,
        member_id: item.member_id,
        status: item.status,
        catatan: item.catatan || "",
        created_by: ctx.user.user_id,
        created_at: now,
        updated_at: now,
      });
    }
  });

  if (toInsert.length) {
    repo.insertMany(toInsert);
  }

  if (toUpdate.length) {
    batchUpdateRows_(sheet, headers, toUpdate);
    repo._invalidateCache();
  }

  invalidateAttendanceCache_(params.meeting_id);

  writeAuditLog_(
    ctx.user.user_id,
    "BULK_SAVE_ATTENDANCE",
    "MEETING",
    params.meeting_id,
  );
  return ok_({ inserted: toInsert.length, updated: toUpdate.length });
}

function deleteAttendance_(ctx, params) {
  if (!params.meeting_id || !params.member_id) {
    return fail_("meeting_id dan member_id wajib diisi");
  }

  var repo = new SheetRepository_("attendance");
  var existing = repo.find(function (a) {
    return (
      a.meeting_id === params.meeting_id && a.member_id === params.member_id
    );
  })[0];

  if (!existing) {
    return ok_({ deleted: 0 });
  }

  var deleted = repo.deleteById("attendance_id", existing.attendance_id);

  invalidateAttendanceCache_(params.meeting_id);

  writeAuditLog_(
    ctx.user.user_id,
    "DELETE_ATTENDANCE",
    "ATTENDANCE",
    existing.attendance_id,
  );

  return ok_({ deleted: deleted ? 1 : 0 });
}

function deleteAttendanceByMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var repo = new SheetRepository_("attendance");
  var all = repo.getAll();
  var toDelete = all.filter(function (a) {
    return a.meeting_id === params.meeting_id;
  });

  if (!toDelete.length) {
    return ok_({ deleted: 0 });
  }

  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var idColIndex = headers.indexOf("attendance_id");
  var lastRow = sheet.getLastRow();

  if (lastRow < 2) return ok_({ deleted: 0 });

  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();
  var idsToDelete = {};
  toDelete.forEach(function (a) {
    idsToDelete[a.attendance_id] = true;
  });

  // Kumpulkan row number (ascending)
  var rowsToDelete = [];
  for (var i = 0; i < idColValues.length; i++) {
    if (idsToDelete[idColValues[i][0]]) {
      rowsToDelete.push(i + 2); // +2: header row + 0-index
    }
  }

  if (!rowsToDelete.length) return ok_({ deleted: 0 });

  rowsToDelete.sort(function (a, b) {
    return a - b;
  });

  // OPTIMASI: group contiguous rows → 1 deleteRows per group
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

  // Hapus dari BELAKANG biar index tidak bergeser
  groups.sort(function (a, b) {
    return b.start - a.start;
  });
  groups.forEach(function (g) {
    sheet.deleteRows(g.start, g.count);
  });

  repo._invalidateCache();
  invalidateAttendanceCache_(params.meeting_id);
  if (ctx && ctx.user) {
    writeAuditLog_(
      ctx.user.user_id,
      "DELETE_ATTENDANCE_BY_MEETING",
      "MEETING",
      params.meeting_id,
    );
  }

  return ok_({ deleted: rowsToDelete.length });
}

function deleteAttendanceByMember_(ctx, params) {
  if (!params.member_id) return fail_("member_id wajib diisi");

  var repo = new SheetRepository_("attendance");
  var all = repo.getAll();
  var toDelete = all.filter(function (a) {
    return a.member_id === params.member_id;
  });

  if (!toDelete.length) return ok_({ deleted: 0 });

  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var idColIndex = headers.indexOf("attendance_id");
  var lastRow = sheet.getLastRow();

  if (lastRow < 2) return ok_({ deleted: 0 });

  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();
  var idsToDelete = {};
  toDelete.forEach(function (a) {
    idsToDelete[a.attendance_id] = true;
  });

  var rowsToDelete = [];
  for (var i = 0; i < idColValues.length; i++) {
    if (idsToDelete[idColValues[i][0]]) {
      rowsToDelete.push(i + 2);
    }
  }

  if (!rowsToDelete.length) return ok_({ deleted: 0 });

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

  repo._invalidateCache();

  var affectedMeetingIds = {};
  toDelete.forEach(function (a) {
    affectedMeetingIds[a.meeting_id] = true;
  });
  Object.keys(affectedMeetingIds).forEach(invalidateAttendanceCache_);

  writeAuditLog_(
    ctx.user.user_id,
    "DELETE_ATTENDANCE_BY_MEMBER",
    "MEMBER",
    params.member_id,
  );
  return ok_({ deleted: rowsToDelete.length });
}

function getMonitoring_(ctx, params) {
  if (!params.member_id) return fail_("member_id wajib diisi");
  if (ctx.user.role === ROLES.TIM_PNKB) {
    var check = getMemberDetail_(ctx, { member_id: params.member_id });
    if (!check.success) return check;
  }
  var repo = new SheetRepository_("monitoring");
  var rows = repo.findByField("member_id", params.member_id);
  rows.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  return ok_(
    rows.map(function (r) {
      var c = Object.assign({}, r);
      delete c._row;
      return c;
    }),
  );
}

function createMonitoring_(ctx, params) {
  if (!params.member_id || !params.status)
    return fail_("member_id dan status wajib diisi");
  if (Object.keys(MONITORING_STATUS).indexOf(params.status) === -1)
    return fail_("Status monitoring tidak valid");

  var repo = new SheetRepository_("monitoring");
  var now = nowIso_();
  var monitoringId = generateMonitoringId();
  var row = {
    monitoring_id: monitoringId,
    member_id: params.member_id,
    tanggal: params.tanggal ? formatDate(params.tanggal) : formatDate(now),
    jenis: params.jenis || "UMUM",
    status: params.status,
    catatan: params.catatan || "",
    tindak_lanjut: params.tindak_lanjut || "",
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);

  var membersRepo = new SheetRepository_("members");
  membersRepo.updateById("member_id", params.member_id, {
    status_pembinaan: params.status,
    updated_at: now,
  });

  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_MONITORING",
    "MONITORING",
    monitoringId,
  );
  return ok_(row);
}

function updateMonitoring_(ctx, params) {
  if (!params.monitoring_id) return fail_("monitoring_id wajib diisi");
  var repo = new SheetRepository_("monitoring");
  var existing = repo.findById("monitoring_id", params.monitoring_id);
  if (!existing) return fail_("Data monitoring tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  ["catatan", "tindak_lanjut"].forEach(function (f) {
    if (params.hasOwnProperty(f)) patch[f] = params[f];
  });
  var updated = repo.updateById("monitoring_id", params.monitoring_id, patch);
  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MONITORING",
    "MONITORING",
    params.monitoring_id,
  );
  return ok_(updated);
}

function getAnnouncementTemplates_(ctx, params) {
  var repo = new SheetRepository_("announcement_templates");
  var all = repo.getAll().filter(function (t) {
    return toBool_(t.status_aktif);
  });
  return ok_(
    all.map(function (t) {
      var c = Object.assign({}, t);
      delete c._row;
      return c;
    }),
  );
}

function renderTemplate_(templateText, data) {
  var text = templateText;
  Object.keys(data).forEach(function (key) {
    var re = new RegExp("\\{\\{\\s*" + key + "\\s*\\}\\}", "g");
    text = text.replace(
      re,
      data[key] === undefined || data[key] === null ? "" : data[key],
    );
  });
  return text;
}

var JADWAL_RUTIN = ["Minggu", "Selasa", "Kamis"];

function generateAnnouncement_(ctx, params) {
  if (!params.template_id || !params.group_id || !params.tanggal) {
    return fail_("template_id, group_id, dan tanggal wajib diisi");
  }
  var tplRepo = new SheetRepository_("announcement_templates");
  var template = tplRepo.findById("template_id", params.template_id);
  if (!template) return fail_("Template tidak ditemukan");

  var groupsRepo = new SheetRepository_("groups");
  var group = groupsRepo.findById("group_id", params.group_id);
  if (!group) return fail_("Kelompok tidak ditemukan");

  var hari = getHariFromDate(params.tanggal);
  var warning =
    JADWAL_RUTIN.indexOf(hari) === -1
      ? "Tanggal ini bukan jadwal rutin pengajian (" +
        JADWAL_RUTIN.join("/") +
        ")."
      : "";

  var data = {
    nama_kelompok: group.group_name,
    hari: hari,
    tanggal: formatDateShort(params.tanggal),
    jam: params.jam || "",
    acara: params.acara || "",
    materi: params.materi || "",
    catatan: params.catatan || "",
    penandatangan: params.penandatangan || group.penandatangan || "",
  };

  var text = renderTemplate_(template.isi_template, data);
  return ok_({
    generated_text: text,
    warning: warning,
    hari: hari,
    data: data,
  });
}

function generateWeeklyAnnouncements_(ctx, params) {
  if (!params.template_id || !params.group_id || !params.week_start) {
    return fail_("template_id, group_id, dan week_start wajib diisi");
  }
  var base = parseDate_(params.week_start);
  if (!base) return fail_("week_start tidak valid");

  var results = [];
  var dayOffsets = { Minggu: 0, Selasa: 2, Kamis: 4 };
  Object.keys(dayOffsets).forEach(function (hari) {
    var d = new Date(base.getTime());
    d.setDate(d.getDate() + dayOffsets[hari]);
    var tanggal = formatDate(d);
    var res = generateAnnouncement_(
      ctx,
      Object.assign({}, params, { tanggal: tanggal }),
    );
    results.push(
      Object.assign({ hari: hari, tanggal: tanggal }, res.data || {}),
    );
  });
  return ok_(results);
}

function createAnnouncement_(ctx, params) {
  var genResult = generateAnnouncement_(ctx, params);
  if (!genResult.success) return genResult;

  var repo = new SheetRepository_("announcements");
  var now = nowIso_();
  var announcementId = generateAnnouncementId();
  var row = {
    announcement_id: announcementId,
    template_id: params.template_id,
    meeting_id: params.meeting_id || "",
    group_id: params.group_id,
    tanggal: formatDate(params.tanggal),
    hari: genResult.data.hari,
    jam: params.jam || "",
    acara: params.acara || "",
    materi: params.materi || "",
    catatan: params.catatan || "",
    generated_text: genResult.data.generated_text,
    status: ANNOUNCEMENT_STATUS.DRAFT,
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_ANNOUNCEMENT",
    "ANNOUNCEMENT",
    announcementId,
  );
  return ok_(row);
}

function updateAnnouncement_(ctx, params) {
  if (!params.announcement_id) return fail_("announcement_id wajib diisi");
  var repo = new SheetRepository_("announcements");
  var existing = repo.findById("announcement_id", params.announcement_id);
  if (!existing) return fail_("Pengumuman tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  ["generated_text", "status", "jam", "acara", "materi", "catatan"].forEach(
    function (f) {
      if (params.hasOwnProperty(f)) patch[f] = params[f];
    },
  );
  var updated = repo.updateById(
    "announcement_id",
    params.announcement_id,
    patch,
  );
  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_ANNOUNCEMENT",
    "ANNOUNCEMENT",
    params.announcement_id,
  );
  return ok_(updated);
}

function getAnnouncements_(ctx, params) {
  var repo = new SheetRepository_("announcements");
  var all = repo.getAll();
  if (params.group_id)
    all = all.filter(function (a) {
      return a.group_id === params.group_id;
    });
  if (params.status)
    all = all.filter(function (a) {
      return a.status === params.status;
    });
  all.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  return ok_(
    all.map(function (a) {
      var c = Object.assign({}, a);
      delete c._row;
      return c;
    }),
  );
}

function getAnnouncementRecipientSummary_(ctx, params) {
  if (!params.group_id) return fail_("group_id wajib diisi");
  var membersRepo = new SheetRepository_("members");
  var members = membersRepo.find(function (m) {
    return m.kelompok === params.group_id && toBool_(m.status_aktif);
  });
  var withWa = members.filter(function (m) {
    return !!m.no_wa;
  });
  return ok_({
    total: members.length,
    dengan_wa: withWa.length,
    tanpa_wa: members.length - withWa.length,
  });
}

function getUsers_(ctx, params) {
  var repo = new SheetRepository_("users");
  return ok_(
    repo.getAll().map(function (u) {
      return publicUser_(Object.assign({}, u));
    }),
  );
}

function getUserDetail_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");

  var repo = new SheetRepository_("users");
  var user = repo.findById("user_id", params.user_id);
  if (!user) return fail_("User tidak ditemukan");

  return ok_(publicUser_(user));
}

function createUser_(ctx, params) {
  if (!params.username || !params.password || !params.role)
    return fail_("username, password, role wajib diisi");
  if (!ROLES[params.role]) return fail_("Role tidak valid");

  // ⭐ BARU: SEMUA role wajib punya member_id
  if (!params.member_id) {
    return fail_(
      "member_id wajib diisi. Setiap user harus terhubung ke jamaah.",
    );
  }

  // Validasi member_id benar-benar ada
  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", params.member_id);
  if (!member) {
    return fail_("Jamaah dengan member_id tersebut tidak ditemukan");
  }

  // ⭐ BARU: 1 member = 1 user (kecuali multi-akun disengaja)
  var repo = new SheetRepository_("users");
  var existingUserForMember = repo.find(function (u) {
    return u.member_id === params.member_id && toBool_(u.status_aktif);
  });
  if (existingUserForMember.length > 0) {
    return fail_("Jamaah ini sudah punya akun aktif");
  }

  var repo = new SheetRepository_("users");
  if (repo.findById("username", params.username))
    return fail_("Username sudah digunakan");

  var now = nowIso_();
  var userId = generateUserId();
  var row = {
    user_id: userId,
    username: params.username,
    password_hash: hashPassword_(params.password),
    nama: params.nama || params.username,
    role: params.role,
    member_id: params.member_id || "",
    status_aktif: true,
    created_at: now,
    updated_at: now,
    last_login_at: "",
  };
  repo.insert(row);
  writeAuditLog_(ctx.user.user_id, "CREATE_USER", "USER", userId);
  return ok_(publicUser_(row));
}

function updateUser_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");
  var repo = new SheetRepository_("users");
  var existing = repo.findById("user_id", params.user_id);
  if (!existing) return fail_("User tidak ditemukan");

  if (params.role && !ROLES[params.role]) {
    return fail_("Role tidak valid");
  }

  if (
    params.role &&
    params.user_id === ctx.user.user_id &&
    params.role !== ROLES.SUPER_ADMIN &&
    ctx.user.role === ROLES.SUPER_ADMIN
  ) {
    return fail_("Tidak bisa mengubah role diri sendiri dari SUPER_ADMIN");
  }

  if (
    params.role &&
    existing.role === ROLES.SUPER_ADMIN &&
    params.role !== ROLES.SUPER_ADMIN
  ) {
    var superAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (superAdmins.length === 0) {
      return fail_("Tidak bisa mengubah role SUPER_ADMIN terakhir");
    }
  }

  if (
    params.hasOwnProperty("status_aktif") &&
    !toBool_(params.status_aktif) &&
    existing.role === ROLES.SUPER_ADMIN
  ) {
    var activeSuperAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (activeSuperAdmins.length === 0) {
      return fail_("Tidak bisa menonaktifkan SUPER_ADMIN terakhir");
    }
  }

  if (
    params.hasOwnProperty("member_id") &&
    params.member_id !== existing.member_id
  ) {
    return fail_("member_id tidak bisa diubah setelah user dibuat");
  }

  var patch = { updated_at: nowIso_() };
  ["nama", "role", "status_aktif"].forEach(function (f) {
    if (params.hasOwnProperty(f)) patch[f] = params[f];
  });
  if (params.password) patch.password_hash = hashPassword_(params.password);

  var updated = repo.updateById("user_id", params.user_id, patch);
  writeAuditLog_(ctx.user.user_id, "UPDATE_USER", "USER", params.user_id);
  return ok_(publicUser_(updated));
}

function updateUserRole_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");
  if (!params.role) return fail_("role wajib diisi");
  if (!ROLES[params.role]) return fail_("Role tidak valid");

  var repo = new SheetRepository_("users");
  var existing = repo.findById("user_id", params.user_id);
  if (!existing) return fail_("User tidak ditemukan");

  if (
    params.user_id === ctx.user.user_id &&
    params.role !== ROLES.SUPER_ADMIN &&
    ctx.user.role === ROLES.SUPER_ADMIN
  ) {
    return fail_("Tidak bisa mengubah role diri sendiri dari SUPER_ADMIN");
  }

  if (
    existing.role === ROLES.SUPER_ADMIN &&
    params.role !== ROLES.SUPER_ADMIN
  ) {
    var superAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (superAdmins.length === 0) {
      return fail_("Tidak bisa mengubah role SUPER_ADMIN terakhir");
    }
  }

  var oldRole = existing.role;
  var now = nowIso_();
  var updated = repo.updateById("user_id", params.user_id, {
    role: params.role,
    updated_at: now,
  });

  writeAuditLog_(ctx.user.user_id, "UPDATE_USER_ROLE", "USER", params.user_id);

  return ok_(publicUser_(updated));
}

function getMemberUserStatus_(memberId) {
  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.find(function (u) {
    return u.member_id === memberId && toBool_(u.status_aktif);
  })[0];

  if (!user) {
    return {
      has_user: false,
      user: null,
    };
  }

  return {
    has_user: true,
    user: publicUser_(user),
  };
}

function getSettings_(ctx, params) {
  var repo = new SheetRepository_("settings");
  var all = repo.getAll();
  var out = {};
  all.forEach(function (s) {
    try {
      out[s.key] = JSON.parse(s.value);
    } catch (e) {
      out[s.key] = s.value;
    }
  });
  return ok_(out);
}

function updateSettings_(ctx, params) {
  if (!params.key) return fail_("key wajib diisi");
  var repo = new SheetRepository_("settings");
  var existing = repo.findById("key", params.key);
  var value =
    typeof params.value === "string"
      ? params.value
      : JSON.stringify(params.value);
  if (existing) {
    repo.updateById("key", params.key, { value: value, updated_at: nowIso_() });
  } else {
    repo.insert({ key: params.key, value: value, updated_at: nowIso_() });
  }
  writeAuditLog_(ctx.user.user_id, "UPDATE_SETTINGS", "SETTINGS", params.key);
  return ok_({ key: params.key, value: JSON.parse(value) });
}

function getAuditLogs_(ctx, params) {
  var repo = new SheetRepository_("audit_logs");
  // Audit log tidak perlu semua — ambil N terakhir saja
  var all = repo.getAllRecent(AUDIT_LOG_RECENT_LIMIT);
  if (params.user_id)
    all = all.filter(function (l) {
      return l.user_id === params.user_id;
    });
  if (params.target_type)
    all = all.filter(function (l) {
      return l.target_type === params.target_type;
    });
  all.sort(function (a, b) {
    return new Date(b.timestamp) - new Date(a.timestamp);
  });
  var limit = params.limit ? Number(params.limit) : 200;
  return ok_(
    all.slice(0, limit).map(function (l) {
      var c = Object.assign({}, l);
      delete c._row;
      return c;
    }),
  );
}

function getAiUsageToday_(ctx) {
  var today = formatDate(nowIso_());
  var repo = new SheetRepository_("ai_usage");
  return repo.find(function (u) {
    return (
      u.user_id === ctx.user.user_id && String(u.timestamp).indexOf(today) === 0
    );
  });
}

function checkAiQuota_(ctx) {
  var limit = AI_DAILY_LIMIT[ctx.user.role];
  if (!limit) return null;

  var todayUsage = getAiUsageToday_(ctx);
  if (todayUsage.length >= limit) {
    return fail_("Batas harian tercapai (" + limit + "x). Coba lagi besok.");
  }
  return null;
}

function logAiUsage_(ctx, provider, inputTokens, outputTokens) {
  try {
    var repo = new SheetRepository_("ai_usage");
    var total = (inputTokens || 0) + (outputTokens || 0);
    repo.insert({
      usage_id: generateUsageId(),
      user_id: ctx.user.user_id,
      user_nama: ctx.user.nama,
      role: ctx.user.role,
      provider: provider || "",
      input_tokens: inputTokens || 0,
      output_tokens: outputTokens || 0,
      total_tokens: total,
      timestamp: nowIso_(),
    });
  } catch (e) {
    Logger.log("Gagal log AI usage: " + e);
  }
}

// ============================================================
// OPTIMASI: getAllRecent — baca N row terakhir, bukan semua
// ============================================================
SheetRepository_.prototype.getAllRecent = function (limit) {
  var sheet = this._sheet();
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return [];

  var headers = this.def.headers;
  var startRow = Math.max(2, lastRow - limit + 1);
  var numRows = lastRow - startRow + 1;

  var values = sheet.getRange(startRow, 1, numRows, headers.length).getValues();
  return this._rowsToObjects(values);
};

// ============================================================
// OPTIMASI: Cache per-meeting untuk attendance
// ============================================================
function getAttendanceByMeeting_(meetingId) {
  if (!meetingId) return [];

  var cacheKey = ATTENDANCE_CACHE_PREFIX + meetingId;
  var cache = CacheService.getScriptCache();

  var cached = cache.get(cacheKey);
  if (cached) {
    try {
      return JSON.parse(cached);
    } catch (e) {
      /* fallthrough */
    }
  }

  var repo = new SheetRepository_("attendance");
  var rows = repo.findByField("meeting_id", meetingId);

  var cleaned = rows.map(function (r) {
    var c = Object.assign({}, r);
    delete c._row;
    return c;
  });

  try {
    var str = JSON.stringify(cleaned);
    if (str.length < 95000) {
      cache.put(cacheKey, str, ATTENDANCE_CACHE_TTL);
    }
  } catch (e) {
    /* ignore */
  }

  return cleaned;
}

function invalidateAttendanceCache_(meetingId) {
  if (!meetingId) return;
  try {
    CacheService.getScriptCache().remove(ATTENDANCE_CACHE_PREFIX + meetingId);
  } catch (e) {
    /* ignore */
  }
}

// ============================================================
// OPTIMASI: Batch update untuk bulkSaveAttendance_
// Update banyak row dalam 1 setValues call (group by contiguous rows)
// ============================================================
function batchUpdateRows_(sheet, headers, updates) {
  if (!updates || !updates.length) return 0;

  // Sort by rowNum ascending
  updates.sort(function (a, b) {
    return a.rowNum - b.rowNum;
  });

  // Group contiguous rows
  var groups = [];
  var current = null;

  updates.forEach(function (u) {
    if (current && u.rowNum === current.endRow + 1) {
      current.endRow = u.rowNum;
      current.items.push(u);
    } else {
      if (current) groups.push(current);
      current = { startRow: u.rowNum, endRow: u.rowNum, items: [u] };
    }
  });
  if (current) groups.push(current);

  // Baca semua range sekaligus, update, tulis balik
  var totalUpdated = 0;

  groups.forEach(function (g) {
    var numRows = g.endRow - g.startRow + 1;
    var range = sheet.getRange(g.startRow, 1, numRows, headers.length);
    var values = range.getValues();

    g.items.forEach(function (u) {
      var idx = u.rowNum - g.startRow;
      var row = values[idx];
      headers.forEach(function (h, c) {
        if (u.patch.hasOwnProperty(h)) {
          row[c] = u.patch[h];
        }
      });
      totalUpdated++;
    });

    range.setValues(values);
  });

  return totalUpdated;
}

// ============================================================
// OPTIMASI: Lock helper — hanya write actions yang butuh lock
// ============================================================
function shouldUseLock_(action) {
  return WRITE_ACTIONS[action] === true;
}

function auditUsersWithoutMember() {
  var usersRepo = new SheetRepository_("users");
  var users = usersRepo.getAll();
  var orphan = users.filter(function (u) {
    return !u.member_id;
  });

  Logger.log("=== AUDIT USER TANPA member_id ===");
  Logger.log("Total user: " + users.length);
  Logger.log("Tanpa member_id: " + orphan.length);
  Logger.log("");

  orphan.forEach(function (u) {
    Logger.log(
      "  - " +
        u.user_id +
        " | @" +
        u.username +
        " | " +
        u.nama +
        " | role: " +
        u.role,
    );
  });

  if (orphan.length === 0) {
    Logger.log("Semua user sudah punya member_id.");
  }

  return {
    total: users.length,
    orphan_count: orphan.length,
    orphan: orphan.map(function (u) {
      return {
        user_id: u.user_id,
        username: u.username,
        nama: u.nama,
        role: u.role,
      };
    }),
  };
}
