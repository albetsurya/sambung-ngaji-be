/**
 * Audit.js
 *
 * Berisi:
 * - Audit log reader (getAuditLogs_)
 * - AI usage logging (logAiUsage_, getAiUsageToday_, getAiUsageStats_)
 * - AI quota enforcement (checkAiQuota_, incrementAiQuota_, getAiQuotaUsed_)
 * - AI quota cleanup trigger (cleanupOldAiQuota_)
 *
 * Perubahan (Fix FASE 1 — Kuota AI):
 * - checkAiQuota_ dan incrementAiQuota_ sekarang pakai PropertiesService
 *   (persist akurat 24 jam, bukan CacheService yang max 6 jam)
 * - Tambah helper _aiQuotaKey_, getAiQuotaUsed_, cleanupOldAiQuota_
 */

/* -------------------------------------------------------------------------- */
/*                              AUDIT LOGS                                    */
/* -------------------------------------------------------------------------- */

function getAuditLogs_(ctx, params) {
  var repo = new SheetRepository_("audit_logs");
  var all = repo.getAllRecent(AUDIT_LOG_RECENT_LIMIT);

  /* -------- Build users map untuk fallback enrichment -------- */
  var usersRepo = new SheetRepository_("users");
  var users = usersRepo.getAll();
  var usersById = {};
  users.forEach(function (u) {
    usersById[u.user_id] = u;
  });

  /* -------- Filter -------- */
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

      // Fallback: kalau user_nama kosong (log lama), enrich dari users map
      if (!c.user_nama && c.user_id && usersById[c.user_id]) {
        c.user_nama = usersById[c.user_id].nama || "";
      }

      delete c._row;
      return c;
    }),
  );
}

/* -------------------------------------------------------------------------- */
/*                            AI USAGE LOGGING                                */
/* -------------------------------------------------------------------------- */

function getAiUsageToday_(ctx) {
  var today = formatDate(nowIso_());
  var repo = new SheetRepository_("ai_usage");
  return repo.find(function (u) {
    return (
      u.user_id === ctx.user.user_id && String(u.timestamp).indexOf(today) === 0
    );
  });
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

/* -------------------------------------------------------------------------- */
/*                        AI QUOTA ENFORCEMENT                                */
/* -------------------------------------------------------------------------- */
/*
 * Kuota disimpan di PropertiesService (persist, tanpa TTL).
 * Key format: aiquota:<user_id>:<YYYY-MM-DD>
 * Cleanup: trigger harian (Setup.js → setupDailyAiQuotaCleanupTrigger)
 */

function _aiQuotaKey_(userId, dateStr) {
  return "aiquota:" + userId + ":" + (dateStr || formatDate(nowIso_()));
}

function checkAiQuota_(ctx) {
  var limit = AI_DAILY_LIMIT[ctx.user.role];
  if (!limit) return null;

  var key = _aiQuotaKey_(ctx.user.user_id);
  var props = PropertiesService.getScriptProperties();
  var count = Number(props.getProperty(key) || 0);

  if (count >= limit) {
    return fail_(
      "Kuota AI chat harian sudah habis (" +
        count +
        "/" +
        limit +
        "). Coba lagi besok.",
    );
  }
  return null;
}

function incrementAiQuota_(ctx) {
  var key = _aiQuotaKey_(ctx.user.user_id);
  var props = PropertiesService.getScriptProperties();
  var count = Number(props.getProperty(key) || 0);
  props.setProperty(key, String(count + 1));
}

/**
 * Utility untuk debugging: cek berapa kuota yang sudah dipakai user hari ini.
 */
function getAiQuotaUsed_(ctx) {
  var key = _aiQuotaKey_(ctx.user.user_id);
  return Number(PropertiesService.getScriptProperties().getProperty(key) || 0);
}

/**
 * Cleanup key kuota AI dari hari-hari sebelumnya.
 * Jalankan via trigger harian (jam 2 pagi) atau manual dari editor GAS.
 */
function cleanupOldAiQuota_() {
  var props = PropertiesService.getScriptProperties();
  var keys = props.getKeys();
  var today = formatDate(nowIso_());
  var todaySuffix = ":" + today;
  var removed = 0;

  keys.forEach(function (k) {
    if (k.indexOf("aiquota:") === 0 && k.indexOf(todaySuffix) === -1) {
      props.deleteProperty(k);
      removed++;
    }
  });

  Logger.log("Cleaned " + removed + " old AI quota keys (kept " + today + ")");
  return removed;
}

/* -------------------------------------------------------------------------- */
/*                          AI USAGE STATS (ADMIN)                            */
/* -------------------------------------------------------------------------- */

function getAiUsageStats_(ctx, params) {
  if (ctx.user.role !== ROLES.SUPER_ADMIN && ctx.user.role !== ROLES.ADMIN) {
    return fail_("Hanya admin yang bisa akses monitoring AI");
  }

  var repo = new SheetRepository_("ai_usage");
  var all = repo.getAll();

  var today = formatDate(nowIso_());
  var monthStart = today.substring(0, 7) + "-01";

  var todayRows = [];
  var monthRows = [];
  var byUser = {};
  var byProvider = {};
  var byRole = {};

  all.forEach(function (u) {
    var ts = String(u.timestamp || "");
    var tsDate = ts.substring(0, 10);
    var tsMonth = ts.substring(0, 7);

    if (tsDate === today) todayRows.push(u);
    if (tsDate >= monthStart) monthRows.push(u);

    if (tsDate >= monthStart) {
      var userId = u.user_id || "unknown";
      if (!byUser[userId]) {
        byUser[userId] = {
          user_id: userId,
          user_nama: u.user_nama || "Unknown",
          role: u.role || "",
          chat_count: 0,
          total_tokens: 0,
        };
      }
      byUser[userId].chat_count++;
      byUser[userId].total_tokens += Number(u.total_tokens) || 0;

      var provider = u.provider || "unknown";
      if (!byProvider[provider]) {
        byProvider[provider] = {
          provider: provider,
          chat_count: 0,
          total_tokens: 0,
        };
      }
      byProvider[provider].chat_count++;
      byProvider[provider].total_tokens += Number(u.total_tokens) || 0;

      var role = u.role || "unknown";
      if (!byRole[role]) {
        byRole[role] = { role: role, chat_count: 0, total_tokens: 0 };
      }
      byRole[role].chat_count++;
      byRole[role].total_tokens += Number(u.total_tokens) || 0;
    }
  });

  var todayTokens = todayRows.reduce(function (sum, u) {
    return sum + (Number(u.total_tokens) || 0);
  }, 0);

  var monthTokens = monthRows.reduce(function (sum, u) {
    return sum + (Number(u.total_tokens) || 0);
  }, 0);

  var topUsers = Object.values(byUser)
    .sort(function (a, b) {
      return b.chat_count - a.chat_count;
    })
    .slice(0, 10);

  return ok_({
    today: {
      chat_count: todayRows.length,
      total_tokens: todayTokens,
    },
    month: {
      chat_count: monthRows.length,
      total_tokens: monthTokens,
    },
    by_provider: Object.values(byProvider).sort(function (a, b) {
      return b.chat_count - a.chat_count;
    }),
    by_role: Object.values(byRole).sort(function (a, b) {
      return b.chat_count - a.chat_count;
    }),
    top_users: topUsers,
  });
}
