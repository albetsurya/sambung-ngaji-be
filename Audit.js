function getAuditLogs_(ctx, params) {
  var repo = new SheetRepository_("audit_logs");
  var all = repo.getAllRecent(AUDIT_LOG_RECENT_LIMIT);
  if (params.user_id)
    all = all.filter(function (l) {
      return l.user_id === params.user_id;
    });
  if (params.target_type)
    all = all.filter(function (l) {
      return l.target_type === params.target_type;
    });
  all.sort(function (a, b) {
    return new Date(b.timestamp) - new Date(a.timestamp);
  });
  var limit = params.limit ? Number(params.limit) : 200;
  return ok_(
    all.slice(0, limit).map(function (l) {
      var c = Object.assign({}, l);
      delete c._row;
      return c;
    }),
  );
}

function getAiUsageToday_(ctx) {
  var today = formatDate(nowIso_());
  var repo = new SheetRepository_("ai_usage");
  return repo.find(function (u) {
    return (
      u.user_id === ctx.user.user_id && String(u.timestamp).indexOf(today) === 0
    );
  });
}

function checkAiQuota_(ctx) {
  var limit = AI_DAILY_LIMIT[ctx.user.role];
  if (!limit) return null;

  var todayUsage = getAiUsageToday_(ctx);
  if (todayUsage.length >= limit) {
    return fail_("Batas harian tercapai (" + limit + "x). Coba lagi besok.");
  }
  return null;
}

function logAiUsage_(ctx, provider, inputTokens, outputTokens) {
  try {
    var repo = new SheetRepository_("ai_usage");
    var total = (inputTokens || 0) + (outputTokens || 0);
    repo.insert({
      usage_id: generateUsageId(),
      user_id: ctx.user.user_id,
      user_nama: ctx.user.nama,
      role: ctx.user.role,
      provider: provider || "",
      input_tokens: inputTokens || 0,
      output_tokens: outputTokens || 0,
      total_tokens: total,
      timestamp: nowIso_(),
    });
  } catch (e) {
    Logger.log("Gagal log AI usage: " + e);
  }
}
