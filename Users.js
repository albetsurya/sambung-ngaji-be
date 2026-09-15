function getUsers_(ctx, params) {
  var repo = new SheetRepository_("users");
  return ok_(
    repo.getAll().map(function (u) {
      return publicUser_(Object.assign({}, u));
    }),
  );
}

function getUserDetail_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");

  var repo = new SheetRepository_("users");
  var user = repo.findById("user_id", params.user_id);
  if (!user) return fail_("User tidak ditemukan");

  return ok_(publicUser_(user));
}

function createUser_(ctx, params) {
  if (!params.username || !params.password || !params.role)
    return fail_("username, password, role wajib diisi");
  if (!ROLES[params.role]) return fail_("Role tidak valid");

  if (!params.member_id) {
    return fail_(
      "member_id wajib diisi. Setiap user harus terhubung ke jamaah.",
    );
  }

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", params.member_id);
  if (!member) {
    return fail_("Jamaah dengan member_id tersebut tidak ditemukan");
  }

  var repo = new SheetRepository_("users");
  var existingUserForMember = repo.find(function (u) {
    return u.member_id === params.member_id && toBool_(u.status_aktif);
  });
  if (existingUserForMember.length > 0) {
    return fail_("Jamaah ini sudah punya akun aktif");
  }

  if (repo.findById("username", params.username))
    return fail_("Username sudah digunakan");

  var now = nowIso_();
  var userId = generateUserId();
  var row = {
    user_id: userId,
    username: params.username,
    password_hash: hashPassword_(params.password),
    nama: params.nama || params.username,
    role: params.role,
    member_id: params.member_id || "",
    status_aktif: true,
    created_at: now,
    updated_at: now,
    last_login_at: "",
  };
  repo.insert(row);

  invalidateDashboardCache_();
  writeAuditLog_(ctx.user.user_id, "CREATE_USER", "USER", userId);
  return ok_(publicUser_(row));
}

function updateUser_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");
  var repo = new SheetRepository_("users");
  var existing = repo.findById("user_id", params.user_id);
  if (!existing) return fail_("User tidak ditemukan");

  if (params.role && !ROLES[params.role]) {
    return fail_("Role tidak valid");
  }

  if (
    params.role &&
    params.user_id === ctx.user.user_id &&
    params.role !== ROLES.SUPER_ADMIN &&
    ctx.user.role === ROLES.SUPER_ADMIN
  ) {
    return fail_("Tidak bisa mengubah role diri sendiri dari SUPER_ADMIN");
  }

  if (
    params.role &&
    existing.role === ROLES.SUPER_ADMIN &&
    params.role !== ROLES.SUPER_ADMIN
  ) {
    var superAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (superAdmins.length === 0) {
      return fail_("Tidak bisa mengubah role SUPER_ADMIN terakhir");
    }
  }

  if (
    params.hasOwnProperty("status_aktif") &&
    !toBool_(params.status_aktif) &&
    existing.role === ROLES.SUPER_ADMIN
  ) {
    var activeSuperAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (activeSuperAdmins.length === 0) {
      return fail_("Tidak bisa menonaktifkan SUPER_ADMIN terakhir");
    }
  }

  if (
    params.hasOwnProperty("member_id") &&
    params.member_id !== existing.member_id
  ) {
    return fail_("member_id tidak bisa diubah setelah user dibuat");
  }

  var patch = { updated_at: nowIso_() };
  ["nama", "role", "status_aktif"].forEach(function (f) {
    if (params.hasOwnProperty(f)) patch[f] = params[f];
  });
  if (params.password) patch.password_hash = hashPassword_(params.password);

  var updated = repo.updateById("user_id", params.user_id, patch);
  invalidateDashboardCache_();
  writeAuditLog_(ctx.user.user_id, "UPDATE_USER", "USER", params.user_id);

  var shouldInvalidate = false;
  if (params.hasOwnProperty("role") && params.role !== existing.role)
    shouldInvalidate = true;
  if (
    params.hasOwnProperty("status_aktif") &&
    toBool_(params.status_aktif) !== toBool_(existing.status_aktif)
  )
    shouldInvalidate = true;
  if (params.password) shouldInvalidate = true;

  if (shouldInvalidate) {
    _invalidateUserSessions_(params.user_id);
  }

  return ok_(publicUser_(updated));
}

function updateUserRole_(ctx, params) {
  if (!params.user_id) return fail_("user_id wajib diisi");
  if (!params.role) return fail_("role wajib diisi");
  if (!ROLES[params.role]) return fail_("Role tidak valid");

  var repo = new SheetRepository_("users");
  var existing = repo.findById("user_id", params.user_id);
  if (!existing) return fail_("User tidak ditemukan");

  if (
    params.user_id === ctx.user.user_id &&
    params.role !== ROLES.SUPER_ADMIN &&
    ctx.user.role === ROLES.SUPER_ADMIN
  ) {
    return fail_("Tidak bisa mengubah role diri sendiri dari SUPER_ADMIN");
  }

  if (
    existing.role === ROLES.SUPER_ADMIN &&
    params.role !== ROLES.SUPER_ADMIN
  ) {
    var superAdmins = repo.find(function (u) {
      return (
        u.role === ROLES.SUPER_ADMIN &&
        toBool_(u.status_aktif) &&
        u.user_id !== params.user_id
      );
    });
    if (superAdmins.length === 0) {
      return fail_("Tidak bisa mengubah role SUPER_ADMIN terakhir");
    }
  }

  var now = nowIso_();
  var updated = repo.updateById("user_id", params.user_id, {
    role: params.role,
    updated_at: now,
  });

  invalidateDashboardCache_();

  writeAuditLog_(ctx.user.user_id, "UPDATE_USER_ROLE", "USER", params.user_id);

  if (params.role !== existing.role) {
    _invalidateUserSessions_(params.user_id);
  }

  return ok_(publicUser_(updated));
}

function getMemberUserStatus_(memberId) {
  var usersRepo = new SheetRepository_("users");
  var user = usersRepo.find(function (u) {
    return u.member_id === memberId && toBool_(u.status_aktif);
  })[0];

  if (!user) {
    return {
      has_user: false,
      user: null,
    };
  }

  return {
    has_user: true,
    user: publicUser_(user),
  };
}

function auditUsersWithoutMember() {
  var usersRepo = new SheetRepository_("users");
  var users = usersRepo.getAll();
  var orphan = users.filter(function (u) {
    return !u.member_id;
  });

  Logger.log("=== AUDIT USER TANPA member_id ===");
  Logger.log("Total user: " + users.length);
  Logger.log("Tanpa member_id: " + orphan.length);
  Logger.log("");

  orphan.forEach(function (u) {
    Logger.log(
      "  - " +
        u.user_id +
        " | @" +
        u.username +
        " | " +
        u.nama +
        " | role: " +
        u.role,
    );
  });

  if (orphan.length === 0) {
    Logger.log("Semua user sudah punya member_id.");
  }

  return {
    total: users.length,
    orphan_count: orphan.length,
    orphan: orphan.map(function (u) {
      return {
        user_id: u.user_id,
        username: u.username,
        nama: u.nama,
        role: u.role,
      };
    }),
  };
}
