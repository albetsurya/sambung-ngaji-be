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