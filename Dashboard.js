var _DASHBOARD_CACHE_ = null;

function _getDashboardCache_() {
  if (!_DASHBOARD_CACHE_) {
    _DASHBOARD_CACHE_ = CacheService.getScriptCache();
  }
  return _DASHBOARD_CACHE_;
}

function getDashboard_(ctx, params) {
  var role = ctx.user.role;
  if (role === ROLES.TIM_PNKB) return getPNKBDashboard_(ctx);
  if (role === ROLES.TIM_ABSENSI) return getAbsensiDashboard_(ctx);
  return getGeneralDashboard_(ctx);
}

function _dashboardCacheKey_(type, suffix) {
  return "dash:" + type + (suffix ? ":" + suffix : "");
}

function _cachedDashboard_(cacheKey, buildFn) {
  var cache = _getDashboardCache_();

  try {
    var cached = cache.get(cacheKey);
    if (cached) {
      try {
        Logger.log("[dash-cache] HIT: " + cacheKey);
        return JSON.parse(cached);
      } catch (e) {
        Logger.log("[dash-cache] parse error: " + e);
      }
    }
  } catch (e) {
    Logger.log("[dash-cache] get error: " + e);
  }

  Logger.log("[dash-cache] MISS: " + cacheKey);
  var result = buildFn();

  try {
    var str = JSON.stringify(result);
    Logger.log("[dash-cache] build size: " + str.length + " bytes");
    if (str.length < 95000) {
      cache.put(cacheKey, str, 60);
      Logger.log("[dash-cache] PUT OK: " + cacheKey);
    } else {
      Logger.log("[dash-cache] SKIP (too big): " + str.length);
    }
  } catch (e) {
    Logger.log("[dash-cache] put error: " + e);
  }

  return result;
}

function invalidateDashboardCache_() {
  try {
    var cache = _getDashboardCache_();
    cache.remove("dash:general:SUPER_ADMIN");
    cache.remove("dash:general:ADMIN");
    cache.remove("dash:general:PENGAWAS");
    cache.remove("dash:general:TIM_PNKB");
    cache.remove("dash:general:TIM_ABSENSI");
    cache.remove("dash:pnkb");
    var today = formatDate(nowIso_());
    var yesterday = formatDate(new Date(Date.now() - 86400000));
    cache.remove("dash:absensi:" + today);
    cache.remove("dash:absensi:" + yesterday);
  } catch (e) {
    Logger.log("invalidateDashboardCache_ error: " + e);
  }
}

function invalidateMyDashboardCache_(memberId) {
  if (!memberId) return;
  try {
    _getDashboardCache_().remove("dash:my:" + memberId);
  } catch (e) {}
}

function getGeneralDashboard_(ctx) {
  var cacheKey = _dashboardCacheKey_("general", ctx.user.role);
  return ok_(
    _cachedDashboard_(cacheKey, function () {
      return _buildGeneralDashboard_(ctx);
    }),
  );
}

function _buildGeneralDashboard_(ctx) {
  var membersRepo = new SheetRepository_("members");
  var members = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var perKategori = {};
  Object.keys(MEMBER_CATEGORY).forEach(function (k) {
    perKategori[k] = 0;
  });
  members.forEach(function (m) {
    var cat = getMemberCategory(m);
    if (cat) perKategori[cat] = (perKategori[cat] || 0) + 1;
  });

  var meetingsRepo = new SheetRepository_("meetings");
  var meetings = meetingsRepo.getAll();
  var today = formatDate(nowIso_());
  var upcoming = meetings
    .filter(function (m) {
      return formatDate(m.tanggal) >= today;
    })
    .sort(function (a, b) {
      return new Date(a.tanggal) - new Date(b.tanggal);
    })[0];

  var attendanceRepo = new SheetRepository_("attendance");
  var recentAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var meetingsById = buildIndexOne_(meetings, "meeting_id");
  var attendanceByMember = buildIndexList_(recentAttendance, "member_id");

  var result = {
    total_jamaah: members.length,
    total_jamaah_aktif: members.length,
    per_kategori: perKategori,
    rata_rata_kehadiran: computeOverallAttendanceRate_(recentAttendance),
    pengajian_terdekat: upcoming
      ? {
          meeting_id: upcoming.meeting_id,
          tanggal: formatDate(upcoming.tanggal),
          hari: upcoming.hari,
          acara: upcoming.acara,
          kategori_target: parseKategoriTarget_(upcoming.kategori_target),
        }
      : null,
    jamaah_perlu_perhatian: buildAttentionList_(
      members,
      attendanceByMember,
      meetingsById,
    ).slice(0, 10),
    data_belum_lengkap: countIncompleteData_(members),
  };

  if (ctx.user.role === ROLES.SUPER_ADMIN || ctx.user.role === ROLES.ADMIN) {
    var usersRepo = new SheetRepository_("users");
    result.user_aktif = usersRepo.find(function (u) {
      return toBool_(u.status_aktif);
    }).length;
  }

  return result;
}

function getPNKBDashboard_(ctx) {
  var cacheKey = _dashboardCacheKey_("pnkb");
  return ok_(
    _cachedDashboard_(cacheKey, function () {
      return _buildPNKBDashboard_(ctx);
    }),
  );
}

function _buildPNKBDashboard_(ctx) {
  var membersRepo = new SheetRepository_("members");
  var allActive = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var pnkb = allActive.filter(function (m) {
    return getMemberCategory(m) === MEMBER_CATEGORY.PRA_NIKAH;
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var meetingsRepo = new SheetRepository_("meetings");
  var allAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var allMeetings = meetingsRepo.getAll();
  var meetingsById = buildIndexOne_(allMeetings, "meeting_id");
  var attendanceByMember = buildIndexList_(allAttendance, "member_id");

  var attention = buildAttentionList_(pnkb, attendanceByMember, meetingsById);

  var pnkbIds = {};
  pnkb.forEach(function (m) {
    pnkbIds[m.member_id] = true;
  });
  var pnkbAttendance = allAttendance.filter(function (a) {
    return pnkbIds[a.member_id];
  });
  var hadir = pnkbAttendance.filter(function (a) {
    return a.status === ATTENDANCE_STATUS.HADIR;
  }).length;
  var rate = pnkbAttendance.length
    ? Math.round((hadir / pnkbAttendance.length) * 100)
    : 0;

  return {
    total: pnkb.length,
    aktif: pnkb.filter(function (m) {
      return m.status_pembinaan === MONITORING_STATUS.AKTIF;
    }).length,
    perlu_perhatian: attention.length,
    kehadiran: rate,
    data_belum_lengkap: countIncompleteData_(pnkb),
  };
}

function getAbsensiDashboard_(ctx, params) {
  var today = formatDate(nowIso_());
  var cacheKey = _dashboardCacheKey_("absensi", today);
  return ok_(
    _cachedDashboard_(cacheKey, function () {
      return _buildAbsensiDashboard_();
    }),
  );
}

function _buildAbsensiDashboard_() {
  var meetingsRepo = new SheetRepository_("meetings");
  var today = formatDate(nowIso_());
  var todayMeetings = meetingsRepo.find(function (m) {
    return formatDate(m.tanggal) === today;
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var allAttendance = attendanceRepo.getAllRecent(ATTENDANCE_RECENT_LIMIT);
  var attendanceByMeeting = buildIndexList_(allAttendance, "meeting_id");

  var membersRepo = new SheetRepository_("members");
  var allMembers = membersRepo.find(function (m) {
    return toBool_(m.status_aktif);
  });

  var summary = { HADIR: 0, IJIN: 0, SAKIT: 0, TANPA_KETERANGAN: 0 };
  var totalJamaah = 0;
  var totalTarget = 0;
  var byCategory = {};

  todayMeetings.forEach(function (meeting) {
    var targets = parseKategoriTarget_(meeting.kategori_target);

    var eligibleMembers = allMembers;
    if (targets.length > 0) {
      eligibleMembers = allMembers.filter(function (m) {
        var cat = getMemberCategory(m);
        return cat && targets.indexOf(cat) !== -1;
      });
    }
    totalTarget += eligibleMembers.length;

    var rows = attendanceByMeeting[meeting.meeting_id] || [];
    totalJamaah += rows.length;
    rows.forEach(function (r) {
      if (summary.hasOwnProperty(r.status)) summary[r.status]++;
    });

    var key = targets.length === 0 ? "SEMUA" : targets.slice().sort().join("|");
    if (!byCategory[key]) {
      byCategory[key] = {
        kategori: targets,
        is_semua: targets.length === 0,
        hadir: 0,
        ijin: 0,
        sakit: 0,
        tanpa_keterangan: 0,
        total_absen: 0,
        total_target: 0,
        meeting_count: 0,
      };
    }
    byCategory[key].meeting_count++;
    byCategory[key].total_target += eligibleMembers.length;
    byCategory[key].total_absen += rows.length;
    rows.forEach(function (r) {
      if (r.status === ATTENDANCE_STATUS.HADIR) byCategory[key].hadir++;
      else if (r.status === ATTENDANCE_STATUS.IJIN) byCategory[key].ijin++;
      else if (r.status === ATTENDANCE_STATUS.SAKIT) byCategory[key].sakit++;
      else if (r.status === ATTENDANCE_STATUS.TANPA_KETERANGAN)
        byCategory[key].tanpa_keterangan++;
    });
  });

  var byCategoryList = Object.keys(byCategory).map(function (key) {
    var item = byCategory[key];
    item.persentase = item.total_target
      ? Math.round((item.hadir / item.total_target) * 100)
      : 0;
    return item;
  });

  byCategoryList.sort(function (a, b) {
    if (a.is_semua !== b.is_semua) return a.is_semua ? 1 : -1;
    return b.total_target - a.total_target;
  });

  return {
    pengajian_hari_ini: todayMeetings.map(function (m) {
      return {
        meeting_id: m.meeting_id,
        tanggal: formatDate(m.tanggal),
        jam: m.jam,
        acara: m.acara,
        group_id: m.group_id,
        kategori_target: parseKategoriTarget_(m.kategori_target),
      };
    }),
    jumlah_jamaah: totalJamaah,
    total_target: totalTarget,
    hadir: summary.HADIR,
    ijin: summary.IJIN,
    sakit: summary.SAKIT,
    tanpa_keterangan: summary.TANPA_KETERANGAN,
    by_category: byCategoryList,
  };
}

function getMyDashboard_(ctx) {
  var check = requireMemberLink_(ctx);
  if (check) return check;

  var cacheKey = "dash:my:" + ctx.user.member_id;
  return ok_(
    _cachedDashboard_(cacheKey, function () {
      return _buildMyDashboard_(ctx);
    }),
  );
}

function _buildMyDashboard_(ctx) {
  var membersRepo = new SheetRepository_("members");
  var member = membersRepo.findById("member_id", ctx.user.member_id);
  if (!member) return null;

  var enriched = enrichMember_(member);

  /* Tambah: pendidikan */
  var educationRepo = new SheetRepository_("education");
  enriched.pendidikan = educationRepo.findByField(
    "member_id",
    ctx.user.member_id,
  );

  var profileFields = [
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
    "pendidikan", // ← tambah
  ];

  var profile = {};
  profileFields.forEach(function (f) {
    if (enriched.hasOwnProperty(f)) profile[f] = enriched[f];
  });

  var attendanceRepo = new SheetRepository_("attendance");
  var attendanceRows = attendanceRepo.findByField(
    "member_id",
    ctx.user.member_id,
  );
  var meetingsRepo = new SheetRepository_("meetings");
  var allMeetings = meetingsRepo.getAll();
  var meetingsById = buildIndexOne_(allMeetings, "meeting_id");

  var attendance = attendanceRows.map(function (a) {
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
  attendance.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  attendance = attendance.slice(0, 50);

  var monitoringRepo = new SheetRepository_("monitoring");
  var monitoring = monitoringRepo.findByField("member_id", ctx.user.member_id);
  monitoring.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });
  monitoring = monitoring.slice(0, 20);
  monitoring = monitoring.map(function (r) {
    var c = Object.assign({}, r);
    delete c._row;
    return c;
  });

  var today = formatDate(nowIso_());
  var memberCategory = getMemberCategory(member);
  var upcoming = allMeetings
    .filter(function (m) {
      if (formatDate(m.tanggal) < today) return false;
      var targets = parseKategoriTarget_(m.kategori_target);
      if (targets.length === 0) return true;
      return memberCategory && targets.indexOf(memberCategory) !== -1;
    })
    .sort(function (a, b) {
      return new Date(a.tanggal) - new Date(b.tanggal);
    })
    .slice(0, 5)
    .map(publicMeeting_);

  return {
    profile: profile,
    attendance: attendance,
    monitoring: monitoring,
    upcoming: upcoming,
  };
}

function computeOverallAttendanceRate_(allAttendance) {
  if (!allAttendance || !allAttendance.length) return 0;
  var hadir = allAttendance.filter(function (a) {
    return a.status === ATTENDANCE_STATUS.HADIR;
  }).length;
  return Math.round((hadir / allAttendance.length) * 100);
}

function buildAttentionList_(members, attendanceByMember, meetingsById) {
  var sixMonthsAgo = new Date();
  sixMonthsAgo.setMonth(sixMonthsAgo.getMonth() - 6);

  var list = [];
  members.forEach(function (m) {
    var rows = (attendanceByMember[m.member_id] || []).slice();
    rows.sort(function (a, b) {
      var ma = meetingsById[a.meeting_id];
      var mb = meetingsById[b.meeting_id];
      return new Date(mb ? mb.tanggal : 0) - new Date(ma ? ma.tanggal : 0);
    });

    var reasons = [];
    if (rows.length >= 3) {
      var hadirCount = rows.filter(function (r) {
        return r.status === ATTENDANCE_STATUS.HADIR;
      }).length;
      var rate = Math.round((hadirCount / rows.length) * 100);
      if (rate < 60) reasons.push("Kehadiran " + rate + "%");

      var last3 = rows.slice(0, 3);
      if (
        last3.length === 3 &&
        last3.every(function (r) {
          return r.status !== ATTENDANCE_STATUS.HADIR;
        })
      ) {
        reasons.push("3x berturut-turut tidak hadir");
      }
    }
    if (!m.updated_at || new Date(m.updated_at) < sixMonthsAgo) {
      reasons.push("Data belum diperbarui > 6 bulan");
    }
    if (
      m.status_pembinaan === MONITORING_STATUS.PERLU_PERHATIAN ||
      m.status_pembinaan === MONITORING_STATUS.TIDAK_AKTIF
    ) {
      reasons.push("Status pembinaan: " + m.status_pembinaan);
    }

    if (reasons.length) {
      list.push({
        member_id: m.member_id,
        nama_lengkap: m.nama_lengkap,
        foto_url: m.foto_url,
        reasons: reasons,
      });
    }
  });
  return list;
}

function countIncompleteData_(members) {
  var requiredFields = ["tanggal_lahir", "kelompok", "desa", "alamat_rumah"];
  return members.filter(function (m) {
    return requiredFields.some(function (f) {
      return !m[f];
    });
  }).length;
}
