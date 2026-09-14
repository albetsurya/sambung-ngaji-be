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

function deleteMeeting_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var meetingsRepo = new SheetRepository_("meetings");
  var existing = meetingsRepo.findById("meeting_id", params.meeting_id);
  if (!existing) return fail_("Jadwal tidak ditemukan");

  // Hapus juga absensi terkait supaya tidak ada orphan
  var attendanceRepo = new SheetRepository_("attendance");
  var relatedAttendance = attendanceRepo.findByField(
    "meeting_id",
    params.meeting_id,
  );
  relatedAttendance.forEach(function (a) {
    attendanceRepo.deleteById("attendance_id", a.attendance_id);
  });

  meetingsRepo.deleteById("meeting_id", params.meeting_id);

  writeAuditLog_(
    ctx.user.user_id,
    "DELETE_MEETING",
    "MEETING",
    params.meeting_id,
  );

  return ok_({
    meeting_id: params.meeting_id,
    deleted_attendance: relatedAttendance.length,
  });
}
