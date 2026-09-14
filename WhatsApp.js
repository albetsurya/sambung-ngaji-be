/* -------------------------------------------------------------------------- */
/*                            WhatsApp notification                           */
/* -------------------------------------------------------------------------- */

function sendWhatsApp_(noWa, message) {
  var enabled = String(getConfig_("WA_GATEWAY_ENABLED", "false")).toLowerCase();
  if (enabled !== "true") {
    Logger.log("WA gateway disabled, pesan tidak dikirim ke " + noWa);
    return { skipped: true };
  }

  var url = getConfig_("WA_GATEWAY_URL", "");
  var token = getConfig_("WA_GATEWAY_TOKEN", "");

  if (!url || !token) {
    Logger.log("WA gateway belum dikonfigurasi (url/token kosong)");
    return { skipped: true };
  }

  var normalized = normalizePhoneNumber(noWa);
  if (!normalized) {
    Logger.log("Nomor WA tidak valid: " + noWa);
    return { skipped: true };
  }

  var target =
    normalized.indexOf("0") === 0 ? "62" + normalized.slice(1) : normalized;

  var options = {
    method: "post",
    contentType: "application/x-www-form-urlencoded",
    headers: { Authorization: token },
    payload: { target: target, message: message, countryCode: "62" },
    muteHttpExceptions: true,
  };

  try {
    var res = UrlFetchApp.fetch(url, options);
    var code = res.getResponseCode();
    var text = res.getContentText();
    Logger.log("WA response (" + code + "): " + text);

    var body = {};
    try {
      body = JSON.parse(text);
    } catch (e) {
      Logger.log("Gagal parse body WA response");
    }

    if (code === 200 && body.status === true) {
      return { ok: true, body: body };
    }

    Logger.log(
      "WA gagal kirim ke " + target + ": " + (body.reason || "unknown"),
    );
    return { ok: false, reason: body.reason || "unknown", body: body };
  } catch (e) {
    Logger.log("Gagal kirim WA: " + e);
    return { ok: false, error: String(e) };
  }
}
