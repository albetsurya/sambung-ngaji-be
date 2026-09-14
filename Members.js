function enrichMember_(member) {
  var out = Object.assign({}, member);
  out.kategori = getMemberCategory(member);
  out.usia = getMemberAge(member.tanggal_lahir);
  delete out._row;
  return out;
}

function filterMemberFieldsByRole_(member, role) {
  var allowed = FIELD_VISIBILITY.UMUM.concat(["kategori", "usia", "no_wa"]);
  if (role === "SUPER_ADMIN" || role === "ADMIN") {
    allowed = allowed.concat(
      FIELD_VISIBILITY.INTERNAL,
      FIELD_VISIBILITY.PNKB,
      FIELD_VISIBILITY.PENDIDIKAN,
    );
  } else if (role === "PENGAWAS") {
    allowed = allowed.concat(
      FIELD_VISIBILITY.INTERNAL,
      FIELD_VISIBILITY.PNKB,
      FIELD_VISIBILITY.PENDIDIKAN,
    );
  } else if (role === "TIM_PNKB") {
    allowed = allowed.concat([
      "pekerjaan",
      "is_muballigh",
      "is_kerja",
      "is_nikah",
      "tinggi_badan",
      "berat_badan",
      "hobi",
      "status_pembinaan",
      "alamat_rumah",
    ]);
  } else if (role === "TIM_ABSENSI") {
    allowed = allowed.concat(["kelompok"]);
  }
  var out = {};
  allowed.forEach(function (f) {
    if (member.hasOwnProperty(f)) out[f] = member[f];
  });
  return out;
}

function getMembers_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();

  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (m) {
      return toBool_(m.status_aktif);
    });
  }
  if (params.kelompok)
    all = all.filter(function (m) {
      return m.kelompok === params.kelompok;
    });
  if (params.jenis_kelamin)
    all = all.filter(function (m) {
      return m.jenis_kelamin === params.jenis_kelamin;
    });
  if (params.desa)
    all = all.filter(function (m) {
      return m.desa === params.desa;
    });
  if (params.search) {
    var q = String(params.search).toLowerCase();
    all = all.filter(function (m) {
      return (
        String(m.nama_lengkap).toLowerCase().indexOf(q) !== -1 ||
        String(m.nama_panggilan).toLowerCase().indexOf(q) !== -1
      );
    });
  }

  var enriched = all.map(enrichMember_);
  if (params.kategori) {
    enriched = enriched.filter(function (m) {
      return m.kategori === params.kategori;
    });
  }

  var role = ctx.user.role;
  var filtered = enriched.map(function (m) {
    return filterMemberFieldsByRole_(m, role);
  });
  return ok_(filtered);
}

function getMembersPaged_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();

  if (String(params.includeInactive) !== "true") {
    all = all.filter(function (m) {
      return toBool_(m.status_aktif);
    });
  }
  if (params.kelompok)
    all = all.filter(function (m) {
      return m.kelompok === params.kelompok;
    });
  if (params.jenis_kelamin)
    all = all.filter(function (m) {
      return m.jenis_kelamin === params.jenis_kelamin;
    });
  if (params.desa)
    all = all.filter(function (m) {
      return m.desa === params.desa;
    });
  if (params.search) {
    var q = String(params.search).toLowerCase();
    all = all.filter(function (m) {
      return (
        String(m.nama_lengkap).toLowerCase().indexOf(q) !== -1 ||
        String(m.nama_panggilan).toLowerCase().indexOf(q) !== -1
      );
    });
  }

  var enriched = all.map(enrichMember_);
  if (params.kategori) {
    enriched = enriched.filter(function (m) {
      return m.kategori === params.kategori;
    });
  }

  enriched.sort(function (a, b) {
    return String(a.nama_lengkap).localeCompare(String(b.nama_lengkap));
  });

  var total = enriched.length;
  var limit = params.limit ? Number(params.limit) : 0;
  var offset = params.offset ? Number(params.offset) : 0;

  var paged = enriched;
  if (limit > 0) {
    paged = enriched.slice(offset, offset + limit);
  }

  var role = ctx.user.role;
  var filtered = paged.map(function (m) {
    return filterMemberFieldsByRole_(m, role);
  });

  return ok_({
    items: filtered,
    total: total,
    limit: limit,
    offset: offset,
    has_more: limit > 0 ? offset + paged.length < total : false,
  });
}

function getPNKBMembers_(ctx, params) {
  return getMembers_(
    ctx,
    Object.assign({}, params, { kategori: MEMBER_CATEGORY.PRA_NIKAH }),
  );
}

function getPNKBMembersPaged_(ctx, params) {
  return getMembersPaged_(
    ctx,
    Object.assign({}, params, { kategori: MEMBER_CATEGORY.PRA_NIKAH }),
  );
}

function getAttendanceMembers_(ctx, params) {
  return getMembers_(ctx, params);
}

function getMemberDetail_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", memberId);
  if (!member) return fail_("Jamaah tidak ditemukan");

  if (ctx.user.role === ROLES.TIM_PNKB) {
    var cat = getMemberCategory(member);
    if (cat !== MEMBER_CATEGORY.PRA_NIKAH)
      return fail_("Tidak memiliki akses ke data jamaah ini");
  }

  var enriched = enrichMember_(member);
  var filtered = filterMemberFieldsByRole_(enriched, ctx.user.role);
  return ok_(filtered);
}

function createMember_(ctx, params) {
  var membersRepo = new SheetRepository_("members");
  var now = nowIso_();
  var memberId = generateMemberId();
  var member = {
    member_id: memberId,
    nama_lengkap: params.nama_lengkap || "",
    nama_panggilan: params.nama_panggilan || "",
    jenis_kelamin: params.jenis_kelamin || "",
    tempat_lahir: params.tempat_lahir || "",
    tanggal_lahir: params.tanggal_lahir || "",
    kelompok: params.kelompok || "",
    desa: params.desa || "",
    daerah: params.daerah || "",
    alamat_rumah: params.alamat_rumah || "",
    no_wa: params.no_wa ? normalizePhoneNumber(params.no_wa) : "",
    is_muballigh: toBool_(params.is_muballigh),
    is_kerja: toBool_(params.is_kerja),
    is_nikah: toBool_(params.is_nikah),
    tinggi_badan: params.tinggi_badan || "",
    berat_badan: params.berat_badan || "",
    hobi: params.hobi || "",
    pekerjaan: params.pekerjaan || "",
    foto_url: params.foto_url || "",
    status_pembinaan: params.status_pembinaan || MONITORING_STATUS.AKTIF,
    status_aktif: true,
    tanggal_masuk: params.tanggal_masuk || formatDate(now),
    tanggal_keluar: "",
    jenjang_pendidikan: params.jenjang_pendidikan || "",
    sekolah: params.sekolah || "",
    jurusan: params.jurusan || "",
    tahun_mulai_pendidikan: params.tahun_mulai_pendidikan || "",
    tahun_selesai_pendidikan: params.tahun_selesai_pendidikan || "",
    created_at: now,
    updated_at: now,
  };
  membersRepo.insert(member);
  writeAuditLog_(ctx.user.user_id, "CREATE_MEMBER", "MEMBER", memberId);
  return ok_(member);
}

var MEMBER_EDITABLE_FIELDS = [
  "nama_lengkap",
  "nama_panggilan",
  "jenis_kelamin",
  "tempat_lahir",
  "tanggal_lahir",
  "kelompok",
  "desa",
  "daerah",
  "alamat_rumah",
  "no_wa",
  "is_muballigh",
  "is_kerja",
  "is_nikah",
  "tinggi_badan",
  "berat_badan",
  "hobi",
  "pekerjaan",
  "foto_url",
  "status_pembinaan",
  "tanggal_masuk",
  "tanggal_keluar",
  "jenjang_pendidikan",
  "sekolah",
  "jurusan",
  "tahun_mulai_pendidikan",
  "tahun_selesai_pendidikan",
];

function updateMember_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");
  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", memberId);
  if (!existing) return fail_("Jamaah tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  MEMBER_EDITABLE_FIELDS.forEach(function (f) {
    if (params.hasOwnProperty(f)) {
      patch[f] =
        f === "no_wa" && params[f]
          ? normalizePhoneNumber(params[f])
          : params[f];
    }
  });
  var updated = membersRepo.updateById("member_id", memberId, patch);
  writeAuditLog_(ctx.user.user_id, "UPDATE_MEMBER", "MEMBER", memberId);
  return ok_(updated);
}

function deactivateMember_(ctx, params) {
  var memberId = params.member_id;
  if (!memberId) return fail_("member_id wajib diisi");
  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", memberId);
  if (!existing) return fail_("Jamaah tidak ditemukan");
  var updated = membersRepo.updateById("member_id", memberId, {
    status_aktif: false,
    tanggal_keluar: formatDate(nowIso_()),
    updated_at: nowIso_(),
  });
  writeAuditLog_(ctx.user.user_id, "DEACTIVATE_MEMBER", "MEMBER", memberId);
  return ok_(updated);
}

var MEMBER_SELF_EDITABLE_FIELDS = [
  "nama_panggilan",
  "no_wa",
  "alamat_rumah",
  "desa",
  "daerah",
  "pekerjaan",
  "hobi",
  "foto_url",
  "tinggi_badan",
  "berat_badan",
  "is_kerja",
  "is_nikah",
  "jenjang_pendidikan",
  "sekolah",
  "jurusan",
  "tahun_mulai_pendidikan",
  "tahun_selesai_pendidikan",
];

function getMyProfile_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  if (!member) return fail_("Data jamaah tidak ditemukan");

  var enriched = enrichMember_(member);
  var allowed = [
    "member_id",
    "nama_lengkap",
    "nama_panggilan",
    "jenis_kelamin",
    "tempat_lahir",
    "tanggal_lahir",
    "foto_url",
    "kelompok",
    "desa",
    "daerah",
    "alamat_rumah",
    "no_wa",
    "pekerjaan",
    "hobi",
    "status_pembinaan",
    "status_aktif",
    "tanggal_masuk",
    "jenjang_pendidikan",
    "sekolah",
    "jurusan",
    "tahun_mulai_pendidikan",
    "tahun_selesai_pendidikan",
    "kategori",
    "usia",
  ];
  var out = {};
  allowed.forEach(function (f) {
    if (enriched.hasOwnProperty(f)) out[f] = enriched[f];
  });
  return ok_(out);
}

function updateMyProfile_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var existing = membersRepo.findById("member_id", ctx.user.member_id);
  if (!existing) return fail_("Data jamaah tidak ditemukan");

  var patch = { updated_at: nowIso_() };
  MEMBER_SELF_EDITABLE_FIELDS.forEach(function (f) {
    if (params.hasOwnProperty(f)) {
      if (f === "no_wa" && params[f]) {
        patch[f] = normalizePhoneNumber(params[f]);
      } else {
        patch[f] = params[f];
      }
    }
  });

  var changedFields = Object.keys(patch).filter(function (k) {
    return k !== "updated_at";
  });
  if (changedFields.length === 0) {
    return fail_("Tidak ada perubahan.");
  }

  var updated = membersRepo.updateById("member_id", ctx.user.member_id, patch);

  writeAuditLog_(
    ctx.user.user_id,
    "UPDATE_MY_PROFILE",
    "MEMBER",
    ctx.user.member_id,
  );

  return ok_(enrichMember_(updated));
}

function getMyAttendance_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var repo = new SheetRepository_("attendance");
  var rows = repo.findByField("member_id", ctx.user.member_id);

  var meetingsRepo = new SheetRepository_("meetings");
  var allMeetings = meetingsRepo.getAll();
  var meetingsById = buildIndexOne_(allMeetings, "meeting_id");

  var enriched = rows.map(function (a) {
    var m = meetingsById[a.meeting_id] || {};
    return {
      attendance_id: a.attendance_id,
      meeting_id: a.meeting_id,
      status: a.status,
      catatan: a.catatan || "",
      tanggal: m.tanggal || "",
      hari: m.hari || "",
      acara: m.acara || "",
      jam: m.jam || "",
      created_at: a.created_at,
    };
  });

  enriched.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  return ok_(enriched);
}

function getMyMonitoring_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var repo = new SheetRepository_("monitoring");
  var rows = repo.findByField("member_id", ctx.user.member_id);

  rows.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  return ok_(
    rows.map(function (r) {
      var c = Object.assign({}, r);
      delete c._row;
      return c;
    }),
  );
}

function getUpcomingMeetings_(ctx, params) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  var memberCategory = member ? getMemberCategory(member) : null;

  var meetingsRepo = new SheetRepository_("meetings");
  var today = formatDate(nowIso_());
  var limit = params.limit ? Number(params.limit) : 10;

  var upcoming = meetingsRepo
    .getAll()
    .filter(function (m) {
      if (formatDate(m.tanggal) < today) return false;
      var targets = parseKategoriTarget_(m.kategori_target);
      if (targets.length === 0) return true;
      return memberCategory && targets.indexOf(memberCategory) !== -1;
    })
    .sort(function (a, b) {
      return parseDate_(a.tanggal) - parseDate_(b.tanggal);
    })
    .slice(0, limit);

  return ok_(upcoming.map(publicMeeting_));
}
