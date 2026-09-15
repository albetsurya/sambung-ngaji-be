function _sessionCacheKey_(token) {
  return SESSION_CACHE_PREFIX + token;
}

function _userSessionKey_(userId) {
  return USER_SESSION_PREFIX + userId;
}

function _buildSessionPayload_(user) {
  return {
    user_id: user.user_id || "",
    username: user.username || "",
    nama: user.nama || "",
    role: user.role || "",
    member_id: user.member_id || "",
    jenis_kelamin: user.jenis_kelamin || "",
    status_aktif: user.status_aktif,
  };
}

function _cacheSession_(token, payload) {
  if (!token || !payload || !payload.user_id) return;
  try {
    var cache = CacheService.getScriptCache();
    cache.put(
      _sessionCacheKey_(token),
      JSON.stringify(payload),
      SESSION_CACHE_TTL,
    );
    var userKey = _userSessionKey_(payload.user_id);
    var existing = cache.get(userKey);
    var tokens = [];
    if (existing) {
      try {
        tokens = JSON.parse(existing);
      } catch (e) {
        tokens = [];
      }
    }
    if (tokens.indexOf(token) === -1) tokens.push(token);
    cache.put(userKey, JSON.stringify(tokens), SESSION_CACHE_TTL);
  } catch (e) {
    Logger.log("[session-cache] put error: " + e);
  }
}

function _getCachedSession_(token) {
  if (!token) return null;
  try {
    var cache = CacheService.getScriptCache();
    var raw = cache.get(_sessionCacheKey_(token));
    if (!raw) return null;
    return JSON.parse(raw);
  } catch (e) {
    Logger.log("[session-cache] get error: " + e);
    return null;
  }
}

function _removeCachedSession_(token) {
  if (!token) return;
  try {
    var cache = CacheService.getScriptCache();
    var payload = null;
    var raw = cache.get(_sessionCacheKey_(token));
    if (raw) {
      try {
        payload = JSON.parse(raw);
      } catch (e) {}
    }
    cache.remove(_sessionCacheKey_(token));
    if (payload && payload.user_id) {
      var userKey = _userSessionKey_(payload.user_id);
      var existing = cache.get(userKey);
      if (existing) {
        try {
          var tokens = JSON.parse(existing).filter(function (t) {
            return t !== token;
          });
          if (tokens.length) {
            cache.put(userKey, JSON.stringify(tokens), SESSION_CACHE_TTL);
          } else {
            cache.remove(userKey);
          }
        } catch (e) {}
      }
    }
  } catch (e) {
    Logger.log("[session-cache] remove error: " + e);
  }
}

function _invalidateUserSessions_(userId, exceptToken) {
  if (!userId) return;
  try {
    var cache = CacheService.getScriptCache();
    var userKey = _userSessionKey_(userId);
    var existing = cache.get(userKey);
    if (!existing) return;
    var tokens = [];
    try {
      tokens = JSON.parse(existing);
    } catch (e) {
      tokens = [];
    }
    tokens.forEach(function (t) {
      if (exceptToken && t === exceptToken) return;
      cache.remove(_sessionCacheKey_(t));
    });
    if (exceptToken && tokens.indexOf(exceptToken) !== -1) {
      cache.put(userKey, JSON.stringify([exceptToken]), SESSION_CACHE_TTL);
    } else {
      cache.remove(userKey);
    }
  } catch (e) {
    Logger.log("[session-cache] invalidateUser error: " + e);
  }
}

function login_(params) {
  var username = String(params.username || "").trim();
  var password = String(params.password || "");
  if (!username || !password) return fail_("Username dan password wajib diisi");

  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.findById("username", username);
  if (!user) return fail_("Username atau password salah");
  if (!toBool_(user.status_aktif)) return fail_("Akun tidak aktif");
  if (!verifyPassword_(password, user.password_hash))
    return fail_("Username atau password salah");

  var token = Utilities.getUuid();
  var now = new Date();
  var expires = new Date(now.getTime() + SESSION_TTL_HOURS * 3600 * 1000);
  var sessionsRepo = new SheetRepository_("sessions");
  sessionsRepo.insert({
    token: token,
    user_id: user.user_id,
    created_at: now.toISOString(),
    expires_at: expires.toISOString(),
  });

  maybeCleanupExpiredSessions_();

  usersRepo.updateById("user_id", user.user_id, {
    last_login_at: now.toISOString(),
  });

  _cacheSession_(token, _buildSessionPayload_(user));

  writeAuditLog_(user.user_id, "LOGIN", "USER", user.user_id);
  return ok_({ token: token, user: publicUser_(user) });
}

function maybeCleanupExpiredSessions_() {
  try {
    var sheet = new SheetRepository_("sessions")._sheet();
    var lastRow = sheet.getLastRow();
    if (lastRow < SESSION_MAX_ROWS_BEFORE_CLEANUP) return; // skip kalau masih kecil
    var repo = new SheetRepository_("sessions");
    var all = repo.getAll();
    var now = Date.now();
    var expiredRows = [];
    all.forEach(function (s) {
      if (new Date(s.expires_at).getTime() < now) expiredRows.push(s._row);
    });
    // hapus dari bawah ke atas biar index tidak geser
    expiredRows.sort(function (a, b) {
      return b - a;
    });
    expiredRows.forEach(function (r) {
      sheet.deleteRow(r);
    });
  } catch (e) {
    Logger.log("cleanup sessions error: " + e);
  }
}

function logout_(ctx) {
  if (!ctx || !ctx.token) return ok_(null);
  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", ctx.token);
  if (session)
    sessionsRepo.updateById("token", ctx.token, { expires_at: nowIso_() });
  _removeCachedSession_(ctx.token);
  if (ctx.user)
    writeAuditLog_(ctx.user.user_id, "LOGOUT", "USER", ctx.user.user_id);
  return ok_(null);
}

function publicUser_(user) {
  return {
    user_id: user.user_id,
    username: user.username,
    nama: user.nama,
    role: user.role,
    member_id: user.member_id || "",
    jenis_kelamin: user.jenis_kelamin || "",
  };
}

function validateSession_(token) {
  if (!token) return null;

  var cached = _getCachedSession_(token);
  if (cached) {
    if (!toBool_(cached.status_aktif)) {
      _removeCachedSession_(token);
      return null;
    }
    return { token: token, user: cached };
  }

  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", token);
  if (!session) return null;
  if (new Date(session.expires_at).getTime() < Date.now()) return null;

  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.findById("user_id", session.user_id);
  if (!user || !toBool_(user.status_aktif)) return null;

  _cacheSession_(token, _buildSessionPayload_(user));

  return { token: token, user: user };
}

function checkPermission_(user, action) {
  var perms = ROLE_PERMISSIONS[user.role];
  if (!perms) return false;
  if (perms.indexOf("*") !== -1) return true;
  return perms.indexOf(action) !== -1;
}

/**
 * Tulis audit log dengan snapshot nama user.
 * - `user_nama` disimpan supaya historical akurat (user bisa rename nanti).
 * - Kalau lookup gagal, `user_nama` diisi "" dan akan di-enrich saat read.
 */
function writeAuditLog_(userId, action, targetType, targetId) {
  try {
    var logsRepo = new SheetRepository_("audit_logs");

    // Lookup nama user dari cache (cheap — users sheet cached)
    var userName = "";
    if (userId) {
      try {
        var usersRepo = new SheetRepository_("users");
        var user = usersRepo.findById("user_id", userId);
        if (user && user.nama) userName = user.nama;
      } catch (e) {
        // ignore — fallback ke empty string
      }
    }

    logsRepo.insert({
      log_id: generateLogId(),
      user_id: userId || "",
      user_nama: userName,
      action: action,
      target_type: targetType || "",
      target_id: targetId || "",
      timestamp: nowIso_(),
    });
  } catch (e) {
    Logger.log("Gagal menulis audit log: " + e);
  }
}

function requireMemberLink_(ctx) {
  if (!ctx.user.member_id) {
    return fail_("Akun Anda belum terhubung ke data jamaah. Hubungi admin.");
  }
  return null;
}
