// Di Apps Script editor, jalankan manual:
function testAttendanceCache() {
  var t1 = Date.now();
  getAttendanceByMeeting_("MTG-XXX"); // ganti dengan meeting_id real
  Logger.log("First call: " + (Date.now() - t1) + "ms");

  var t2 = Date.now();
  getAttendanceByMeeting_("MTG-XXX");
  Logger.log("Cached call: " + (Date.now() - t2) + "ms"); // harus <100ms
}

function testInvalidate() {
  getAttendanceByMeeting_("MTG-XXX"); // cache
  invalidateAttendanceCache_("MTG-XXX");
  var cached = CacheService.getScriptCache().get("att_MTG-XXX");
  Logger.log("After invalidate: " + (cached === null ? "OK" : "FAIL"));
}

function testGetAllRecent() {
  // DEBUG: cek spreadsheet apa yang dipakai
  var ss = getSpreadsheet_();
  Logger.log("Spreadsheet name: " + ss.getName());
  Logger.log("Spreadsheet ID: " + ss.getId());
  Logger.log(
    "All sheets: " +
      ss
        .getSheets()
        .map(function (s) {
          return s.getName();
        })
        .join(", "),
  );

  var sheet = ss.getSheetByName("attendance");
  Logger.log("attendance sheet exists: " + (sheet !== null));
  if (sheet) {
    Logger.log("attendance lastRow: " + sheet.getLastRow());
    Logger.log("attendance lastCol: " + sheet.getLastColumn());
  }

  var repo = new SheetRepository_("attendance");

  var t1 = Date.now();
  var all = repo.getAll();
  Logger.log(
    "getAll(): " + all.length + " rows in " + (Date.now() - t1) + "ms",
  );

  var t2 = Date.now();
  var recent = repo.getAllRecent(3000);
  Logger.log(
    "getAllRecent(3000): " +
      recent.length +
      " rows in " +
      (Date.now() - t2) +
      "ms",
  );
}

function testBatchDelete() {
  // 1. Buat meeting dummy dengan banyak attendance
  // 2. Ukur waktu delete
  var t1 = Date.now();
  var result = deleteAttendanceByMeeting_(null, { meeting_id: "MTG-XXX" });
  Logger.log(
    "Deleted " + result.data.deleted + " rows in " + (Date.now() - t1) + "ms",
  );
}

function testBatchDeleteReal() {
  var meetingId = "MTGFD913B20";

  // 1. Cek dulu berapa row
  var before = new SheetRepository_("attendance").findByField(
    "meeting_id",
    meetingId,
  );
  Logger.log("Before: " + before.length + " rows");

  // 2. Ukur waktu delete — pakai ctx dummy
  var dummyCtx = { user: { user_id: "TEST", role: "SUPER_ADMIN" } };
  var t1 = Date.now();
  var result = deleteAttendanceByMeeting_(dummyCtx, { meeting_id: meetingId });
  var elapsed = Date.now() - t1;

  Logger.log("Deleted: " + result.data.deleted + " rows in " + elapsed + "ms");

  // 3. Cek sisa
  var after = new SheetRepository_("attendance").findByField(
    "meeting_id",
    meetingId,
  );
  Logger.log("After: " + after.length + " rows");
}

function generateDummyMembers() {
  var TOTAL = 1000;
  var membersRepo = new SheetRepository_("members");
  var now = nowIso_();

  var namaDepanL = [
    "Ahmad",
    "Budi",
    "Choirul",
    "Dedi",
    "Eko",
    "Fajar",
    "Gunawan",
    "Hadi",
    "Irfan",
    "Joko",
    "Krisna",
    "Lukman",
    "Muhammad",
    "Nur",
    "Omar",
    "Putra",
    "Rahmat",
    "Slamet",
    "Taufik",
    "Umar",
  ];
  var namaDepanP = [
    "Aisyah",
    "Bella",
    "Citra",
    "Dewi",
    "Eka",
    "Fitri",
    "Gita",
    "Hana",
    "Indah",
    "Jihan",
    "Kartika",
    "Lina",
    "Maya",
    "Nadia",
    "Oktaviani",
    "Putri",
    "Rina",
    "Sari",
    "Tina",
    "Umi",
  ];
  var namaBelakang = [
    "Wijaya",
    "Santoso",
    "Kusuma",
    "Pratama",
    "Hidayat",
    "Nugroho",
    "Setiawan",
    "Firmansyah",
    "Ramadhan",
    "Maulana",
    "Saputra",
    "Hartono",
    "Susanto",
    "Gunadi",
    "Wibowo",
    "Rahmawati",
    "Anggraini",
    "Puspita",
    "Lestari",
    "Safitri",
  ];

  var kelompokList = ["GRP_A", "GRP_B", "GRP_C", "GRP_D", "GRP_E"];
  var desaList = [
    "Desa Sukamaju",
    "Desa Mekarsari",
    "Desa Banyuwangi",
    "Desa Wonorejo",
    "Desa Sidoarjo",
  ];
  var daerahList = [
    "Kecamatan A",
    "Kecamatan B",
    "Kecamatan C",
    "Kecamatan D",
    "Kecamatan E",
  ];
  var pekerjaanList = [
    "Karyawan",
    "Wiraswasta",
    "Mahasiswa",
    "Pelajar",
    "Guru",
    "Petani",
    "Buruh",
    "PNS",
  ];
  var hobiList = [
    "Membaca",
    "Olahraga",
    "Traveling",
    "Memasak",
    "Berkebun",
    "Menulis",
    "Musik",
    "Fotografi",
  ];
  var jenjangList = ["SD", "SMP", "SMA", "SMK", "S1", "S2"];
  var sekolahList = [
    "SDN 1",
    "SMPN 2",
    "SMAN 3",
    "SMKN 4",
    "Universitas A",
    "Universitas B",
  ];
  var jurusanList = [
    "IPA",
    "IPS",
    "Teknik Informatika",
    "Ekonomi",
    "Hukum",
    "Kedokteran",
    "Psikologi",
  ];
  var statusPembinaanList = [
    "AKTIF",
    "AKTIF",
    "AKTIF",
    "AKTIF",
    "PERLU_PERHATIAN",
    "KURANG_AKTIF",
  ];

  var genderOptions = ["L", "P"];

  var batchSize = 100;
  var batches = Math.ceil(TOTAL / batchSize);
  var inserted = 0;

  for (var b = 0; b < batches; b++) {
    var batch = [];
    var batchCount = Math.min(batchSize, TOTAL - inserted);

    for (var i = 0; i < batchCount; i++) {
      var idx = inserted + i + 1;
      var gender = genderOptions[Math.floor(Math.random() * 2)];
      var namaDepan =
        gender === "L"
          ? namaDepanL[Math.floor(Math.random() * namaDepanL.length)]
          : namaDepanP[Math.floor(Math.random() * namaDepanP.length)];
      var namaBelakang =
        namaBelakang[Math.floor(Math.random() * namaBelakang.length)];
      var namaLengkap = "DUMMY_" + namaDepan + " " + namaBelakang + " " + idx;

      var usia = 5 + Math.floor(Math.random() * 60);
      var tahunLahir = new Date().getFullYear() - usia;
      var bulanLahir = 1 + Math.floor(Math.random() * 12);
      var tanggalLahir = 1 + Math.floor(Math.random() * 28);
      var tanggalLahirStr =
        tahunLahir +
        "-" +
        (bulanLahir < 10 ? "0" : "") +
        bulanLahir +
        "-" +
        (tanggalLahir < 10 ? "0" : "") +
        tanggalLahir;

      var jenjang = jenjangList[Math.floor(Math.random() * jenjangList.length)];

      batch.push({
        member_id: generateMemberId(),
        nama_lengkap: namaLengkap,
        nama_panggilan: namaDepan,
        jenis_kelamin: gender,
        tempat_lahir: "Kota " + String.fromCharCode(65 + (idx % 26)),
        tanggal_lahir: tanggalLahirStr,
        foto_url: "",
        no_wa: "62812" + String(100000000 + idx).slice(-8),
        alamat_rumah:
          "Jl. Dummy No. " +
          idx +
          ", RT " +
          (1 + (idx % 10)) +
          "/RW " +
          (1 + (idx % 5)),
        desa: desaList[idx % desaList.length],
        daerah: daerahList[idx % daerahList.length],
        kelompok: kelompokList[idx % kelompokList.length],
        is_muballigh: idx % 20 === 0,
        is_kerja: usia >= 17 && idx % 3 === 0,
        is_nikah: usia >= 22 && idx % 4 === 0,
        tinggi_badan:
          usia >= 10 ? String(150 + Math.floor(Math.random() * 30)) : "",
        berat_badan:
          usia >= 10 ? String(40 + Math.floor(Math.random() * 40)) : "",
        hobi: hobiList[idx % hobiList.length],
        pekerjaan:
          usia >= 17 ? pekerjaanList[idx % pekerjaanList.length] : "Pelajar",
        status_pembinaan: statusPembinaanList[idx % statusPembinaanList.length],
        status_aktif: true,
        tanggal_masuk: formatDate(now),
        tanggal_keluar: "",
        jenjang_pendidikan: jenjang,
        sekolah: sekolahList[idx % sekolahList.length],
        jurusan: usia >= 17 ? jurusanList[idx % jurusanList.length] : "",
        tahun_mulai_pendidikan: String(tahunLahir + 6),
        tahun_selesai_pendidikan: String(tahunLahir + 6 + 6),
        created_at: now,
        updated_at: now,
      });
    }

    membersRepo.insertMany(batch);
    inserted += batchCount;
    Logger.log(
      "Batch " +
        (b + 1) +
        "/" +
        batches +
        " selesai (" +
        inserted +
        "/" +
        TOTAL +
        ")",
    );
  }

  Logger.log("=== SELESAI ===");
  Logger.log("Total dummy member: " + inserted);

  return { total: inserted };
}

function deleteDummyMembers() {
  var membersRepo = new SheetRepository_("members");
  var sheet = membersRepo._sheet();
  var headers = membersRepo.def.headers;
  var lastRow = sheet.getLastRow();

  if (lastRow < 2) {
    Logger.log("Sheet kosong");
    return { deleted: 0 };
  }

  var nameColIndex = headers.indexOf("nama_lengkap");
  var values = sheet.getRange(2, nameColIndex + 1, lastRow - 1, 1).getValues();

  var rowsToDelete = [];
  for (var i = 0; i < values.length; i++) {
    var nama = String(values[i][0] || "");
    if (nama.indexOf("DUMMY_") === 0) {
      rowsToDelete.push(i + 2);
    }
  }

  if (rowsToDelete.length === 0) {
    Logger.log("Tidak ada dummy member");
    return { deleted: 0 };
  }

  rowsToDelete.sort(function (a, b) {
    return a - b;
  });

  var groups = [];
  var start = rowsToDelete[0];
  var prev = rowsToDelete[0];
  for (var j = 1; j < rowsToDelete.length; j++) {
    var curr = rowsToDelete[j];
    if (curr === prev + 1) {
      prev = curr;
    } else {
      groups.push({ start: start, count: prev - start + 1 });
      start = curr;
      prev = curr;
    }
  }
  groups.push({ start: start, count: prev - start + 1 });

  groups.sort(function (a, b) {
    return b.start - a.start;
  });
  groups.forEach(function (g) {
    sheet.deleteRows(g.start, g.count);
  });

  membersRepo._invalidateCache();

  Logger.log("=== SELESAI ===");
  Logger.log("Dummy member dihapus: " + rowsToDelete.length);

  return { deleted: rowsToDelete.length };
}

function countDummyMembers() {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();
  var dummy = all.filter(function (m) {
    return String(m.nama_lengkap || "").indexOf("DUMMY_") === 0;
  });

  Logger.log("Total member: " + all.length);
  Logger.log("Dummy member: " + dummy.length);
  Logger.log("Member asli: " + (all.length - dummy.length));

  return {
    total: all.length,
    dummy: dummy.length,
    real: all.length - dummy.length,
  };
}

function testGetMembers() {
  var membersRepo = new SheetRepository_("members");
  var all = membersRepo.getAll();
  Logger.log("Total members: " + all.length);
  Logger.log(
    "Dummy members: " +
      all.filter(function (m) {
        return String(m.nama_lengkap || "").indexOf("DUMMY_") === 0;
      }).length,
  );

  var active = all.filter(function (m) {
    return toBool_(m.status_aktif);
  });
  Logger.log("Active members: " + active.length);
}
