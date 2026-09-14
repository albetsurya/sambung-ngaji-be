function getConfig_(key, fallback) {
  try {
    var props = PropertiesService.getScriptProperties();
    var value = props.getProperty(key);
    if (value !== null && value !== undefined && value !== "") {
      return value;
    }
  } catch (e) {
    Logger.log("Error baca property " + key + ": " + e);
  }
  return fallback || "";
}

function getSpreadsheetId_() {
  return getConfig_(PROP_KEY_SPREADSHEET_ID, SPREADSHEET_ID);
}

function getDriveFolderId_() {
  return getConfig_(PROP_KEY_DRIVE_FOLDER_ID, DRIVE_FOLDER_ID);
}

function getDriveArchiveFolderId_() {
  return getConfig_(PROP_KEY_DRIVE_ARCHIVE_FOLDER_ID, DRIVE_ARCHIVE_FOLDER_ID);
}

function getSpreadsheet_() {
  var id = getSpreadsheetId_();
  return id
    ? SpreadsheetApp.openById(id)
    : SpreadsheetApp.getActiveSpreadsheet();
}

function getPhotoFolder_() {
  var id = getDriveFolderId_();
  if (!id) {
    throw new Error("DRIVE_FOLDER_ID belum diatur di Script Properties");
  }
  return DriveApp.getFolderById(id);
}

function getPhotoArchiveFolder_() {
  var id = getDriveArchiveFolderId_();
  if (!id) {
    throw new Error(
      "DRIVE_ARCHIVE_FOLDER_ID belum diatur di Script Properties",
    );
  }
  return DriveApp.getFolderById(id);
}

function shouldUseLock_(action) {
  return WRITE_ACTIONS[action] === true;
}

var PUBLIC_ACTIONS = {
  login: true,
  submitPublicRegistration: true,
  checkUsernameAvailability: true,
};

var ACTION_HANDLERS = {
  login: function (ctx, p) {
    return login_(p);
  },
  logout: function (ctx, p) {
    return logout_(ctx);
  },
  validateSession: function (ctx, p) {
    return ok_({ user: publicUser_(ctx.user) });
  },

  aiChat: function (ctx, p) {
    var body = {
      message: p.message,
      history: p.history ? JSON.parse(p.history) : [],
      provider: p.provider,
    };
    return handleAiChat_(body, ctx);
  },

  getMemberUserStatus: function (ctx, p) {
    if (!p.member_id) return fail_("member_id wajib diisi");
    return ok_(getMemberUserStatus_(p.member_id));
  },

  getCurrentProvider: function (ctx, p) {
    return getCurrentProvider();
  },
  setAIProvider: function (ctx, p) {
    return setAIProvider(p.provider);
  },

  getDashboard: getDashboard_,
  getMyDashboard: getMyDashboard_,

  changeMyPassword: changeMyPassword_,
  resetUserPassword: resetUserPassword_,

  getAiUsageStats: getAiUsageStats_,

  getMembers: getMembers_,
  getMembersPaged: getMembersPaged_,
  getPNKBMembers: getPNKBMembers_,
  getPNKBMembersPaged: getPNKBMembersPaged_,
  getAttendanceMembers: getAttendanceMembers_,
  getMemberDetail: getMemberDetail_,
  createMember: createMember_,
  updateMember: updateMember_,
  deactivateMember: deactivateMember_,

  getGroups: getGroups_,
  saveGroup: saveGroup_,

  getMeetings: getMeetings_,
  createMeeting: createMeeting_,
  updateMeeting: updateMeeting_,

  getAttendance: getAttendance_,
  saveAttendance: saveAttendance_,
  bulkSaveAttendance: bulkSaveAttendance_,

  deleteAttendance: deleteAttendance_,
  deleteAttendanceByMeeting: deleteAttendanceByMeeting_,
  deleteAttendanceByMember: deleteAttendanceByMember_,

  getMonitoring: getMonitoring_,
  createMonitoring: createMonitoring_,
  updateMonitoring: updateMonitoring_,

  getAnnouncementTemplates: getAnnouncementTemplates_,
  generateAnnouncement: generateAnnouncement_,
  generateWeeklyAnnouncements: generateWeeklyAnnouncements_,
  createAnnouncement: createAnnouncement_,
  updateAnnouncement: updateAnnouncement_,
  getAnnouncements: getAnnouncements_,
  getAnnouncementRecipientSummary: getAnnouncementRecipientSummary_,

  uploadPhoto: uploadPhoto_,
  deletePhoto: deletePhoto_,

  getUsers: getUsers_,
  getUserDetail: getUserDetail_,
  createUser: createUser_,
  updateUser: updateUser_,
  updateUserRole: updateUserRole_,

  getSettings: getSettings_,
  updateSettings: updateSettings_,

  getAuditLogs: getAuditLogs_,

  getMyProfile: getMyProfile_,
  updateMyProfile: updateMyProfile_,
  getMyAttendance: getMyAttendance_,
  getMyMonitoring: getMyMonitoring_,
  getUpcomingMeetings: getUpcomingMeetings_,

  submitPublicRegistration: submitPublicRegistration_,
  checkUsernameAvailability: checkUsernameAvailability_,
  getPendingMembers: getPendingMembers_,
  getPendingMemberDetail: getPendingMemberDetail_,
  approvePendingMember: approvePendingMember_,
  rejectPendingMember: rejectPendingMember_,
};

var SUPER_ADMIN_ONLY_ACTIONS = {
  getUsers: true,
  getUserDetail: true,
  createUser: true,
  updateUser: true,
  updateUserRole: true,
  getAuditLogs: true,
  updateSettings: true,
  resetUserPassword: true,
};

function doGet(e) {
  return handleRequest_(e);
}

function doPost(e) {
  return handleRequest_(e);
}

function handleRequest_(e) {
  var params = parseParams_(e);
  var action = params.action;
  if (!action) return jsonOutput_(fail_("Parameter action wajib diisi"));

  var handler = ACTION_HANDLERS[action];
  if (!handler) return jsonOutput_(fail_("Action tidak dikenal: " + action));

  var ctx = {};
  if (!PUBLIC_ACTIONS[action]) {
    ctx = validateSession_(params.token);
    if (!ctx)
      return jsonOutput_(
        fail_("Unauthorized: sesi tidak valid atau kadaluarsa"),
      );

    if (
      SUPER_ADMIN_ONLY_ACTIONS[action] &&
      ctx.user.role !== ROLES.SUPER_ADMIN
    ) {
      return jsonOutput_(fail_("Forbidden: hanya SUPER_ADMIN"));
    }
    if (!checkPermission_(ctx.user, action)) {
      return jsonOutput_(
        fail_(
          "Forbidden: role " +
            ctx.user.role +
            " tidak memiliki akses ke " +
            action,
        ),
      );
    }
  }

  var lock = null;
  var hasLock = true;

  if (shouldUseLock_(action)) {
    lock = LockService.getScriptLock();
    hasLock = lock.tryLock(30000);
    if (!hasLock) {
      return jsonOutput_(fail_("Server sedang sibuk. Coba lagi."));
    }
  }

  try {
    var result = handler(ctx, params);
    return jsonOutput_(result);
  } catch (err) {
    Logger.log("handleRequest_ error: " + err + "\n" + (err && err.stack));
    return jsonOutput_(fail_("Terjadi kesalahan server: " + err));
  } finally {
    if (lock && hasLock) {
      try {
        lock.releaseLock();
      } catch (e) {}
    }
  }
}

function parseParams_(e) {
  var params = {};

  if (e && e.parameter) {
    Object.keys(e.parameter).forEach(function (k) {
      params[k] = e.parameter[k];
    });
  }

  if (e && e.postData && e.postData.contents) {
    var contents = e.postData.contents;
    var type = String(e.postData.type || "").toLowerCase();

    if (contents.charAt(0) === "{") {
      try {
        var body = JSON.parse(contents);
        Object.keys(body).forEach(function (k) {
          params[k] = body[k];
        });
      } catch (err) {
        Logger.log("parseParams_ JSON error: " + err);
      }
    } else if (type.indexOf("application/x-www-form-urlencoded") === 0) {
      try {
        contents.split("&").forEach(function (pair) {
          var idx = pair.indexOf("=");
          if (idx > 0) {
            var key = decodeURIComponent(
              pair.slice(0, idx).replace(/\+/g, " "),
            );
            var val = decodeURIComponent(
              pair.slice(idx + 1).replace(/\+/g, " "),
            );
            params[key] = val;
          }
        });
      } catch (err) {
        Logger.log("parseParams_ urlencoded error: " + err);
      }
    }
  }

  if (!params.action && e && e.parameter) {
    Object.keys(e.parameter).forEach(function (k) {
      var v = String(e.parameter[k]);
      if (v.charAt(0) === "{") {
        try {
          var parsed = JSON.parse(v);
          Object.keys(parsed).forEach(function (pk) {
            if (!params[pk]) params[pk] = parsed[pk];
          });
        } catch (e) {}
      }
    });
  }

  return params;
}
