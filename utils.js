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

  /* Handle Date object — pakai Utilities.formatDate dengan timezone
     eksplisit supaya tidak bergeser karena GAS runtime pakai UTC. */
  if (dateStr instanceof Date) {
    var tz = "Asia/Jakarta"; // hardcode WIB untuk konsistensi
    var formatted = Utilities.formatDate(dateStr, tz, "yyyy-MM-dd");
    var parts = formatted.split("-");
    return {
      year: Number(parts[0]),
      month: Number(parts[1]),
      day: Number(parts[2]),
    };
  }

  var iso = String(dateStr).slice(0, 10);
  var parts2 = iso.split("-");
  var yy = Number(parts2[0]);
  var mm = Number(parts2[1]);
  var dd = Number(parts2[2]);
  if (!yy || !mm || !dd) return null;
  return { year: yy, month: mm, day: dd };
}

function parseDate_(str) {
  var p = parseIsoParts_(str);
  if (!p) return null;
  return new Date(p.year, p.month - 1, p.day);
}

function formatDate(dateStr) {
  if (!dateStr) return "";

  if (dateStr instanceof Date) {
    /* Fallback: hitung manual WIB offset (UTC+7).
       Cara ini tidak bergantung pada GAS timezone config. */
    var utcMs = dateStr.getTime();
    var wibMs = utcMs + 7 * 60 * 60 * 1000; // +7 jam
    var wibDate = new Date(wibMs);
    var y = wibDate.getUTCFullYear();
    var m = wibDate.getUTCMonth() + 1;
    var d = wibDate.getUTCDate();
    var pad = function (n) {
      return String(n).length < 2 ? "0" + n : String(n);
    };
    return y + "-" + pad(m) + "-" + pad(d);
  }

  var p = parseIsoParts_(dateStr);
  if (!p) return "";
  var pad2 = function (n) {
    return String(n).length < 2 ? "0" + n : String(n);
  };
  return p.year + "-" + pad2(p.month) + "-" + pad2(p.day);
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

/* -------------------------------------------------------------------------- */
/*                          KATEGORI JAMAAH                                   */
/* -------------------------------------------------------------------------- */
/*
 * Logika kategorisasi (urutan prioritas):
 *
 * 1. Sudah menikah:
 *    - 60+ tahun          → ISTIMEWA
 *    - < 60 tahun         → DEWASA
 *
 * 2. Berdasarkan jenjang pendidikan:
 *    - PAUD / TK          → CABERAWIT  (selalu, apapun usia)
 *    - SD                 → CABERAWIT
 *    - SMP                → PRA_REMAJA
 *    - SMA / SMK          → REMAJA
 *
 * 3. Fallback berdasarkan usia (untuk yang belum sekolah / jenjang kosong):
 *    - < 6 tahun          → BALITA
 *    - 6–12 tahun         → CABERAWIT
 *    - 13–15 tahun        → PRA_REMAJA
 *    - 16–18 tahun        → REMAJA
 *    - 60+ tahun          → ISTIMEWA
 *    - else               → PRA_NIKAH
 *
 * Catatan: BALITA hanya untuk anak yang BELUM sekolah (jenjang kosong).
 *          Begitu tercatat PAUD/TK/SD, otomatis CABERAWIT.
 */
function getMemberCategory(member) {
  if (!member) return null;

  /* -------- 1. Sudah menikah -------- */
  if (toBool_(member.is_nikah)) {
    var age = getMemberAge(member.tanggal_lahir);
    return age !== null && age >= 60
      ? MEMBER_CATEGORY.ISTIMEWA
      : MEMBER_CATEGORY.DEWASA;
  }

  /* -------- 2. Berdasarkan jenjang pendidikan -------- */
  var jenjang = String(member.jenjang_pendidikan || "").toUpperCase();

  if (jenjang === "PAUD" || jenjang === "TK") {
    return MEMBER_CATEGORY.CABERAWIT;
  }
  if (jenjang === "SD") return MEMBER_CATEGORY.CABERAWIT;
  if (jenjang === "SMP") return MEMBER_CATEGORY.PRA_REMAJA;
  if (jenjang === "SMA" || jenjang === "SMK") return MEMBER_CATEGORY.REMAJA;

  /* -------- 3. Fallback berdasarkan usia -------- */
  /* Hanya untuk yang belum punya jenjang pendidikan. */
  var age2 = getMemberAge(member.tanggal_lahir);
  if (age2 !== null && age2 >= 60) return MEMBER_CATEGORY.ISTIMEWA;
  if (age2 !== null && age2 < 6) return MEMBER_CATEGORY.BALITA;
  if (age2 !== null && age2 < 13) return MEMBER_CATEGORY.CABERAWIT;
  if (age2 !== null && age2 < 16) return MEMBER_CATEGORY.PRA_REMAJA;
  if (age2 !== null && age2 < 19) return MEMBER_CATEGORY.REMAJA;

  /* -------- 4. Default -------- */
  return MEMBER_CATEGORY.PRA_NIKAH;
}

function toBool_(v) {
  return v === true || v === "true" || v === "TRUE" || v === 1 || v === "1";
}

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

function normalizeIso_(val) {
  if (!val) return "";
  if (val instanceof Date) return val.toISOString();
  return String(val);
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
