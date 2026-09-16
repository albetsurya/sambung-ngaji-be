function getGroups_(ctx, params) {
  var repo = new SheetRepository_("groups");
  var all = repo.getAll();
  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (g) {
      return toBool_(g.status_aktif);
    });
  }
  return ok_(
    all.map(function (g) {
      var c = Object.assign({}, g);
      delete c._row;
      return c;
    }),
  );
}

function saveGroup_(ctx, params) {
  var repo = new SheetRepository_("groups");
  var now = nowIso_();

  if (params.group_id) {
    var existing = repo.findById("group_id", params.group_id);
    if (!existing) return fail_("Kelompok tidak ditemukan");
    var patch = { updated_at: now };
    [
      "group_code",
      "group_name",
      "pembina",
      "penandatangan",
      "jadwal",
      "status_aktif",
    ].forEach(function (f) {
      if (params.hasOwnProperty(f)) patch[f] = params[f];
    });
    var updated = repo.updateById("group_id", params.group_id, patch);
    invalidateDashboardCache_();
    writeAuditLog_(ctx.user.user_id, "UPDATE_GROUP", "GROUP", params.group_id);
    return ok_(updated);
  }

  var groupId = generateGroupId();
  var row = {
    group_id: groupId,
    group_code: params.group_code || "",
    group_name: params.group_name || "",
    pembina: params.pembina || "",
    penandatangan: params.penandatangan || "",
    jadwal: params.jadwal || "",
    status_aktif: true,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);
  invalidateDashboardCache_();
  writeAuditLog_(ctx.user.user_id, "CREATE_GROUP", "GROUP", groupId);
  return ok_(row);
}
