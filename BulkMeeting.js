/**
 * BulkMeeting.js
 *
 * Fitur: Bulk create jadwal pengajian dalam 1 bulan + auto kirim WA
 * untuk kategori CABERAWIT & BALITA.
 */

var BULK_MEETING_MAX_WA = 300;
var BULK_MEETING_WA_KATEGORI = ["CABERAWIT", "BALITA"];

var HARI_INDEX_ = {
  Minggu: 0,
  Senin: 1,
  Selasa: 2,
  Rabu: 3,
  Kamis: 4,
  Jumat: 5,
  Sabtu: 6,
};

function pad2_(n) {
  return String(n).length < 2 ? "0" + n : String(n);
}

/* -------------------------------------------------------------------------- */
/*                              VALIDATION                                    */
/* -------------------------------------------------------------------------- */

function validateBulkParams_(params) {
  if (!params.tahun || !params.bulan)
    return { error: "tahun dan bulan wajib diisi" };

  var tahun = Number(params.tahun);
  var bulan = Number(params.bulan);
  if (isNaN(tahun) || tahun < 2020 || tahun > 2100)
    return { error: "tahun tidak valid" };
  if (isNaN(bulan) || bulan < 1 || bulan > 12)
    return { error: "bulan tidak valid" };

  var hari = params.hari;
  if (typeof hari === "string") {
    try {
      hari = JSON.parse(hari);
    } catch (e) {
      return { error: "hari tidak valid" };
    }
  }
  if (!Array.isArray(hari) || hari.length === 0)
    return { error: "hari wajib dipilih minimal 1" };

  if (!params.jam) return { error: "jam wajib diisi" };
  if (!params.acara) return { error: "acara wajib diisi" };
  if (!params.group_id) return { error: "kelompok wajib dipilih" };

  return { tahun: tahun, bulan: bulan, hari: hari };
}

/* -------------------------------------------------------------------------- */
/*                              DATE BUILDER                                  */
/* -------------------------------------------------------------------------- */

function buildDateList_(v) {
  var dayIndexes = v.hari
    .map(function (h) {
      return HARI_INDEX_[h];
    })
    .filter(function (i) {
      return typeof i === "number";
    });
  if (dayIndexes.length === 0) return [];

  var daysInMonth = new Date(v.tahun, v.bulan, 0).getDate();
  var result = [];
  for (var d = 1; d <= daysInMonth; d++) {
    var date = new Date(v.tahun, v.bulan - 1, d);
    var dayOfWeek = date.getDay();
    if (dayIndexes.indexOf(dayOfWeek) === -1) continue;
    var iso = v.tahun + "-" + pad2_(v.bulan) + "-" + pad2_(d);
    result.push({ iso: iso, hari: HARI_ID[dayOfWeek] });
  }
  return result;
}

/* -------------------------------------------------------------------------- */
/*                              TARGET MEMBERS                                */
/* -------------------------------------------------------------------------- */

function getBulkTargetMembers_(params) {
  var targets = params.kategori_target || [];
  if (typeof targets === "string") {
    try {
      targets = JSON.parse(targets);
    } catch (e) {
      targets = [];
    }
  }
  if (!Array.isArray(targets)) targets = [];

  var waKategori = targets.filter(function (k) {
    return BULK_MEETING_WA_KATEGORI.indexOf(k) !== -1;
  });
  if (waKategori.length === 0) return [];

  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();

  return all.filter(function (m) {
    if (!toBool_(m.status_aktif)) return false;
    if (!m.no_wa) return false;
    if (params.group_id && m.kelompok !== params.group_id) return false;
    var cat = getMemberCategory(m);
    return waKategori.indexOf(cat) !== -1;
  });
}

/* -------------------------------------------------------------------------- */
/*                              PREVIEW                                       */
/* -------------------------------------------------------------------------- */

function previewBulkMeetings_(ctx, params) {
  var v = validateBulkParams_(params);
  if (v.error) return fail_(v.error);

  var dates = buildDateList_(v);
  if (dates.length === 0)
    return fail_(
      "Tidak ada tanggal valid di bulan ini untuk hari yang dipilih.",
    );

  var meetingsRepo = new SheetRepository_("meetings");
  var existingKey = {};
  meetingsRepo.getAll().forEach(function (m) {
    existingKey[formatDate(m.tanggal) + "|" + (m.group_id || "")] = true;
  });

  var meetings = dates.map(function (d) {
    return {
      tanggal: d.iso,
      tanggal_display: formatDateShort(d.iso),
      hari: d.hari,
      sudah_ada: existingKey[d.iso + "|" + (v.group_id || "")] === true,
    };
  });

  var toCreate = meetings.filter(function (m) {
    return !m.sudah_ada;
  }).length;
  var memberTarget = getBulkTargetMembers_(params);
  var totalWa = toCreate * memberTarget.length;

  return ok_({
    total_dates: dates.length,
    total_new: toCreate,
    total_existing: meetings.length - toCreate,
    meetings: meetings,
    member_target: memberTarget.map(function (m) {
      return {
        member_id: m.member_id,
        nama_lengkap: m.nama_lengkap,
        no_wa: m.no_wa,
        kategori: getMemberCategory(m),
      };
    }),
    total_wa: totalWa,
    wa_overflow: totalWa > BULK_MEETING_MAX_WA,
    wa_max: BULK_MEETING_MAX_WA,
  });
}

/* -------------------------------------------------------------------------- */
/*                              BULK CREATE                                   */
/* -------------------------------------------------------------------------- */

function bulkCreateMeetings_(ctx, params) {
  var v = validateBulkParams_(params);
  if (v.error) return fail_(v.error);

  var sendWa = String(params.send_wa) === "true" || params.send_wa === true;
  var templateId = params.template_id || "";

  var template = null;
  if (sendWa) {
    if (!templateId) return fail_("Template WA wajib dipilih");
    var tplRepo = new SheetRepository_("announcement_templates");
    template = tplRepo.findById("template_id", templateId);
    if (!template) return fail_("Template tidak ditemukan");
    if (!toBool_(template.status_aktif)) return fail_("Template tidak aktif");
  }

  var dates = buildDateList_(v);
  var meetingsRepo = new SheetRepository_("meetings");
  var existingKey = {};
  meetingsRepo.getAll().forEach(function (m) {
    existingKey[formatDate(m.tanggal) + "|" + (m.group_id || "")] = true;
  });

  var newDates = dates.filter(function (d) {
    return !existingKey[d.iso + "|" + (v.group_id || "")];
  });

  if (newDates.length === 0)
    return fail_("Semua tanggal sudah ada. Tidak ada yang dibuat.");

  var memberTarget = sendWa ? getBulkTargetMembers_(params) : [];
  var totalWa = newDates.length * memberTarget.length;

  if (sendWa && totalWa > BULK_MEETING_MAX_WA) {
    return fail_(
      "Total WA " +
        totalWa +
        " melebihi batas " +
        BULK_MEETING_MAX_WA +
        ". Kurangi jumlah hari atau kirim per minggu.",
    );
  }

  var groupsRepo = new SheetRepository_("groups");
  var group = groupsRepo.findById("group_id", v.group_id);
  var groupName = group ? group.group_name : "";
  var penandatangan = (group && group.penandatangan) || "";

  var now = nowIso_();
  var createdMeetings = [];
  var waSent = 0;
  var waFailed = 0;
  var waDetails = [];

  newDates.forEach(function (d) {
    var meeting = {
      meeting_id: generateMeetingId(),
      tanggal: d.iso,
      hari: d.hari,
      jam: params.jam,
      group_id: v.group_id,
      acara: params.acara,
      materi: params.materi || "",
      status: "SCHEDULED",
      catatan: params.catatan || "",
      kategori_target: JSON.stringify(params.kategori_target || []),
      created_by: ctx.user.user_id,
      created_at: now,
      updated_at: now,
    };
    meetingsRepo.insert(meeting);
    createdMeetings.push(meeting);

    if (sendWa && memberTarget.length > 0) {
      var templateData = {
        nama_kelompok: groupName,
        hari: d.hari,
        tanggal: formatDateShort(d.iso),
        jam: params.jam,
        acara: params.acara,
        materi: params.materi || "",
        catatan: params.catatan || "",
        penandatangan: penandatangan,
      };
      var messageText = renderTemplate_(template.isi_template, templateData);

      memberTarget.forEach(function (m) {
        var result = sendWhatsApp_(m.no_wa, messageText);
        if (result && result.ok) waSent++;
        else if (result && result.skipped) {
          /* gateway off, skip */
        } else {
          waFailed++;
          waDetails.push({
            member_id: m.member_id,
            no_wa: m.no_wa,
            error: (result && (result.reason || result.error)) || "unknown",
          });
        }
      });
    }
  });

  invalidateDashboardCache_();

  writeAuditLog_(
    ctx.user.user_id,
    "BULK_CREATE_MEETINGS",
    "MEETING",
    v.group_id,
  );

  return ok_({
    created: createdMeetings.length,
    skipped: dates.length - newDates.length,
    wa_sent: waSent,
    wa_failed: waFailed,
    wa_details: waDetails.slice(0, 10),
    meetings: createdMeetings.map(function (m) {
      return {
        meeting_id: m.meeting_id,
        tanggal: m.tanggal,
        tanggal_display: formatDateShort(m.tanggal),
        hari: m.hari,
      };
    }),
  });
}

/* -------------------------------------------------------------------------- */
/*                              TEMPLATE LIST                                 */
/* -------------------------------------------------------------------------- */

function getBulkMeetingTemplates_(ctx) {
  var repo = new SheetRepository_("announcement_templates");
  var all = repo.getAll().filter(function (t) {
    return toBool_(t.status_aktif);
  });
  return ok_(
    all.map(function (t) {
      return {
        template_id: t.template_id,
        nama_template: t.nama_template,
        kode: t.kode,
        isi_template: t.isi_template,
      };
    }),
  );
}
