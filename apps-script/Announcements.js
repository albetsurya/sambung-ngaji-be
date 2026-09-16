var JADWAL_RUTIN = ["Minggu", "Selasa", "Kamis"];

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
  invalidateDashboardCache_();
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
  invalidateDashboardCache_();
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

  var isPaged = String(params.paged) === "true";

  var limit = params.limit !== undefined ? Number(params.limit) : 50;
  if (isNaN(limit) || limit <= 0) limit = 50;
  if (limit > 200) limit = 200;

  var offset = params.offset !== undefined ? Number(params.offset) : 0;
  if (isNaN(offset) || offset < 0) offset = 0;

  var total = all.length;
  var paged = all.slice(offset, offset + limit);

  var items = paged.map(function (a) {
    var c = Object.assign({}, a);
    delete c._row;
    return c;
  });

  if (!isPaged) {
    return ok_(items);
  }

  return ok_({
    items: items,
    total: total,
    limit: limit,
    offset: offset,
    has_more: offset + items.length < total,
  });
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

/* -------------------------------------------------------------------------- */
/*                      ANNOUNCEMENT TEMPLATE CRUD                            */
/* -------------------------------------------------------------------------- */

/**
 * Ambil detail template (termasuk inactive) — untuk edit.
 */
function getAnnouncementTemplateDetail_(ctx, params) {
  if (!params.template_id) return fail_("template_id wajib diisi");
  var repo = new SheetRepository_("announcement_templates");
  var t = repo.findById("template_id", params.template_id);
  if (!t) return fail_("Template tidak ditemukan");
  var c = Object.assign({}, t);
  delete c._row;
  return ok_(c);
}

/**
 * Ambil semua template (termasuk inactive) — untuk halaman kelola.
 */
function getAllAnnouncementTemplates_(ctx, params) {
  var repo = new SheetRepository_("announcement_templates");
  var all = repo.getAll();
  if (String(params.include_inactive) !== "true") {
    all = all.filter(function (t) {
      return toBool_(t.status_aktif);
    });
  }
  all.sort(function (a, b) {
    return String(a.nama_template || "").localeCompare(
      String(b.nama_template || ""),
    );
  });
  return ok_(
    all.map(function (t) {
      var c = Object.assign({}, t);
      delete c._row;
      return c;
    }),
  );
}

/**
 * Create template baru.
 */
function createAnnouncementTemplate_(ctx, params) {
  var nama = String(params.nama_template || "").trim();
  var kode = String(params.kode || "")
    .trim()
    .toUpperCase();
  var isi = String(params.isi_template || "");

  if (!nama) return fail_("Nama template wajib diisi");
  if (!kode) return fail_("Kode wajib diisi");
  if (!isi) return fail_("Isi template wajib diisi");
  if (!/^[A-Z0-9_]{3,30}$/.test(kode)) {
    return fail_("Kode harus 3-30 karakter (huruf, angka, underscore)");
  }

  var repo = new SheetRepository_("announcement_templates");
  var dup = repo.find(function (t) {
    return String(t.kode || "").toUpperCase() === kode;
  });
  if (dup.length > 0) return fail_("Kode template sudah dipakai");

  var now = nowIso_();
  var row = {
    template_id: newId_("TPL"),
    nama_template: nama,
    kode: kode,
    isi_template: isi,
    status_aktif: params.status_aktif === false ? false : true,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_ANNOUNCEMENT_TEMPLATE",
    "TEMPLATE",
    row.template_id,
  );
  return ok_(row);
}

/**
 * Update template.
 */
function updateAnnouncementTemplate_(ctx, params) {
  if (!params.template_id) return fail_("template_id wajib diisi");
  var repo = new SheetRepository_("announcement_templates");
  var existing = repo.findById("template_id", params.template_id);
  if (!existing) return fail_("Template tidak ditemukan");

  if (params.kode) {
    var kodeUp = String(params.kode).toUpperCase();
    if (!/^[A-Z0-9_]{3,30}$/.test(kodeUp)) {
      return fail_("Kode harus 3-30 karakter (huruf, angka, underscore)");
    }
    var dup = repo.find(function (t) {
      return (
        t.template_id !== params.template_id &&
        String(t.kode || "").toUpperCase() === kodeUp
      );
    });
    if (dup.length > 0) return fail_("Kode template sudah dipakai");
  }

  var patch = { updated_at: nowIso_() };
  if (params.hasOwnProperty("nama_template"))
    patch.nama_template = String(params.nama_template).trim();
  if (params.hasOwnProperty("kode"))
    patch.kode = String(params.kode).toUpperCase();
  if (params.hasOwnProperty("isi_template"))
    patch.isi_template = params.isi_template;
  if (params.hasOwnProperty("status_aktif"))
    patch.status_aktif = params.status_aktif !== false;

  var updated = repo.updateById("template_id", params.template_id, patch);
  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_ANNOUNCEMENT_TEMPLATE",
    "TEMPLATE",
    params.template_id,
  );
  return ok_(updated);
}

/**
 * Soft delete — set status_aktif = false.
 */
function deleteAnnouncementTemplate_(ctx, params) {
  if (!params.template_id) return fail_("template_id wajib diisi");
  var repo = new SheetRepository_("announcement_templates");
  var existing = repo.findById("template_id", params.template_id);
  if (!existing) return fail_("Template tidak ditemukan");

  repo.updateById("template_id", params.template_id, {
    status_aktif: false,
    updated_at: nowIso_(),
  });
  writeAuditLog_(
    ctx.user.user_id,
    "DELETE_ANNOUNCEMENT_TEMPLATE",
    "TEMPLATE",
    params.template_id,
  );
  return ok_({ deleted: true, template_id: params.template_id });
}

/**
 * Create template dari hasil WA/announcement existing.
 * params:
 *  - source_announcement_id (opsional — kalau mau copy dari pengumuman)
 *  - isi_template (opsional — kalau input manual)
 *  - nama_template (wajib)
 *  - kode (wajib)
 */
function createTemplateFromAnnouncement_(ctx, params) {
  var text = "";

  if (params.source_announcement_id) {
    var annRepo = new SheetRepository_("announcements");
    var ann = annRepo.findById(
      "announcement_id",
      params.source_announcement_id,
    );
    if (!ann) return fail_("Pengumuman sumber tidak ditemukan");
    text = ann.generated_text || "";
  } else if (params.isi_template) {
    text = params.isi_template;
  }

  if (!text) return fail_("Isi template kosong");

  return createAnnouncementTemplate_(ctx, {
    nama_template: params.nama_template,
    kode: params.kode,
    isi_template: text,
    status_aktif: true,
  });
}
