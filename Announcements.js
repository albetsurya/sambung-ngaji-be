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
