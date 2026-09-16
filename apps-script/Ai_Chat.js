var AI_PROVIDER_DEFAULT = "omniroute";
var PROP_KEY_PROVIDER = "AI_PROVIDER";
var ALLOWED_PROVIDERS = ["omniroute", "gemini", "groq", "auto"];
var AI_CHAT_MAX_TOOL_ROUNDS = 6;
var AI_CHAT_MAX_ROWS = 100;

var GEMINI_CONFIG = {
  model: "gemini-3.6-flash",
  endpoint: "https://generativelanguage.googleapis.com/v1beta/models/",
  get apiKey() {
    return PropertiesService.getScriptProperties().getProperty(
      "GEMINI_API_KEY",
    );
  },
};

var OMNIROUTE_CONFIG = {
  get endpoint() {
    return PropertiesService.getScriptProperties().getProperty(
      "OMNIROUTE_ENDPOINT",
    );
  },
  get model() {
    return (
      PropertiesService.getScriptProperties().getProperty("OMNIROUTE_MODEL") ||
      "auto/best-vision"
    );
  },
  timeout: 15000,
  get apiKey() {
    return PropertiesService.getScriptProperties().getProperty(
      "OMNIROUTE_API_KEY",
    );
  },
};

var GROQ_CONFIG = {
  endpoint: "https://api.groq.com/openai/v1",
  models: [
    "llama-3.3-70b-versatile",
    "llama-3.1-8b-instant",
    "openai/gpt-oss-120b",
  ],
  timeout: 30000,
  get apiKey() {
    return PropertiesService.getScriptProperties().getProperty("GROQ_API_KEY");
  },
};

var CACHE_PROVIDER = CacheService.getScriptCache();
var CACHE_KEY = "active_provider";
var CACHE_EXPIRATION = 300;

function getStoredProvider_() {
  try {
    var props = PropertiesService.getScriptProperties();
    var stored = props.getProperty(PROP_KEY_PROVIDER);
    if (stored && ALLOWED_PROVIDERS.indexOf(stored) !== -1) return stored;
  } catch (e) {
    console.error("Gagal baca provider:", e);
  }
  return AI_PROVIDER_DEFAULT;
}

function storeProvider_(provider) {
  try {
    if (ALLOWED_PROVIDERS.indexOf(provider) === -1) return false;
    PropertiesService.getScriptProperties().setProperty(
      PROP_KEY_PROVIDER,
      provider,
    );
    return true;
  } catch (e) {
    console.error("Gagal simpan provider:", e);
    return false;
  }
}

function isOmniRouteAvailable() {
  try {
    var cached = CACHE_PROVIDER.get(CACHE_KEY);
    if (cached === "available") return true;
    if (cached === "unavailable") return false;

    var url = OMNIROUTE_CONFIG.endpoint + "/models";
    var response = UrlFetchApp.fetch(url, {
      method: "get",
      headers: { "ngrok-skip-browser-warning": "true" },
      muteHttpExceptions: true,
      timeout: OMNIROUTE_CONFIG.timeout,
    });

    if (response.getResponseCode() >= 200 && response.getResponseCode() < 300) {
      CACHE_PROVIDER.put(CACHE_KEY, "available", CACHE_EXPIRATION);
      return true;
    }
  } catch (e) {
    console.log("OmniRoute tidak tersedia:", e.message);
  }

  CACHE_PROVIDER.put(CACHE_KEY, "unavailable", CACHE_EXPIRATION);
  return false;
}

function getActiveProvider() {
  var stored = getStoredProvider_();
  if (stored === "omniroute") return "omniroute";
  if (stored === "gemini") return "gemini";
  if (stored === "groq") return "groq";
  return isOmniRouteAvailable() ? "omniroute" : "gemini";
}

function setAIProvider(provider) {
  if (ALLOWED_PROVIDERS.indexOf(provider) === -1) {
    return fail_("Provider tidak valid");
  }
  if (!storeProvider_(provider)) {
    return fail_("Gagal menyimpan provider");
  }
  CACHE_PROVIDER.remove(CACHE_KEY);
  return ok_({
    provider: provider,
    active: provider === "auto" ? getActiveProvider() : provider,
  });
}

function getCurrentProvider() {
  var stored = getStoredProvider_();
  return ok_({
    provider: stored,
    active: stored === "auto" ? getActiveProvider() : stored,
  });
}

function handleAiChat_(body, ctx) {
  var isMember = ctx && ctx.user && ctx.user.role === ROLES.MEMBER;

  var quotaCheck = checkAiQuota_(ctx);
  if (quotaCheck) return quotaCheck;

  if (isMember) {
    return handleAiChatMember_(body, ctx);
  }

  var provider = body.provider || getStoredProvider_();
  var user = ctx && ctx.user ? ctx.user.nama : "Pengguna";

  var result;
  if (provider === "groq") {
    result = groqChat_(body, user, ctx);
    if (!result.success && isQuotaError_(result.message)) {
      var omniFallback = omnirouteChat_(body, user, ctx);
      if (omniFallback.success) {
        omniFallback.data.provider = "omniroute";
        omniFallback.data.requestedProvider = "groq";
        return omniFallback;
      }
      var geminiFallback = geminiChat_(body, user, ctx);
      if (geminiFallback.success) {
        geminiFallback.data.provider = "gemini";
        geminiFallback.data.requestedProvider = "groq";
      }
      return geminiFallback;
    }
  } else if (provider === "omniroute") {
    result = omnirouteChat_(body, user, ctx);
    if (!result.success && isQuotaError_(result.message)) {
      var groqFallback = groqChat_(body, user, ctx);
      if (groqFallback.success) {
        groqFallback.data.provider = "groq";
        groqFallback.data.requestedProvider = "omniroute";
        return groqFallback;
      }
      var geminiFallback2 = geminiChat_(body, user, ctx);
      if (geminiFallback2.success) {
        geminiFallback2.data.provider = "gemini";
        geminiFallback2.data.requestedProvider = "omniroute";
      }
      return geminiFallback2;
    }
  } else {
    result = geminiChat_(body, user, ctx);
    if (!result.success && isQuotaError_(result.message)) {
      var groqFallback2 = groqChat_(body, user, ctx);
      if (groqFallback2.success) {
        groqFallback2.data.provider = "groq";
        groqFallback2.data.requestedProvider = "gemini";
        return groqFallback2;
      }
      var omniFallback2 = omnirouteChat_(body, user, ctx);
      if (omniFallback2.success) {
        omniFallback2.data.provider = "omniroute";
        omniFallback2.data.requestedProvider = "gemini";
      }
      return omniFallback2;
    }
  }
  return result;
}

function handleAiChatMember_(body, ctx) {
  var user = ctx.user.nama;

  var omniResult = omnirouteChatMember_(body, user, ctx);

  if (omniResult.success) {
    return omniResult;
  }

  if (isQuotaError_(omniResult.message)) {
    var groqResult = groqChatMember_(body, user, ctx);
    if (groqResult.success) {
      groqResult.data.provider = "groq";
      groqResult.data.requestedProvider = "omniroute";
      return groqResult;
    }
  }

  return fail_("Asisten AI sedang tidak tersedia. Silakan coba lagi nanti.", {
    provider: "omniroute",
  });
}

function isQuotaError_(message) {
  if (!message) return false;
  var msg = String(message);
  return (
    msg.indexOf("429") !== -1 ||
    msg.indexOf("RESOURCE_EXHAUSTED") !== -1 ||
    msg.indexOf("quota") !== -1 ||
    msg.indexOf("rate limit") !== -1
  );
}

var AI_TOOL_DEFS_ = [
  {
    name: "get_dashboard_summary",
    description:
      "Ambil ringkasan dashboard: total jamaah aktif, rata-rata kehadiran, jumlah yang perlu perhatian, dan data belum lengkap.",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_members_list",
    description:
      "Ambil daftar jamaah. Bisa difilter berdasarkan kategori, jenis kelamin, kelompok, atau pencarian nama.",
    parameters: {
      type: "object",
      properties: {
        kategori: {
          type: "string",
          description:
            "BALITA | CABERAWIT | PRA_REMAJA | REMAJA | PRA_NIKAH | DEWASA | ISTIMEWA",
        },
        jenis_kelamin: { type: "string", description: "L atau P" },
        kelompok: { type: "string", description: "group_id" },
        search: { type: "string", description: "Kata kunci nama" },
        limit: { type: "number", description: "Maksimal hasil, default 50" },
      },
      required: [],
    },
  },
  {
    name: "get_member_detail",
    description: "Ambil detail lengkap satu jamaah berdasarkan member_id.",
    parameters: {
      type: "object",
      properties: { member_id: { type: "string", description: "ID jamaah" } },
      required: ["member_id"],
    },
  },
  {
    name: "get_groups_list",
    description: "Ambil daftar kelompok pengajian beserta pembina dan jadwal.",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_attendance_summary",
    description:
      "Ambil ringkasan absensi: total hadir/ijin/sakit/alpa dalam periode tertentu.",
    parameters: {
      type: "object",
      properties: {
        from: { type: "string", description: "Tanggal mulai YYYY-MM-DD" },
        to: { type: "string", description: "Tanggal akhir YYYY-MM-DD" },
        group_id: { type: "string", description: "Filter kelompok (opsional)" },
      },
      required: [],
    },
  },
  {
    name: "get_attendance_by_meeting",
    description: "Ambil detail absensi satu meeting tertentu.",
    parameters: {
      type: "object",
      properties: { meeting_id: { type: "string", description: "ID meeting" } },
      required: ["meeting_id"],
    },
  },
  {
    name: "get_upcoming_meetings",
    description: "Ambil daftar jadwal pengajian yang akan datang.",
    parameters: {
      type: "object",
      properties: {
        limit: { type: "number", description: "Maksimal hasil, default 10" },
      },
      required: [],
    },
  },
  {
    name: "get_monitoring_list",
    description:
      "Ambil daftar monitoring/pembinaan jamaah. Bisa filter berdasarkan status.",
    parameters: {
      type: "object",
      properties: {
        status: {
          type: "string",
          description: "AKTIF | PERLU_PERHATIAN | KURANG_AKTIF | TIDAK_AKTIF",
        },
        member_id: { type: "string", description: "Filter member tertentu" },
      },
      required: [],
    },
  },
  {
    name: "get_announcements_list",
    description: "Ambil daftar pengumuman yang pernah dibuat.",
    parameters: {
      type: "object",
      properties: {
        status: {
          type: "string",
          description: "DRAFT | READY | SHARED | CANCELLED",
        },
        group_id: { type: "string", description: "Filter kelompok" },
      },
      required: [],
    },
  },
];

var AI_TOOL_DEFS_MEMBER_ = [
  {
    name: "get_my_profile",
    description:
      "Ambil biodata diri sendiri (nama, usia, kelompok, alamat, pendidikan, dll).",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_my_attendance",
    description:
      "Ambil riwayat absensi pengajian diri sendiri (tanggal, acara, status hadir/ijin/sakit/alpa).",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_my_attendance_stats",
    description:
      "Ambil statistik kehadiran pribadi: total hadir, ijin, sakit, alpa, dan persentase kehadiran.",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_my_monitoring",
    description:
      "Ambil riwayat pembinaan/monitoring diri sendiri (tanggal, status, catatan).",
    parameters: { type: "object", properties: {}, required: [] },
  },
  {
    name: "get_my_upcoming_meetings",
    description:
      "Ambil jadwal pengajian yang akan datang (untuk informasi pribadi).",
    parameters: {
      type: "object",
      properties: {
        limit: { type: "number", description: "Maksimal hasil, default 5" },
      },
      required: [],
    },
  },
];

function executeAiTool_(name, args, ctx) {
  args = args || {};
  try {
    var result;
    switch (name) {
      case "get_dashboard_summary":
        result = getDashboard_(ctx, {});
        break;
      case "get_members_list":
        result = getMembers_(ctx, {
          kategori: args.kategori || "",
          jenis_kelamin: args.jenis_kelamin || "",
          kelompok: args.kelompok || "",
          search: args.search || "",
          limit: args.limit || 50,
        });
        break;
      case "get_member_detail":
        if (!args.member_id) return fail_("member_id wajib diisi");
        result = getMemberDetail_(ctx, { member_id: args.member_id });
        break;
      case "get_groups_list":
        result = getGroups_(ctx, {});
        break;
      case "get_attendance_summary":
        result = getAttendanceSummaryForAi_(args);
        break;
      case "get_attendance_by_meeting":
        if (!args.meeting_id) return fail_("meeting_id wajib diisi");
        result = getAttendance_(ctx, { meeting_id: args.meeting_id });
        break;
      case "get_upcoming_meetings":
        result = getUpcomingMeetingsForAi_(args.limit || 10);
        break;
      case "get_monitoring_list":
        result = getMonitoringListForAi_(args);
        break;
      case "get_announcements_list":
        result = getAnnouncements_(ctx, {
          status: args.status || "",
          group_id: args.group_id || "",
          limit: 50,
        });
        break;
      default:
        return fail_("Tool tidak dikenal: " + name);
    }
    return capToolResultRows_(result);
  } catch (err) {
    console.error("[AI Tool] Error menjalankan " + name + ":", err);
    return fail_("Gagal mengambil data: " + err.message);
  }
}

function executeAiToolMember_(name, args, ctx) {
  args = args || {};
  var memberId = ctx.user.member_id;

  if (!memberId) {
    return fail_("Akun Anda belum terhubung ke data jamaah. Hubungi admin.");
  }

  try {
    var result;
    switch (name) {
      case "get_my_profile":
        result = getMyProfile_(ctx, {});
        break;
      case "get_my_attendance":
        result = getMyAttendance_(ctx, {});
        break;
      case "get_my_attendance_stats":
        result = getMyAttendanceStats_(ctx);
        break;
      case "get_my_monitoring":
        result = getMyMonitoring_(ctx, {});
        break;
      case "get_my_upcoming_meetings":
        result = getUpcomingMeetings_(ctx, { limit: args.limit || 5 });
        break;
      default:
        return fail_("Tool tidak dikenal: " + name);
    }
    return capToolResultRows_(result);
  } catch (err) {
    console.error("[AI Tool Member] Error " + name + ":", err);
    return fail_("Gagal mengambil data: " + err.message);
  }
}

function getMyAttendanceStats_(ctx) {
  var repo = new SheetRepository_("attendance");
  var rows = repo.findByField("member_id", ctx.user.member_id);

  var counts = {
    HADIR: 0,
    IJIN: 0,
    SAKIT: 0,
    TANPA_KETERANGAN: 0,
  };

  rows.forEach(function (r) {
    if (counts.hasOwnProperty(r.status)) counts[r.status]++;
  });

  var total = rows.length;
  var persentase = total ? Math.round((counts.HADIR / total) * 100) : 0;

  return ok_({
    total_absensi: total,
    hadir: counts.HADIR,
    ijin: counts.IJIN,
    sakit: counts.SAKIT,
    tanpa_keterangan: counts.TANPA_KETERANGAN,
    persentase_hadir: persentase + "%",
  });
}

function getAttendanceSummaryForAi_(args) {
  var attendanceRepo = new SheetRepository_("attendance");
  var meetingsRepo = new SheetRepository_("meetings");
  var all = attendanceRepo.getAll();
  var meetingsById = buildIndexOne_(meetingsRepo.getAll(), "meeting_id");

  var from = args.from || "";
  var to = args.to || "";
  var groupId = args.group_id || "";

  var filtered = all.filter(function (a) {
    var m = meetingsById[a.meeting_id];
    if (!m) return false;
    if (groupId && m.group_id !== groupId) return false;
    var tgl = formatDate(m.tanggal);
    if (from && tgl < from) return false;
    if (to && tgl > to) return false;
    return true;
  });

  var summary = {
    HADIR: 0,
    IJIN: 0,
    SAKIT: 0,
    TANPA_KETERANGAN: 0,
    TOTAL: filtered.length,
  };
  filtered.forEach(function (a) {
    if (summary.hasOwnProperty(a.status)) summary[a.status]++;
  });

  var rate = summary.TOTAL
    ? Math.round((summary.HADIR / summary.TOTAL) * 100)
    : 0;

  return ok_({
    periode: { from: from || "awal", to: to || "sekarang" },
    total_absensi: summary.TOTAL,
    hadir: summary.HADIR,
    ijin: summary.IJIN,
    sakit: summary.SAKIT,
    tanpa_keterangan: summary.TANPA_KETERANGAN,
    persentase_hadir: rate + "%",
  });
}

function getUpcomingMeetingsForAi_(limit) {
  var repo = new SheetRepository_("meetings");
  var today = formatDate(nowIso_());
  var upcoming = repo
    .getAll()
    .filter(function (m) {
      return formatDate(m.tanggal) >= today;
    })
    .sort(function (a, b) {
      return new Date(a.tanggal) - new Date(b.tanggal);
    })
    .slice(0, limit);

  return ok_(
    upcoming.map(function (m) {
      return {
        meeting_id: m.meeting_id,
        tanggal: formatDate(m.tanggal), // ✅ FIX
        hari: m.hari,
        jam: m.jam,
        acara: m.acara,
        group_id: m.group_id,
      };
    }),
  );
}

function getMonitoringListForAi_(args) {
  var repo = new SheetRepository_("monitoring");
  var all = repo.getAll();
  if (args.member_id)
    all = all.filter(function (m) {
      return m.member_id === args.member_id;
    });
  if (args.status)
    all = all.filter(function (m) {
      return m.status === args.status;
    });
  all.sort(function (a, b) {
    return new Date(b.tanggal) - new Date(a.tanggal);
  });

  return ok_(
    all.slice(0, 100).map(function (m) {
      var c = Object.assign({}, m);
      delete c._row;
      return c;
    }),
  );
}

function capToolResultRows_(obj) {
  if (obj === null || typeof obj !== "object") return obj;

  function walk(node) {
    if (Array.isArray(node)) {
      if (node.length > AI_CHAT_MAX_ROWS) {
        var truncated = node.slice(-AI_CHAT_MAX_ROWS);
        truncated._truncatedNotice =
          "Menampilkan " +
          AI_CHAT_MAX_ROWS +
          " dari " +
          node.length +
          " baris.";
        return truncated;
      }
      return node;
    }
    if (node && typeof node === "object") {
      var out = {};
      for (var key in node) {
        if (Object.prototype.hasOwnProperty.call(node, key)) {
          out[key] = walk(node[key]);
        }
      }
      return out;
    }
    return node;
  }

  return walk(obj);
}

function buildAiSystemPrompt_(user) {
  var today = Utilities.formatDate(
    new Date(),
    Session.getScriptTimeZone() || "Asia/Jakarta",
    "EEEE, d MMMM yyyy",
  );

  return [
    "Kamu adalah asisten AI untuk aplikasi Manajemen Pengajian yang mencakup data jamaah, kelompok, absensi, monitoring, dan pengumuman.",
    "Hari ini: " + today + ".",
    user ? "Kamu sedang berbicara dengan: " + user + "." : "",
    "",
    "FORMAT JAWABAN (WAJIB DIIKUTI):",
    "1. Selalu pakai bullet list dengan tanda '-' untuk daftar.",
    "2. Setiap item di baris terpisah.",
    "3. Pakai **bold** untuk nama orang dan angka penting.",
    "4. Beri jarak kosong antar bagian.",
    "5. Akhiri dengan ringkasan singkat atau catatan penting.",
    "",
    "PENTING:",
    "- Gunakan tools untuk mengambil data ASLI sebelum menjawab pertanyaan yang menyebut angka, nama, atau statistik.",
    "- JANGAN pernah mengarang atau menebak data.",
    "- Jika data tidak ditemukan, katakan dengan jujur.",
    "- Jawab dalam Bahasa Indonesia yang ramah dan profesional.",
  ]
    .filter(Boolean)
    .join("\n");
}

function buildAiSystemPromptMember_(user) {
  var today = Utilities.formatDate(
    new Date(),
    Session.getScriptTimeZone() || "Asia/Jakarta",
    "EEEE, d MMMM yyyy",
  );

  return [
    "Kamu adalah asisten AI pribadi untuk jamaah pengajian.",
    "Hari ini: " + today + ".",
    "Kamu sedang berbicara dengan: " + user + ".",
    "",
    "ATURAN KETAT:",
    "1. Kamu HANYA boleh menjawab pertanyaan tentang DATA DIRI user ini:",
    "   - Biodata pribadi (nama, usia, kelompok, alamat, pendidikan)",
    "   - Riwayat absensi pribadi",
    "   - Statistik kehadiran pribadi",
    "   - Riwayat pembinaan pribadi",
    "   - Jadwal pengajian mendatang",
    "",
    "2. Kamu DILARANG KERAS:",
    "   - Menyebut nama jamaah lain",
    "   - Memberi data statistik global",
    "   - Menjawab tentang kelompok lain",
    "   - Membahas dashboard atau struktur organisasi",
    "   - Membahas data user lain",
    "",
    "3. Kalau user bertanya tentang orang lain atau data global, jawab:",
    "   Maaf, saya hanya bisa membantu dengan data pribadi Anda.",
    "",
    "4. Jangan pernah mengarang data.",
    "",
    "FORMAT JAWABAN:",
    "- Pakai bullet list dengan tanda '-' untuk daftar",
    "- Pakai **bold** untuk angka penting",
    "- Jawab dengan ramah dan hangat",
    "- Bahasa Indonesia",
    "",
    "PENTING:",
    "- Gunakan tools untuk ambil data ASLI sebelum menjawab.",
    "- JANGAN pernah mengarang atau menebak angka.",
  ].join("\n");
}

function historyToChatMessages_(history) {
  if (!Array.isArray(history)) return [];
  return history
    .filter(function (h) {
      return h && h.text;
    })
    .slice(-10)
    .map(function (h) {
      return {
        role: h.role === "assistant" ? "assistant" : "user",
        content: String(h.text).substring(0, 1000),
      };
    });
}

function omnirouteChat_(body, user, ctx) {
  var endpoint = OMNIROUTE_CONFIG.endpoint;
  var apiKey = OMNIROUTE_CONFIG.apiKey;

  if (!endpoint) {
    return fail_("OMNIROUTE_ENDPOINT belum diatur.");
  }

  var messages = [{ role: "system", content: buildAiSystemPrompt_(user) }]
    .concat(historyToChatMessages_(body.history))
    .concat([{ role: "user", content: String(body.message || "") }]);

  var tools = AI_TOOL_DEFS_.map(function (t) {
    return {
      type: "function",
      function: {
        name: t.name,
        description: t.description,
        parameters: t.parameters,
      },
    };
  });

  var rounds = 0;
  while (rounds < AI_CHAT_MAX_TOOL_ROUNDS) {
    rounds++;

    var payload = {
      model: OMNIROUTE_CONFIG.model,
      messages: messages,
      tools: tools,
      tool_choice: "auto",
      temperature: 0.3,
    };

    var response;
    try {
      response = UrlFetchApp.fetch(endpoint + "/chat/completions", {
        method: "post",
        contentType: "application/json",
        headers: {
          Authorization: apiKey ? "Bearer " + apiKey : "",
          "ngrok-skip-browser-warning": "true",
        },
        payload: JSON.stringify(payload),
        muteHttpExceptions: true,
        timeout: OMNIROUTE_CONFIG.timeout,
      });
    } catch (err) {
      return fail_("Gagal menghubungi OmniRoute: " + err.message);
    }

    var status = response.getResponseCode();
    var responseText = response.getContentText();

    if (responseText.indexOf("<!DOCTYPE html>") === 0) {
      return fail_("OmniRoute mengembalikan HTML (ngrok warning?).");
    }

    if (status < 200 || status >= 300) {
      return fail_("OmniRoute error HTTP " + status);
    }

    var data;
    try {
      data = JSON.parse(responseText);
    } catch (e) {
      return fail_("Respons OmniRoute bukan JSON.");
    }

    var choice = data.choices && data.choices[0];
    var message = choice && choice.message;
    if (!message) return fail_("Respons OmniRoute kosong.");

    var toolCalls = message.tool_calls;
    if (toolCalls && toolCalls.length > 0) {
      messages.push({
        role: "assistant",
        content: message.content || null,
        tool_calls: toolCalls,
      });

      toolCalls.forEach(function (call) {
        var args = {};
        try {
          args = JSON.parse(call.function.arguments || "{}");
        } catch (e) {}
        var toolResult = executeAiTool_(call.function.name, args, ctx);
        messages.push({
          role: "tool",
          tool_call_id: call.id,
          content: JSON.stringify(toolResult),
        });
      });

      continue;
    }

    if (data.usage) {
      logAiUsage_(
        ctx,
        "omniroute",
        data.usage.prompt_tokens,
        data.usage.completion_tokens,
      );
    }

    incrementAiQuota_(ctx); // FASE 1 FIX: kuota AI

    return ok_({
      reply: message.content || "Maaf, saya tidak mendapatkan jawaban.",
    });
  }

  return fail_("Terlalu banyak proses pengambilan data.");
}

function omnirouteChatMember_(body, user, ctx) {
  var endpoint = OMNIROUTE_CONFIG.endpoint;
  var apiKey = OMNIROUTE_CONFIG.apiKey;

  if (!endpoint) {
    return fail_("Konfigurasi AI belum lengkap.");
  }

  var messages = [{ role: "system", content: buildAiSystemPromptMember_(user) }]
    .concat(historyToChatMessages_(body.history))
    .concat([{ role: "user", content: String(body.message || "") }]);

  var tools = AI_TOOL_DEFS_MEMBER_.map(function (t) {
    return {
      type: "function",
      function: {
        name: t.name,
        description: t.description,
        parameters: t.parameters,
      },
    };
  });

  var rounds = 0;
  while (rounds < AI_CHAT_MAX_TOOL_ROUNDS) {
    rounds++;

    var payload = {
      model: OMNIROUTE_CONFIG.model,
      messages: messages,
      tools: tools,
      tool_choice: "auto",
      temperature: 0.3,
    };

    var response;
    try {
      response = UrlFetchApp.fetch(endpoint + "/chat/completions", {
        method: "post",
        contentType: "application/json",
        headers: {
          Authorization: apiKey ? "Bearer " + apiKey : "",
          "ngrok-skip-browser-warning": "true",
        },
        payload: JSON.stringify(payload),
        muteHttpExceptions: true,
        timeout: OMNIROUTE_CONFIG.timeout,
      });
    } catch (err) {
      return fail_("Gagal menghubungi asisten AI: " + err.message);
    }

    var status = response.getResponseCode();
    var responseText = response.getContentText();

    if (responseText.indexOf("<!DOCTYPE html>") === 0) {
      return fail_("Asisten AI mengembalikan respons tidak valid.");
    }

    if (status < 200 || status >= 300) {
      return fail_("Asisten AI error (HTTP " + status + ").");
    }

    var data;
    try {
      data = JSON.parse(responseText);
    } catch (e) {
      return fail_("Respons asisten AI tidak valid.");
    }

    var choice = data.choices && data.choices[0];
    var message = choice && choice.message;
    if (!message) return fail_("Asisten AI tidak memberi jawaban.");

    var toolCalls = message.tool_calls;
    if (toolCalls && toolCalls.length > 0) {
      messages.push({
        role: "assistant",
        content: message.content || null,
        tool_calls: toolCalls,
      });

      toolCalls.forEach(function (call) {
        var args = {};
        try {
          args = JSON.parse(call.function.arguments || "{}");
        } catch (e) {}
        var toolResult = executeAiToolMember_(call.function.name, args, ctx);
        messages.push({
          role: "tool",
          tool_call_id: call.id,
          content: JSON.stringify(toolResult),
        });
      });

      continue;
    }

    if (data.usage) {
      logAiUsage_(
        ctx,
        "omniroute",
        data.usage.prompt_tokens,
        data.usage.completion_tokens,
      );
    }

    incrementAiQuota_(ctx); // FASE 1 FIX: kuota AI

    return ok_({
      reply: message.content || "Maaf, saya tidak mendapatkan jawaban.",
      provider: "omniroute",
    });
  }

  return fail_("Terlalu banyak proses. Coba pertanyaan yang lebih spesifik.");
}

function geminiChat_(body, user, ctx) {
  if (!GEMINI_CONFIG.apiKey) {
    return fail_("GEMINI_API_KEY belum diatur.");
  }

  var url =
    GEMINI_CONFIG.endpoint +
    GEMINI_CONFIG.model +
    ":generateContent?key=" +
    encodeURIComponent(GEMINI_CONFIG.apiKey);

  var contents = historyToChatMessages_(body.history)
    .map(function (m) {
      return {
        role: m.role === "assistant" ? "model" : "user",
        parts: [{ text: m.content }],
      };
    })
    .concat([{ role: "user", parts: [{ text: String(body.message || "") }] }]);

  var tools = [
    {
      functionDeclarations: AI_TOOL_DEFS_.map(function (t) {
        return {
          name: t.name,
          description: t.description,
          parameters: t.parameters,
        };
      }),
    },
  ];

  var rounds = 0;
  while (rounds < AI_CHAT_MAX_TOOL_ROUNDS) {
    rounds++;

    var payload = {
      systemInstruction: { parts: [{ text: buildAiSystemPrompt_(user) }] },
      contents: contents,
      tools: tools,
      generationConfig: { temperature: 0.3 },
    };

    var response;
    try {
      response = UrlFetchApp.fetch(url, {
        method: "post",
        contentType: "application/json",
        payload: JSON.stringify(payload),
        muteHttpExceptions: true,
      });
    } catch (err) {
      return fail_("Gagal menghubungi Gemini: " + err.message);
    }

    var status = response.getResponseCode();
    if (status < 200 || status >= 300) {
      return fail_("Gemini error HTTP " + status);
    }

    var data;
    try {
      data = JSON.parse(response.getContentText());
    } catch (e) {
      return fail_("Respons Gemini bukan JSON.");
    }

    var candidate = data.candidates && data.candidates[0];
    var parts = candidate && candidate.content && candidate.content.parts;
    if (!parts || parts.length === 0) {
      return fail_("Respons Gemini kosong.");
    }

    var functionCalls = parts.filter(function (p) {
      return p.functionCall;
    });

    if (functionCalls.length > 0) {
      contents.push({
        role: "model",
        parts: functionCalls.map(function (p) {
          return {
            functionCall: p.functionCall,
            thoughtSignature: p.thoughtSignature,
          };
        }),
      });

      var functionResponseParts = functionCalls.map(function (p) {
        var toolResult = executeAiTool_(
          p.functionCall.name,
          p.functionCall.args || {},
          ctx,
        );
        return {
          functionResponse: { name: p.functionCall.name, response: toolResult },
        };
      });

      contents.push({ role: "user", parts: functionResponseParts });
      continue;
    }

    var textParts = parts.filter(function (p) {
      return typeof p.text === "string";
    });
    var reply = textParts
      .map(function (p) {
        return p.text;
      })
      .join("\n")
      .trim();

    if (data.usageMetadata) {
      logAiUsage_(
        ctx,
        "gemini",
        data.usageMetadata.promptTokenCount,
        data.usageMetadata.candidatesTokenCount,
      );
    }

    incrementAiQuota_(ctx); // FASE 1 FIX: kuota AI

    return ok_({ reply: reply || "Maaf, saya tidak mendapatkan jawaban." });
  }

  return fail_("Terlalu banyak proses pengambilan data.");
}

function groqChat_(body, user, ctx) {
  if (!GROQ_CONFIG.apiKey) {
    return fail_("GROQ_API_KEY belum diatur.");
  }

  var models = GROQ_CONFIG.models || [];
  var lastError = null;

  for (var mi = 0; mi < models.length; mi++) {
    var currentModel = models[mi];
    var result = groqChatWithModel_(body, user, currentModel, ctx);
    if (result.success) return result;
    if (!isQuotaError_(result.message)) return result;
    lastError = result;
  }

  return lastError || fail_("Semua model Groq quota habis.");
}

function groqChatMember_(body, user, ctx) {
  if (!GROQ_CONFIG.apiKey) {
    return fail_("Konfigurasi AI belum lengkap.");
  }

  var models = GROQ_CONFIG.models || [];
  var lastError = null;

  for (var mi = 0; mi < models.length; mi++) {
    var currentModel = models[mi];
    var result = groqChatMemberWithModel_(body, user, currentModel, ctx);
    if (result.success) return result;
    if (!isQuotaError_(result.message)) return result;
    lastError = result;
  }

  return lastError || fail_("Semua model AI quota habis.");
}

function groqChatWithModel_(body, user, modelName, ctx) {
  var endpoint = GROQ_CONFIG.endpoint;
  var apiKey = GROQ_CONFIG.apiKey;

  var messages = [{ role: "system", content: buildAiSystemPrompt_(user) }]
    .concat(historyToChatMessages_(body.history))
    .concat([{ role: "user", content: String(body.message || "") }]);

  var tools = AI_TOOL_DEFS_.map(function (t) {
    return {
      type: "function",
      function: {
        name: t.name,
        description: t.description,
        parameters: t.parameters,
      },
    };
  });

  var rounds = 0;
  while (rounds < AI_CHAT_MAX_TOOL_ROUNDS) {
    rounds++;

    var payload = {
      model: modelName,
      messages: messages,
      tools: tools,
      tool_choice: "auto",
      temperature: 0.3,
    };

    var response;
    try {
      response = UrlFetchApp.fetch(endpoint + "/chat/completions", {
        method: "post",
        contentType: "application/json",
        headers: { Authorization: "Bearer " + apiKey },
        payload: JSON.stringify(payload),
        muteHttpExceptions: true,
        timeout: GROQ_CONFIG.timeout,
      });
    } catch (err) {
      return fail_("Gagal menghubungi Groq: " + err.message);
    }

    var status = response.getResponseCode();
    var responseText = response.getContentText();

    if (status < 200 || status >= 300) {
      return fail_("Groq error HTTP " + status);
    }

    var data;
    try {
      data = JSON.parse(responseText);
    } catch (e) {
      return fail_("Respons Groq bukan JSON.");
    }

    var choice = data.choices && data.choices[0];
    var message = choice && choice.message;
    if (!message) return fail_("Respons Groq kosong.");

    var toolCalls = message.tool_calls;
    if (toolCalls && toolCalls.length > 0) {
      messages.push({
        role: "assistant",
        content: message.content || null,
        tool_calls: toolCalls,
      });

      toolCalls.forEach(function (call) {
        var args = {};
        try {
          args = JSON.parse(call.function.arguments || "{}");
        } catch (e) {}
        var toolResult = executeAiTool_(call.function.name, args, ctx);
        messages.push({
          role: "tool",
          tool_call_id: call.id,
          content: JSON.stringify(toolResult),
        });
      });

      continue;
    }

    if (data.usage) {
      logAiUsage_(
        ctx,
        "groq",
        data.usage.prompt_tokens,
        data.usage.completion_tokens,
      );
    }

    incrementAiQuota_(ctx); // FASE 1 FIX: kuota AI

    return ok_({
      reply: message.content || "Maaf, saya tidak mendapatkan jawaban.",
      model: modelName,
    });
  }

  return fail_("Terlalu banyak proses pengambilan data.");
}

function groqChatMemberWithModel_(body, user, modelName, ctx) {
  var endpoint = GROQ_CONFIG.endpoint;
  var apiKey = GROQ_CONFIG.apiKey;

  var messages = [{ role: "system", content: buildAiSystemPromptMember_(user) }]
    .concat(historyToChatMessages_(body.history))
    .concat([{ role: "user", content: String(body.message || "") }]);

  var tools = AI_TOOL_DEFS_MEMBER_.map(function (t) {
    return {
      type: "function",
      function: {
        name: t.name,
        description: t.description,
        parameters: t.parameters,
      },
    };
  });

  var rounds = 0;
  while (rounds < AI_CHAT_MAX_TOOL_ROUNDS) {
    rounds++;

    var payload = {
      model: modelName,
      messages: messages,
      tools: tools,
      tool_choice: "auto",
      temperature: 0.3,
    };

    var response;
    try {
      response = UrlFetchApp.fetch(endpoint + "/chat/completions", {
        method: "post",
        contentType: "application/json",
        headers: { Authorization: "Bearer " + apiKey },
        payload: JSON.stringify(payload),
        muteHttpExceptions: true,
        timeout: GROQ_CONFIG.timeout,
      });
    } catch (err) {
      return fail_("Gagal menghubungi asisten AI: " + err.message);
    }

    var status = response.getResponseCode();
    var responseText = response.getContentText();

    if (status < 200 || status >= 300) {
      return fail_("Asisten AI error HTTP " + status);
    }

    var data;
    try {
      data = JSON.parse(responseText);
    } catch (e) {
      return fail_("Respons asisten AI bukan JSON.");
    }

    var choice = data.choices && data.choices[0];
    var message = choice && choice.message;
    if (!message) return fail_("Respons asisten AI kosong.");

    var toolCalls = message.tool_calls;
    if (toolCalls && toolCalls.length > 0) {
      messages.push({
        role: "assistant",
        content: message.content || null,
        tool_calls: toolCalls,
      });

      toolCalls.forEach(function (call) {
        var args = {};
        try {
          args = JSON.parse(call.function.arguments || "{}");
        } catch (e) {}
        var toolResult = executeAiToolMember_(call.function.name, args, ctx);
        messages.push({
          role: "tool",
          tool_call_id: call.id,
          content: JSON.stringify(toolResult),
        });
      });

      continue;
    }

    if (data.usage) {
      logAiUsage_(
        ctx,
        "groq",
        data.usage.prompt_tokens,
        data.usage.completion_tokens,
      );
    }

    incrementAiQuota_(ctx); // FASE 1 FIX: kuota AI

    return ok_({
      reply: message.content || "Maaf, saya tidak mendapatkan jawaban.",
      provider: "groq",
    });
  }

  return fail_("Terlalu banyak proses. Coba pertanyaan yang lebih spesifik.");
}
