function getMonitoring_(ctx, params) {
  if (!params.member_id) return fail_("member_id wajib diisi");
  if (ctx.user.role === ROLES.TIM_PNKB) {
    var check = getMemberDetail_(ctx, { member_id: params.member_id });
    if (!check.success) return check;
  }

  var repo = new SheetRepository_("monitoring");
  var rows = repo.findByField("member_id", params.member_id);

  rows.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  var isPaged = String(params.paged) === "true";

  var limit =
    params.limit !== undefined ? Number(params.limit) : isPaged ? 20 : 20;
  if (isNaN(limit) || limit <= 0) limit = 20;
  if (limit > 100) limit = 100;

  var offset = params.offset !== undefined ? Number(params.offset) : 0;
  if (isNaN(offset) || offset < 0) offset = 0;

  var total = rows.length;
  var paged = rows.slice(offset, offset + limit);

  var items = paged.map(function (r) {
    var c = Object.assign({}, r);
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

function createMonitoring_(ctx, params) {
  if (!params.member_id || !params.status)
    return fail_("member_id dan status wajib diisi");
  if (Object.keys(MONITORING_STATUS).indexOf(params.status) === -1)
    return fail_("Status monitoring tidak valid");

  var repo = new SheetRepository_("monitoring");
  var now = nowIso_();
  var monitoringId = generateMonitoringId();
  var row = {
    monitoring_id: monitoringId,
    member_id: params.member_id,
    tanggal: params.tanggal ? formatDate(params.tanggal) : formatDate(now),
    jenis: params.jenis || "UMUM",
    status: params.status,
    catatan: params.catatan || "",
    tindak_lanjut: params.tindak_lanjut || "",
    created_by: ctx.user.user_id,
    created_at: now,
    updated_at: now,
  };
  repo.insert(row);

  var membersRepo = new SheetRepository_("members");
  membersRepo.updateById("member_id", params.member_id, {
    status_pembinaan: params.status,
    updated_at: now,
  });

  invalidateMyDashboardCache_(params.member_id);
  invalidateDashboardCache_();

  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_MONITORING",
    "MONITORING",
    monitoringId,
  );
  return ok_(row);
}

function updateMonitoring_(ctx, params) {
  if (!params.monitoring_id) return fail_("monitoring_id wajib diisi");
  var repo = new SheetRepository_("monitoring");
  var existing = repo.findById("monitoring_id", params.monitoring_id);
  if (!existing) return fail_("Data monitoring tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  ["catatan", "tindak_lanjut"].forEach(function (f) {
    if (params.hasOwnProperty(f)) patch[f] = params[f];
  });
  var updated = repo.updateById("monitoring_id", params.monitoring_id, patch);

  invalidateMyDashboardCache_(existing.member_id);
  invalidateDashboardCache_();

  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MONITORING",
    "MONITORING",
    params.monitoring_id,
  );
  return ok_(updated);
}
