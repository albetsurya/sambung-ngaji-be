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

  usersRepo.updateById("user_id", user.user_id, {
    last_login_at: now.toISOString(),
  });
  writeAuditLog_(user.user_id, "LOGIN", "USER", user.user_id);
  return ok_({ token: token, user: publicUser_(user) });
}

function logout_(ctx) {
  if (!ctx || !ctx.token) return ok_(null);
  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", ctx.token);
  if (session)
    sessionsRepo.updateById("token", ctx.token, { expires_at: nowIso_() });
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
  var sessionsRepo = new SheetRepository_("sessions");
  var session = sessionsRepo.findById("token", token);
  if (!session) return null;
  if (new Date(session.expires_at).getTime() < Date.now()) return null;

  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.findById("user_id", session.user_id);
  if (!user || !toBool_(user.status_aktif)) return null;
  return { token: token, user: user };
}

function checkPermission_(user, action) {
  var perms = ROLE_PERMISSIONS[user.role];
  if (!perms) return false;
  if (perms.indexOf("*") !== -1) return true;
  return perms.indexOf(action) !== -1;
}

function writeAuditLog_(userId, action, targetType, targetId) {
  try {
    var logsRepo = new SheetRepository_("audit_logs");
    logsRepo.insert({
      log_id: generateLogId(),
      user_id: userId || "",
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
