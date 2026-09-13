var MAX_PHOTO_SIZE_BYTES = 5 * 1024 * 1024;
var ALLOWED_PHOTO_TYPES = ['image/jpeg', 'image/png', 'image/webp'];



function uploadPhoto_(ctx, params) {
  if (!params.member_id) return fail_('member_id wajib diisi');
  if (!params.base64 || !params.mime_type) return fail_('File foto wajib diisi');
  if (ALLOWED_PHOTO_TYPES.indexOf(params.mime_type) === -1) return fail_('Tipe file harus JPEG/PNG/WEBP');

  var bytes;
  try { bytes = Utilities.base64Decode(params.base64); }
  catch (e) { return fail_('Data foto tidak valid'); }
  if (bytes.length > MAX_PHOTO_SIZE_BYTES) return fail_('Ukuran foto maksimum 5MB');

  var ext = params.mime_type === 'image/png' ? 'png' : (params.mime_type === 'image/webp' ? 'webp' : 'jpg');
  var fileName = params.member_id + '_' + new Date().getTime() + '.' + ext;
  var blob = Utilities.newBlob(bytes, params.mime_type, fileName);

  var folder = getPhotoFolder_();
  var archiveFolder = getPhotoArchiveFolder_();

  // ✅ PINDAH foto lama ke folder arsip (bukan hapus)
  var existingFiles = folder.getFiles();
  while (existingFiles.hasNext()) {
    var f = existingFiles.next();
    if (f.getName().indexOf(params.member_id + '_') === 0) {
      try {
        // Copy ke arsip
        archiveFolder.addFile(f);
        // Hapus dari folder utama
        folder.removeFile(f);
      } catch (e) {
        Logger.log("Gagal pindah file lama: " + e.message);
      }
    }
  }

  // Simpan foto baru di folder utama
  var file = folder.createFile(blob);
  file.setSharing(DriveApp.Access.ANYONE_WITH_LINK, DriveApp.Permission.VIEW);

  var url = 'https://drive.google.com/thumbnail?id=' + file.getId() + '&sz=w1000';

  var membersRepo = new SheetRepository_('members');
  membersRepo.updateById('member_id', params.member_id, { foto_url: url, updated_at: nowIso_() });

  writeAuditLog_(ctx.user.user_id, 'UPLOAD_PHOTO', 'MEMBER', params.member_id);
  return ok_({ foto_url: url });
}

function deletePhoto_(ctx, params) {
  var memberId;

  // Member bisa hapus foto sendiri
  if (ctx.user.role === ROLES.MEMBER) {
    memberId = ctx.user.member_id;
    if (!memberId) return fail_('Akun Anda belum terhubung ke data jamaah.');
  } else {
    // Admin bisa hapus foto siapa saja
    memberId = params.member_id;
    if (!memberId) return fail_('member_id wajib diisi');
  }

  var membersRepo = new SheetRepository_('members');
  var member = membersRepo.findById('member_id', memberId);
  if (!member) return fail_('Jamaah tidak ditemukan');

  if (!member.foto_url) {
    return fail_('Jamaah ini tidak memiliki foto');
  }

  // Hapus file di Drive
  try {
    var folder = getPhotoFolder_();
    var files = folder.getFiles();
    var deleted = 0;

    while (files.hasNext()) {
      var f = files.next();
      if (f.getName().indexOf(memberId + '_') === 0) {
        f.setTrashed(true);
        deleted++;
      }
    }

    Logger.log('Deleted ' + deleted + ' file(s) for member ' + memberId);
  } catch (e) {
    Logger.log('Gagal hapus file Drive: ' + e.message);
    // Lanjutkan — hapus dari sheet meskipun file Drive gagal
  }

  // Update sheet — kosongkan foto_url
  membersRepo.updateById('member_id', memberId, {
    foto_url: '',
    updated_at: nowIso_()
  });

  writeAuditLog_(ctx.user.user_id, 'DELETE_PHOTO', 'MEMBER', memberId);

  return ok_({ deleted: true });
}

/**
 * Hapus foto di folder arsip yang lebih tua dari X bulan.
 * Jalankan via trigger bulanan.
 */
function cleanupArchivedPhotos() {
  var ARCHIVE_MAX_MONTHS = 3;   // ← configurable

  Logger.log("=== CLEANUP ARCHIVED PHOTOS ===");
  Logger.log("Max age: " + ARCHIVE_MAX_MONTHS + " bulan");

  var folder = getPhotoArchiveFolder_();
  var files = folder.getFiles();

  var cutoffDate = new Date();
  cutoffDate.setMonth(cutoffDate.getMonth() - ARCHIVE_MAX_MONTHS);

  Logger.log("Cutoff: " + cutoffDate.toISOString());
  Logger.log("");

  var deleted = 0;
  var kept = 0;

  while (files.hasNext()) {
    var f = files.next();
    var dateCreated = f.getDateCreated();
    var name = f.getName();

    if (dateCreated < cutoffDate) {
      try {
        f.setTrashed(true);
        deleted++;
        Logger.log("🗑️  Deleted: " + name + " (dibuat: " + dateCreated.toISOString() + ")");
      } catch (e) {
        Logger.log("❌ Gagal hapus " + name + ": " + e.message);
      }
    } else {
      kept++;
    }
  }

  Logger.log("");
  Logger.log("═══════════════════════════════════════");
  Logger.log("Deleted: " + deleted);
  Logger.log("Kept:    " + kept);
  Logger.log("═══════════════════════════════════════");
}

function restoreArchivedPhoto_(ctx, params) {
  if (!params.member_id) return fail_('member_id wajib diisi');
  if (!params.file_id) return fail_('file_id wajib diisi');

  // Ambil file dari arsip
  var archiveFile = DriveApp.getFileById(params.file_id);
  if (!archiveFile) return fail_('File tidak ditemukan');

  // Pindah ke folder utama
  var folder = getPhotoFolder_();
  var archiveFolder = getPhotoArchiveFolder_();

  folder.addFile(archiveFile);
  archiveFolder.removeFile(archiveFile);

  // Set sharing
  archiveFile.setSharing(DriveApp.Access.ANYONE_WITH_LINK, DriveApp.Permission.VIEW);

  // Update URL di sheet
  var url = 'https://drive.google.com/thumbnail?id=' + archiveFile.getId() + '&sz=w1000';
  var membersRepo = new SheetRepository_('members');
  membersRepo.updateById('member_id', params.member_id, { foto_url: url, updated_at: nowIso_() });

  writeAuditLog_(ctx.user.user_id, 'RESTORE_PHOTO', 'MEMBER', params.member_id);

  return ok_({ foto_url: url });
}

function getArchivedPhotos_(ctx, params) {
  if (!params.member_id) return fail_('member_id wajib diisi');

  var archiveFolder = getPhotoArchiveFolder_();
  var files = archiveFolder.getFiles();

  var result = [];

  while (files.hasNext()) {
    var f = files.next();
    if (f.getName().indexOf(params.member_id + '_') === 0) {
      result.push({
        file_id: f.getId(),
        name: f.getName(),
        created_at: f.getDateCreated().toISOString(),
        url: 'https://drive.google.com/thumbnail?id=' + f.getId() + '&sz=w1000'
      });
    }
  }

  // Sort by created date, newest first
  result.sort(function (a, b) {
    return new Date(b.created_at) - new Date(a.created_at);
  });

  return ok_(result);
}

function changeMyPassword_(ctx, params) {
  var oldPassword = String(params.old_password || '');
  var newPassword = String(params.new_password || '');

  if (!oldPassword) return fail_('Password lama wajib diisi');
  if (!newPassword) return fail_('Password baru wajib diisi');
  if (newPassword.length < 6) return fail_('Password baru minimal 6 karakter');
  if (newPassword === oldPassword) {
    return fail_('Password baru harus berbeda dari password lama');
  }

  var usersRepo = new SheetRepository_('users');
  var user = usersRepo.findById('user_id', ctx.user.user_id);
  if (!user) return fail_('User tidak ditemukan');

  if (!verifyPassword_(oldPassword, user.password_hash)) {
    return fail_('Password lama salah');
  }

  usersRepo.updateById('user_id', ctx.user.user_id, {
    password_hash: hashPassword_(newPassword),
    updated_at: nowIso_()
  });

  writeAuditLog_(ctx.user.user_id, 'CHANGE_PASSWORD', 'USER', ctx.user.user_id);

  return ok_({ changed: true });
}

function resetUserPassword_(ctx, params) {
  var userId = String(params.user_id || '');
  var newPassword = String(params.new_password || '');

  if (!userId) return fail_('user_id wajib diisi');
  if (!newPassword) return fail_('Password baru wajib diisi');
  if (newPassword.length < 6) return fail_('Password baru minimal 6 karakter');

  var usersRepo = new SheetRepository_('users');
  var user = usersRepo.findById('user_id', userId);
  if (!user) return fail_('User tidak ditemukan');

  usersRepo.updateById('user_id', userId, {
    password_hash: hashPassword_(newPassword),
    updated_at: nowIso_()
  });

  writeAuditLog_(ctx.user.user_id, 'RESET_USER_PASSWORD', 'USER', userId);

  return ok_({ reset: true, user_id: userId });
}