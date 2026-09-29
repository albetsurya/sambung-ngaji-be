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
    var formatted = Utilities.formatDate(dateStr, "Asia/Jakarta", "yyyy-MM-dd");
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
    return Utilities.formatDate(dateStr, "Asia/Jakarta", "yyyy-MM-dd");
  }

  var str = String(dateStr).trim();

  if (/^\d{4}-\d{2}-\d{2}$/.test(str)) {
    return str;
  }

  if (/^\d{2}-\d{2}-\d{4}$/.test(str)) {
    var parts = str.split("-");
    return parts[2] + "-" + parts[1] + "-" + parts[0];
  }

  if (/^\d{4}-\d{2}-\d{2}T/.test(str)) {
    return Utilities.formatDate(new Date(str), "Asia/Jakarta", "yyyy-MM-dd");
  }

  var p = parseIsoParts_(str);
  if (!p) return "";

  return (
    String(p.year) +
    "-" +
    String(p.month).padStart(2, "0") +
    "-" +
    String(p.day).padStart(2, "0")
  );
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
      ? MEMBER_CATEGORY.ISTIMEWA
      : MEMBER_CATEGORY.DEWASA;
  }

  var jenjang = String(member.jenjang_pendidikan || "").toUpperCase();

  if (jenjang === "PAUD" || jenjang === "TK") {
    return MEMBER_CATEGORY.CABERAWIT;
  }
  if (jenjang === "SD") return MEMBER_CATEGORY.CABERAWIT;
  if (jenjang === "SMP") return MEMBER_CATEGORY.PRA_REMAJA;
  if (jenjang === "SMA" || jenjang === "SMK") return MEMBER_CATEGORY.REMAJA;

  var age2 = getMemberAge(member.tanggal_lahir);
  if (age2 !== null && age2 >= 60) return MEMBER_CATEGORY.ISTIMEWA;
  if (age2 !== null && age2 < 6) return MEMBER_CATEGORY.BALITA;
  if (age2 !== null && age2 < 13) return MEMBER_CATEGORY.CABERAWIT;
  if (age2 !== null && age2 < 16) return MEMBER_CATEGORY.PRA_REMAJA;
  if (age2 !== null && age2 < 19) return MEMBER_CATEGORY.REMAJA;

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
  // JANGAN pakai toISOString() untuk tanggal: sheet 1 Agu 00:00 WIB
  // = 31 Jul 17:00 UTC → mundur 1 hari. Format sebagai kalender WIB.
  if (val instanceof Date) {
    return Utilities.formatDate(val, "Asia/Jakarta", "yyyy-MM-dd");
  }
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
