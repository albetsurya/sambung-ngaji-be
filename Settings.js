function getSettings_(ctx, params) {
  var repo = new SheetRepository_("settings");
  var all = repo.getAll();
  var out = {};
  all.forEach(function (s) {
    try {
      out[s.key] = JSON.parse(s.value);
    } catch (e) {
      out[s.key] = s.value;
    }
  });
  return ok_(out);
}

function updateSettings_(ctx, params) {
  if (!params.key) return fail_("key wajib diisi");
  var repo = new SheetRepository_("settings");
  var existing = repo.findById("key", params.key);
  var value =
    typeof params.value === "string"
      ? params.value
      : JSON.stringify(params.value);
  if (existing) {
    repo.updateById("key", params.key, { value: value, updated_at: nowIso_() });
  } else {
    repo.insert({ key: params.key, value: value, updated_at: nowIso_() });
  }
  writeAuditLog_(ctx.user.user_id, "UPDATE_SETTINGS", "SETTINGS", params.key);
  return ok_({ key: params.key, value: JSON.parse(value) });
}
