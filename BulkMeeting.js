/**
 * BulkMeeting.js
 *
 * Fitur: Bulk create jadwal pengajian dalam 1 bulan.
 * TANPA WA otomatis — fokus bulk create saja.
 *
 * WA bisa dikirim manual via fitur announcement per pertemuan.
 */

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

  return ok_({
    total_dates: dates.length,
    total_new: toCreate,
    total_existing: meetings.length - toCreate,
    meetings: meetings,
  });
}

/* -------------------------------------------------------------------------- */
/*                              BULK CREATE                                   */
/* -------------------------------------------------------------------------- */

function bulkCreateMeetings_(ctx, params) {
  var v = validateBulkParams_(params);
  if (v.error) return fail_(v.error);

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

  var now = nowIso_();
  var createdMeetings = [];

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
/*                              TEMPLATE LIST (untuk future WA)               */
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
