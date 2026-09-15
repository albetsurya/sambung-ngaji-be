/**
 * WaQueue.js
 *
 * Sistem antrian WA otomatis H-8 jam dari waktu pengajian.
 * Setiap queue terikat ke 1 meeting.
 *
 * Flow:
 *  1. Meeting dibuat → user toggle "WA Auto" → createWaQueue_()
 *  2. Trigger processWaQueue_() setiap 15 menit → cek queue PENDING yang due
 *  3. Kirim WA via sendWhatsApp_() → update status SENT/FAILED
 */

/* -------------------------------------------------------------------------- */
/*                              HELPERS                                       */
/* -------------------------------------------------------------------------- */

/**
 * Resolve jam_start default berdasarkan kategori + acara.
 *
 * Prioritas:
 *  1. Acara mengandung "maghrib" → 18:00
 *  2. Kategori hanya BALITA/CABERAWIT → 15:00
 *  3. Default → 19:00
 */
function resolveDefaultJamStart_(params) {
  var acara = String(params.acara || "").toLowerCase();
  if (acara.indexOf("maghrib") !== -1) return WA_JAM_MAGHRIB;

  var kategori = params.kategori_target || [];
  if (typeof kategori === "string") {
    try {
      kategori = JSON.parse(kategori);
    } catch (e) {
      kategori = [];
    }
  }
  if (Array.isArray(kategori) && kategori.length > 0) {
    var allAnak = kategori.every(function (k) {
      return WA_KATEGORI_ANAK.indexOf(k) !== -1;
    });
    if (allAnak) return WA_JAM_CABERAWIT_BALITA;
  }

  return WA_JAM_DEFAULT;
}

/**
 * Hitung send_at = meeting_date + jam_start - 8 jam.
 *
 * @param {string} meetingDate - "YYYY-MM-DD"
 * @param {string} jamStart - "HH:MM"
 * @returns {string} ISO timestamp UTC, atau "" kalau invalid.
 */
function computeWaSendAt_(meetingDate, jamStart) {
  var p = parseIsoParts_(meetingDate);
  if (!p) return "";

  var jamParts = String(jamStart || "").split(":");
  var hh = Number(jamParts[0]);
  var mm = Number(jamParts[1] || 0);
  if (isNaN(hh) || hh < 0 || hh > 23) hh = 19;
  if (isNaN(mm) || mm < 0 || mm > 59) mm = 0;

  /* Buat Date di timezone lokal (WIB via appsscript.json). */
  var dt = new Date(p.year, p.month - 1, p.day, hh, mm, 0);

  /* Kurangi 8 jam. */
  dt.setHours(dt.getHours() - WA_QUEUE_SEND_OFFSET_HOURS);

  return dt.toISOString();
}

/**
 * Generate queue_id unik.
 */
function generateQueueId_() {
  return newId_("WQ");
}

/* -------------------------------------------------------------------------- */
/*                          CREATE / CANCEL / STATUS                          */
/* -------------------------------------------------------------------------- */

/**
 * Buat queue WA untuk 1 meeting.
 *
 * @param {object} ctx - request context
 * @param {object} params - { meeting_id, template_id, jam_start (opsional) }
 */
function createWaQueue_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");
  if (!params.template_id) return fail_("template_id wajib diisi");

  var meetingsRepo = new SheetRepository_("meetings");
  var meeting = meetingsRepo.findById("meeting_id", params.meeting_id);
  if (!meeting) return fail_("Meeting tidak ditemukan");

  var tplRepo = new SheetRepository_("announcement_templates");
  var template = tplRepo.findById("template_id", params.template_id);
  if (!template) return fail_("Template tidak ditemukan");
  if (!toBool_(template.status_aktif)) return fail_("Template tidak aktif");

  var queueRepo = new SheetRepository_("wa_queue");

  /* Cek existing queue untuk meeting ini. */
  var existing = queueRepo.find(function (q) {
    return q.meeting_id === params.meeting_id;
  });
  var existingActive = existing.filter(function (q) {
    return q.status === WA_QUEUE_STATUS.PENDING;
  });
  if (existingActive.length > 0) {
    return fail_("Meeting sudah punya queue PENDING");
  }

  /* Resolve jam_start. */
  var meetingDate = formatDate(meeting.tanggal);
  var jamStart =
    params.jam_start ||
    meeting.jam_start ||
    resolveDefaultJamStart_({
      acara: meeting.acara,
      kategori_target: meeting.kategori_target,
    });

  var sendAt = computeWaSendAt_(meetingDate, jamStart);
  if (!sendAt) return fail_("Gagal menghitung jadwal kirim WA");

  /* Count target member. */
  var targets = getBulkTargetMembers_({
    kategori_target: meeting.kategori_target,
    group_id: meeting.group_id,
  });

  var now = nowIso_();
  var queue = {
    queue_id: generateQueueId_(),
    meeting_id: params.meeting_id,
    meeting_date: meetingDate,
    jam_start: jamStart,
    send_at: sendAt,
    status: WA_QUEUE_STATUS.PENDING,
    template_id: params.template_id,
    member_count: targets.length,
    sent_count: 0,
    failed_count: 0,
    error_log: "",
    sent_at: "",
    created_at: now,
    updated_at: now,
  };

  queueRepo.insert(queue);

  /* Update meeting jam_start kalau belum ada. */
  if (!meeting.jam_start) {
    meetingsRepo.updateById("meeting_id", params.meeting_id, {
      jam_start: jamStart,
      updated_at: now,
    });
  }

  writeAuditLog_(
    ctx.user.user_id,
    "CREATE_WA_QUEUE",
    "MEETING",
    params.meeting_id,
  );

  return ok_(queue);
}

/**
 * Cancel queue untuk 1 meeting (set status CANCELLED).
 */
function cancelWaQueue_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var queueRepo = new SheetRepository_("wa_queue");
  var queues = queueRepo.find(function (q) {
    return q.meeting_id === params.meeting_id;
  });
  var pending = queues.filter(function (q) {
    return q.status === WA_QUEUE_STATUS.PENDING;
  });

  if (pending.length === 0) return ok_({ cancelled: 0 });

  pending.forEach(function (q) {
    queueRepo.updateById("queue_id", q.queue_id, {
      status: WA_QUEUE_STATUS.CANCELLED,
      updated_at: nowIso_(),
    });
  });

  writeAuditLog_(
    ctx.user.user_id,
    "CANCEL_WA_QUEUE",
    "MEETING",
    params.meeting_id,
  );

  return ok_({ cancelled: pending.length });
}

/**
 * Ambil status queue untuk 1 meeting (untuk render toggle di AttendancePage).
 */
function getWaQueueStatus_(ctx, params) {
  if (!params.meeting_id) return fail_("meeting_id wajib diisi");

  var queueRepo = new SheetRepository_("wa_queue");
  var queues = queueRepo.find(function (q) {
    return q.meeting_id === params.meeting_id;
  });

  if (queues.length === 0) {
    return ok_({
      has_queue: false,
      active: null,
      latest: null,
    });
  }

  /* Sort terbaru dulu. */
  queues.sort(function (a, b) {
    return String(b.created_at || "").localeCompare(String(a.created_at || ""));
  });

  var active = queues.filter(function (q) {
    return q.status === WA_QUEUE_STATUS.PENDING;
  })[0];

  return ok_({
    has_queue: true,
    active: active || null,
    latest: queues[0],
  });
}

/**
 * Bulk get status untuk banyak meeting sekaligus (untuk list).
 * Params: { meeting_ids: string[] }
 */
function bulkGetWaQueueStatus_(ctx, params) {
  var ids = params.meeting_ids || [];
  if (typeof ids === "string") {
    try {
      ids = JSON.parse(ids);
    } catch (e) {
      ids = [];
    }
  }
  if (!Array.isArray(ids) || ids.length === 0) return ok_({});

  var idMap = {};
  ids.forEach(function (id) {
    idMap[id] = true;
  });

  var queueRepo = new SheetRepository_("wa_queue");
  var all = queueRepo.getAll();
  var result = {};

  all.forEach(function (q) {
    if (!idMap[q.meeting_id]) return;
    if (!result[q.meeting_id]) result[q.meeting_id] = [];
    result[q.meeting_id].push(q);
  });

  /* Untuk setiap meeting, ambil yang PENDING (kalau ada) atau yang terbaru. */
  var out = {};
  Object.keys(result).forEach(function (mid) {
    var queues = result[mid];
    queues.sort(function (a, b) {
      return String(b.created_at || "").localeCompare(
        String(a.created_at || ""),
      );
    });
    var active = queues.filter(function (q) {
      return q.status === WA_QUEUE_STATUS.PENDING;
    })[0];
    out[mid] = {
      has_queue: true,
      active: active || null,
      latest: queues[0],
    };
  });

  return ok_(out);
}

/**
 * List queue dengan filter (untuk WaQueuePage).
 * Params: { status?, from?, to?, limit? }
 */
function listWaQueue_(ctx, params) {
  var repo = new SheetRepository_("wa_queue");
  var all = repo.getAll();

  if (params.status) {
    all = all.filter(function (q) {
      return q.status === params.status;
    });
  }
  if (params.from) {
    all = all.filter(function (q) {
      return String(q.meeting_date || "") >= String(params.from);
    });
  }
  if (params.to) {
    all = all.filter(function (q) {
      return String(q.meeting_date || "") <= String(params.to);
    });
  }

  all.sort(function (a, b) {
    return String(b.send_at || "").localeCompare(String(a.send_at || ""));
  });

  var limit = params.limit ? Number(params.limit) : 100;
  if (isNaN(limit) || limit <= 0) limit = 100;
  if (limit > 500) limit = 500;

  var sliced = all.slice(0, limit);

  return ok_({
    items: sliced.map(function (q) {
      var c = Object.assign({}, q);
      delete c._row;
      return c;
    }),
    total: all.length,
    limit: limit,
  });
}

/**
 * Retry queue yang FAILED → reset ke PENDING.
 */
function retryWaQueue_(ctx, params) {
  if (!params.queue_id) return fail_("queue_id wajib diisi");

  var repo = new SheetRepository_("wa_queue");
  var queue = repo.findById("queue_id", params.queue_id);
  if (!queue) return fail_("Queue tidak ditemukan");
  if (queue.status !== WA_QUEUE_STATUS.FAILED) {
    return fail_("Hanya queue FAILED yang bisa di-retry");
  }

  var updated = repo.updateById("queue_id", params.queue_id, {
    status: WA_QUEUE_STATUS.PENDING,
    error_log: "",
    updated_at: nowIso_(),
  });

  writeAuditLog_(
    ctx.user.user_id,
    "RETRY_WA_QUEUE",
    "MEETING",
    queue.meeting_id,
  );

  return ok_(updated);
}

/* -------------------------------------------------------------------------- */
/*                          TRIGGER HANDLER                                   */
/* -------------------------------------------------------------------------- */

/**
 * Trigger runner: dipanggil setiap 15 menit.
 * Cek queue PENDING yang due → kirim WA → update status.
 */
function processWaQueue_() {
  var lock = LockService.getScriptLock();
  if (!lock.tryLock(30000)) {
    Logger.log("[WA Queue] Lock gagal, skip");
    return;
  }

  try {
    var repo = new SheetRepository_("wa_queue");
    var all = repo.getAll();
    var now = Date.now();

    var due = all.filter(function (q) {
      if (q.status !== WA_QUEUE_STATUS.PENDING) return false;
      var sendAt = new Date(q.send_at).getTime();
      if (isNaN(sendAt)) return false;
      return sendAt <= now;
    });

    Logger.log("[WA Queue] Total queue: " + all.length);
    Logger.log("[WA Queue] Due queues: " + due.length);

    if (due.length === 0) return;

    due = due.slice(0, WA_QUEUE_BATCH_SIZE);

    due.forEach(function (q) {
      try {
        _processOneQueue_(q, repo);
      } catch (e) {
        Logger.log("[WA Queue] Error queue " + q.queue_id + ": " + e);
      }
    });

    repo._invalidateCache();
  } finally {
    lock.releaseLock();
  }
}

/**
 * Proses 1 queue: kirim WA ke semua target, update status.
 */
function _processOneQueue_(queue, queueRepo) {
  var meetingRepo = new SheetRepository_("meetings");
  var meeting = meetingRepo.findById("meeting_id", queue.meeting_id);
  if (!meeting) {
    queueRepo.updateById("queue_id", queue.queue_id, {
      status: WA_QUEUE_STATUS.FAILED,
      error_log: JSON.stringify(["Meeting tidak ditemukan"]),
      sent_at: nowIso_(),
      updated_at: nowIso_(),
    });
    return;
  }

  var tplRepo = new SheetRepository_("announcement_templates");
  var template = tplRepo.findById("template_id", queue.template_id);
  if (!template) {
    queueRepo.updateById("queue_id", queue.queue_id, {
      status: WA_QUEUE_STATUS.FAILED,
      error_log: JSON.stringify(["Template tidak ditemukan"]),
      sent_at: nowIso_(),
      updated_at: nowIso_(),
    });
    return;
  }

  var targets = getBulkTargetMembers_({
    kategori_target: meeting.kategori_target,
    group_id: meeting.group_id,
  });

  if (targets.length === 0) {
    queueRepo.updateById("queue_id", queue.queue_id, {
      status: WA_QUEUE_STATUS.FAILED,
      error_log: JSON.stringify(["Tidak ada target member"]),
      member_count: 0,
      sent_at: nowIso_(),
      updated_at: nowIso_(),
    });
    return;
  }

  /* Resolve group info. */
  var groupName = "";
  var penandatangan = "";
  if (meeting.group_id) {
    var groupsRepo = new SheetRepository_("groups");
    var group = groupsRepo.findById("group_id", meeting.group_id);
    if (group) {
      groupName = group.group_name || "";
      penandatangan = group.penandatangan || "";
    }
  }

  /* Build message text. */
  var templateData = {
    nama_kelompok: groupName,
    hari: meeting.hari,
    tanggal: formatDateShort(meeting.tanggal),
    jam: meeting.jam,
    acara: meeting.acara,
    materi: meeting.materi || "",
    catatan: meeting.catatan || "",
    penandatangan: penandatangan,
  };
  var messageText = renderTemplate_(template.isi_template, templateData);

  /* Send to all targets. */
  var sentCount = 0;
  var failedCount = 0;
  var errors = [];

  targets.forEach(function (m) {
    try {
      var res = sendWhatsApp_(m.no_wa, messageText);
      if (res && res.ok) {
        sentCount++;
      } else if (res && res.skipped) {
        /* gateway off, skip silently */
      } else {
        failedCount++;
        if (errors.length < 5) {
          errors.push(m.no_wa + ": " + (res.reason || res.error || "unknown"));
        }
      }
    } catch (e) {
      failedCount++;
      if (errors.length < 5) errors.push(m.no_wa + ": " + e.message);
    }
  });

  /* Determine final status. */
  var finalStatus;
  if (sentCount > 0 && failedCount === 0) {
    finalStatus = WA_QUEUE_STATUS.SENT;
  } else if (sentCount > 0 && failedCount > 0) {
    finalStatus = WA_QUEUE_STATUS.SENT; /* partial, tetap SENT */
  } else if (sentCount === 0 && failedCount > 0) {
    finalStatus = WA_QUEUE_STATUS.FAILED;
  } else {
    /* sentCount = 0, failedCount = 0 → semua skip (gateway off) */
    finalStatus = WA_QUEUE_STATUS.FAILED;
    errors.push("WA gateway disabled / skip semua");
  }

  queueRepo.updateById("queue_id", queue.queue_id, {
    status: finalStatus,
    member_count: targets.length,
    sent_count: sentCount,
    failed_count: failedCount,
    error_log: errors.length > 0 ? JSON.stringify(errors) : "",
    sent_at: nowIso_(),
    updated_at: nowIso_(),
  });

  Logger.log(
    "[WA Queue] Processed " +
      queue.queue_id +
      ": sent=" +
      sentCount +
      ", failed=" +
      failedCount +
      ", status=" +
      finalStatus,
  );
}

/* -------------------------------------------------------------------------- */
/*                          CLEANUP                                           */
/* -------------------------------------------------------------------------- */

/**
 * Bersihkan queue SENT/FAILED/CANCELLED yang lebih tua dari X hari.
 * Jalankan via trigger harian.
 */
function cleanupOldWaQueue_() {
  var cutoffDays = 30;
  var cutoff = new Date();
  cutoff.setDate(cutoff.getDate() - cutoffDays);
  var cutoffIso = cutoff.toISOString();

  var repo = new SheetRepository_("wa_queue");
  var sheet = repo._sheet();
  var headers = repo.def.headers;
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return 0;

  var createdAtColIdx = headers.indexOf("created_at");
  var statusColIdx = headers.indexOf("status");
  var range = sheet.getRange(2, 1, lastRow - 1, headers.length);
  var values = range.getValues();

  var rowsToDelete = [];
  for (var i = 0; i < values.length; i++) {
    var status = values[i][statusColIdx];
    var createdAt = String(values[i][createdAtColIdx] || "");
    if (status === WA_QUEUE_STATUS.PENDING) continue;
    if (createdAt && createdAt < cutoffIso) {
      rowsToDelete.push(i + 2);
    }
  }

  if (rowsToDelete.length === 0) return 0;

  rowsToDelete.sort(function (a, b) {
    return b - a;
  });
  rowsToDelete.forEach(function (r) {
    sheet.deleteRow(r);
  });

  repo._invalidateCache();
  Logger.log("[WA Queue] Cleaned " + rowsToDelete.length + " old queue rows");
  return rowsToDelete.length;
}
