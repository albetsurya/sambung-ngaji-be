var CACHE_TTL_SECONDS = 600;

var CACHEABLE_SHEETS = {
  users: true,
  members: true,
  groups: true,
  meetings: true,
  monitoring: true,
  announcements: true,
  announcement_templates: true,
  settings: true,
};

function SheetRepository_(sheetKey) {
  this.key = sheetKey;
  this.def = SHEETS[sheetKey];
  if (!this.def) throw new Error("Sheet tidak dikenal: " + sheetKey);
}

SheetRepository_.prototype._sheet = function () {
  var ss = getSpreadsheet_();
  var sheet = ss.getSheetByName(this.def.name);
  if (!sheet)
    throw new Error(
      "Sheet belum dibuat: " +
        this.def.name +
        ". Jalankan setupSpreadsheet() dulu.",
    );
  return sheet;
};

SheetRepository_.prototype._rowsToObjects = function (values) {
  var headers = this.def.headers;
  var out = [];
  for (var i = 1; i < values.length; i++) {
    var row = values[i];
    if (row.join("") === "") continue;
    var obj = {};
    for (var c = 0; c < headers.length; c++) {
      obj[headers[c]] = row[c];
    }
    obj._row = i + 1;
    out.push(obj);
  }
  return out;
};

SheetRepository_.prototype.getAll = function () {
  if (CACHEABLE_SHEETS[this.key]) {
    var cached = CacheService.getScriptCache().get("sheet_" + this.key);
    if (cached) {
      try {
        return JSON.parse(cached);
      } catch (e) {}
    }
  }

  var sheet = this._sheet();
  var lastRow = sheet.getLastRow();
  var lastCol = this.def.headers.length;
  if (lastRow < 2) return [];

  var values = sheet.getRange(1, 1, lastRow, lastCol).getValues();
  var objects = this._rowsToObjects(values);

  if (CACHEABLE_SHEETS[this.key]) {
    try {
      var str = JSON.stringify(objects);
      if (str.length < 95000) {
        CacheService.getScriptCache().put(
          "sheet_" + this.key,
          str,
          CACHE_TTL_SECONDS,
        );
      }
    } catch (e) {}
  }
  return objects;
};

SheetRepository_.prototype.getAllRecent = function (limit) {
  var sheet = this._sheet();
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return [];

  var headers = this.def.headers;
  var startRow = Math.max(2, lastRow - limit + 1);
  var numRows = lastRow - startRow + 1;

  var values = sheet.getRange(startRow, 1, numRows, headers.length).getValues();
  return this._rowsToObjects(values);
};

SheetRepository_.prototype._invalidateCache = function () {
  if (CACHEABLE_SHEETS[this.key]) {
    CacheService.getScriptCache().remove("sheet_" + this.key);
  }
};

SheetRepository_.prototype.findById = function (idField, idValue) {
  var all = this.getAll();
  for (var i = 0; i < all.length; i++) {
    if (all[i][idField] === idValue) return all[i];
  }
  return null;
};

SheetRepository_.prototype.find = function (filterFn) {
  return this.getAll().filter(filterFn);
};

SheetRepository_.prototype.insert = function (obj) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var row = headers.map(function (h) {
    if (!obj.hasOwnProperty(h)) return "";
    var val = obj[h];
    if (val === null || val === undefined) return "";
    if (typeof val === "number" || typeof val === "boolean") {
      return "'" + String(val);
    }
    return val;
  });
  sheet.appendRow(row);
  this._invalidateCache();
  return obj;
};

SheetRepository_.prototype.insertMany = function (objs) {
  if (!objs || !objs.length) return objs;
  var sheet = this._sheet();
  var headers = this.def.headers;
  var rows = objs.map(function (obj) {
    return headers.map(function (h) {
      if (!obj.hasOwnProperty(h)) return "";
      var val = obj[h];
      if (val === null || val === undefined) return "";
      if (typeof val === "number" || typeof val === "boolean") {
        return "'" + String(val);
      }
      return val;
    });
  });

  var startRow = sheet.getLastRow() + 1;
  sheet.getRange(startRow, 1, rows.length, headers.length).setValues(rows);
  this._invalidateCache();
  return objs;
};

SheetRepository_.prototype.updateById = function (idField, idValue, patch) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return null;
  var idColIndex = headers.indexOf(idField);
  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();

  for (var i = 0; i < idColValues.length; i++) {
    if (idColValues[i][0] === idValue) {
      var rowNum = i + 2;
      var currentRow = sheet
        .getRange(rowNum, 1, 1, headers.length)
        .getValues()[0];
      var current = {};
      headers.forEach(function (h, idx) {
        current[h] = currentRow[idx];
      });
      var updated = Object.assign({}, current, patch);
      var newRow = headers.map(function (h) {
        return updated.hasOwnProperty(h) ? updated[h] : "";
      });
      sheet.getRange(rowNum, 1, 1, headers.length).setValues([newRow]);
      this._invalidateCache();
      return updated;
    }
  }
  return null;
};

SheetRepository_.prototype.findByField = function (field, value) {
  return this.find(function (o) {
    return o[field] === value;
  });
};

SheetRepository_.prototype.deleteById = function (idField, idValue) {
  var sheet = this._sheet();
  var headers = this.def.headers;
  var lastRow = sheet.getLastRow();
  if (lastRow < 2) return false;
  var idColIndex = headers.indexOf(idField);
  var idColValues = sheet
    .getRange(2, idColIndex + 1, lastRow - 1, 1)
    .getValues();

  for (var i = 0; i < idColValues.length; i++) {
    if (idColValues[i][0] === idValue) {
      sheet.deleteRow(i + 2);
      this._invalidateCache();
      return true;
    }
  }
  return false;
};

function batchUpdateRows_(sheet, headers, updates) {
  if (!updates || !updates.length) return 0;

  updates.sort(function (a, b) {
    return a.rowNum - b.rowNum;
  });

  var groups = [];
  var current = null;

  updates.forEach(function (u) {
    if (current && u.rowNum === current.endRow + 1) {
      current.endRow = u.rowNum;
      current.items.push(u);
    } else {
      if (current) groups.push(current);
      current = { startRow: u.rowNum, endRow: u.rowNum, items: [u] };
    }
  });
  if (current) groups.push(current);

  var totalUpdated = 0;

  groups.forEach(function (g) {
    var numRows = g.endRow - g.startRow + 1;
    var range = sheet.getRange(g.startRow, 1, numRows, headers.length);
    var values = range.getValues();

    g.items.forEach(function (u) {
      var idx = u.rowNum - g.startRow;
      var row = values[idx];
      headers.forEach(function (h, c) {
        if (u.patch.hasOwnProperty(h)) {
          row[c] = u.patch[h];
        }
      });
      totalUpdated++;
    });

    range.setValues(values);
  });

  return totalUpdated;
}
