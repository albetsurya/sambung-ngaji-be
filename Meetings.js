function getMeetings_(ctx, params) {
  var repo = new SheetRepository_("meetings");
  var all = repo.getAll();

  if (params.from) {
    all = all.filter(function (m) {
      return String(m.tanggal) >= String(params.from);
    });
  }
  if (params.to) {
    all = all.filter(function (m) {
      return String(m.tanggal) <= String(params.to);
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

  var meeting = {
    meeting_id: meetingId,
    tanggal: formatDate(params.tanggal),
    hari: getHariFromDate_(params.tanggal),
    jam: params.jam || "",
    group_id: params.group_id || "",
    acara: params.acara,
    materi: params.materi || "",
    status: params.status || "SCHEDULED",
    catatan: params.catatan || "",
    kategori_target: JSON.stringify(params.kategori_target || []),
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };

  var repo = new SheetRepository_("meetings");
  repo.insert(meeting);

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
    patch.hari = getHariFromDate_(params.tanggal);
  }
  if (params.hasOwnProperty("jam")) patch.jam = params.jam;
  if (params.hasOwnProperty("group_id")) patch.group_id = params.group_id;
  if (params.hasOwnProperty("acara")) patch.acara = params.acara;
  if (params.hasOwnProperty("materi")) patch.materi = params.materi;
  if (params.hasOwnProperty("status")) patch.status = params.status;
  if (params.hasOwnProperty("catatan")) patch.catatan = params.catatan;
  if (params.hasOwnProperty("kategori_target")) {
    patch.kategori_target = JSON.stringify(params.kategori_target || []);
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

/* -------------------------------------------------------------------------- */
/*                              Delete meeting                                */
/* -------------------------------------------------------------------------- */

function deleteMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var meetingsRepo = new SheetRepository_("meetings");
  var existing = meetingsRepo.findById("meeting_id", params.meeting_id);
  if (!existing) return fail_("Jadwal tidak ditemukan");

  // Hapus semua absensi terkait supaya tidak ada orphan record
  var attendanceRepo = new SheetRepository_("attendance");
  var related = attendanceRepo.findByField("meeting_id", params.meeting_id);

  var deletedAttendance = 0;
  related.forEach(function (a) {
    try {
      attendanceRepo.deleteById("attendance_id", a.attendance_id);
      deletedAttendance++;
    } catch (e) {
      Logger.log("Gagal hapus attendance " + a.attendance_id + ": " + e);
    }
  });

  // Hapus meeting
  meetingsRepo.deleteById("meeting_id", params.meeting_id);

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
  try {
    out.kategori_target = JSON.parse(m.kategori_target || "[]");
  } catch (e) {
    out.kategori_target = [];
  }
  return out;
}
