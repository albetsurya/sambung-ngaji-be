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
  var toUpdate = [];

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

function getAttendanceByMeeting_(meetingId) {
  if (!meetingId) return [];

  var cacheKey = ATTENDANCE_CACHE_PREFIX + meetingId;
  var cache = CacheService.getScriptCache();

  var cached = cache.get(cacheKey);
  if (cached) {
    try {
      return JSON.parse(cached);
    } catch (e) {}
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
  } catch (e) {}

  return cleaned;
}

function invalidateAttendanceCache_(meetingId) {
  if (!meetingId) return;
  try {
    CacheService.getScriptCache().remove(ATTENDANCE_CACHE_PREFIX + meetingId);
  } catch (e) {}
}
