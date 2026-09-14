/* -------------------------------------------------------------------------- */
/*                            WhatsApp notification                           */
/* -------------------------------------------------------------------------- */

/**
 * Kirim pesan WhatsApp via gateway yang dikonfigurasi.
 * Provider: Fonnte (default). Ganti sesuai provider yang dipakai.
 *
 * Script Properties yang dibutuhkan:
 *   WA_GATEWAY_URL   -> endpoint provider (mis. https://api.fonnte.com/send)
 *   WA_GATEWAY_TOKEN -> API key / token
 *   WA_GATEWAY_ENABLED -> "true" untuk aktifkan
 */
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

  // Fonnte menerima format 62812xxx
  var target =
    normalized.indexOf("0") === 0 ? "62" + normalized.slice(1) : normalized;

  var payload = {
    target: target,
    message: message,
    countryCode: "62",
  };

  var options = {
    method: "post",
    contentType: "application/x-www-form-urlencoded",
    headers: { Authorization: token },
    payload: payload,
    muteHttpExceptions: true,
  };

  try {
    var res = UrlFetchApp.fetch(url, options);
    var body = res.getContentText();
    Logger.log("WA response (" + res.getResponseCode() + "): " + body);
    return { ok: res.getResponseCode() === 200, body: body };
  } catch (e) {
    Logger.log("Gagal kirim WA: " + e);
    return { ok: false, error: String(e) };
  }
}
