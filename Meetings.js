function getMeetings_(ctx, params) {
  var repo = new SheetRepository_("meetings");
  var all = repo.getAll();

  if (params.from) {
    all = all.filter(function (m) {
      return String(formatDate(m.tanggal)) >= String(params.from);
    });
  }
  if (params.to) {
    all = all.filter(function (m) {
      return String(formatDate(m.tanggal)) <= String(params.to);
    });
  }
  if (params.group_id) {
    all = all.filter(function (m) {
      return m.group_id === params.group_id;
    });
  }

  all.sort(function (a, b) {
    return parseDate_(b.tanggal) - parseDate_(a.tanggal);
  });

  return ok_(all.map(publicMeeting_));
}

function createMeeting_(ctx, params) {
  if (!params.tanggal) return fail_("Tanggal wajib diisi");
  if (!params.acara) return fail_("Acara wajib diisi");

  var now = nowIso_();
  var meetingId = generateMeetingId();

  /* Resolve jam_start: dari params, atau default dari kategori/acara. */
  var kategoriTarget = params.kategori_target || [];
  var jamStart =
    params.jam_start ||
    resolveDefaultJamStart_({
      acara: params.acara,
      kategori_target: kategoriTarget,
    });

  var meeting = {
    meeting_id: meetingId,
    tanggal: formatDate(params.tanggal),
    hari: getHariFromDate(params.tanggal),
    jam: params.jam || "",
    jam_start: jamStart,
    group_id: params.group_id || "",
    acara: params.acara,
    materi: params.materi || "",
    status: params.status || "SCHEDULED",
    catatan: params.catatan || "",
    kategori_target: JSON.stringify(kategoriTarget),
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };

  var repo = new SheetRepository_("meetings");
  repo.insert(meeting);

  invalidateDashboardCache_();

  writeAuditLog_(ctx.user.user_id, "CREATE_MEETING", "MEETING", meetingId);

  return ok_(publicMeeting_(meeting));
}

function updateMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var repo = new SheetRepository_("meetings");
  var existing = repo.findById("meeting_id", params.meeting_id);
  if (!existing) return fail_("Jadwal tidak ditemukan");

  var patch = { updated_at: nowIso_() };

  if (params.hasOwnProperty("tanggal")) {
    patch.tanggal = formatDate(params.tanggal);
    patch.hari = getHariFromDate(params.tanggal);
  }
  if (params.hasOwnProperty("jam")) patch.jam = params.jam;
  if (params.hasOwnProperty("jam_start")) patch.jam_start = params.jam_start;
  if (params.hasOwnProperty("group_id")) patch.group_id = params.group_id;
  if (params.hasOwnProperty("acara")) patch.acara = params.acara;
  if (params.hasOwnProperty("materi")) patch.materi = params.materi;
  if (params.hasOwnProperty("status")) patch.status = params.status;
  if (params.hasOwnProperty("catatan")) patch.catatan = params.catatan;
  if (params.hasOwnProperty("kategori_target")) {
    patch.kategori_target = JSON.stringify(params.kategori_target || []);
  }

  var updated = repo.updateById("meeting_id", params.meeting_id, patch);

  invalidateDashboardCache_();
  invalidateAttendanceCache_(params.meeting_id);

  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MEETING",
    "MEETING",
    params.meeting_id,
  );

  return ok_(publicMeeting_(updated));
}

/* -------------------------------------------------------------------------- */
/*                              Delete meeting                                */
/* -------------------------------------------------------------------------- */

function deleteMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var meetingsRepo = new SheetRepository_("meetings");
  var existing = meetingsRepo.findById("meeting_id", params.meeting_id);
  if (!existing) return fail_("Jadwal tidak ditemukan");

  /* Hapus semua absensi terkait via batch delete. */
  var deleteAttendanceResult = deleteAttendanceByMeeting_(ctx, {
    meeting_id: params.meeting_id,
  });
  var deletedAttendance = deleteAttendanceResult.success
    ? deleteAttendanceResult.data.deleted
    : 0;

  meetingsRepo.deleteById("meeting_id", params.meeting_id);

  invalidateDashboardCache_();
  invalidateAttendanceCache_(params.meeting_id);

  writeAuditLog_(
    ctx.user.user_id,
    "DELETE_MEETING",
    "MEETING",
    params.meeting_id,
  );

  return ok_({
    meeting_id: params.meeting_id,
    deleted_attendance: deletedAttendance,
  });
}

function publicMeeting_(m) {
  var out = Object.assign({}, m);
  delete out._row;

  out.tanggal = formatDate(m.tanggal);

  try {
    out.kategori_target = JSON.parse(m.kategori_target || "[]");
  } catch (e) {
    out.kategori_target = [];
  }
  return out;
}
