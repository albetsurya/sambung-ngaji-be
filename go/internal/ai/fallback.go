package ai

import (
	"strings"
)

// IsModelError: error yang artinya "model ini tidak bisa dipakai" —
// model tidak ada / dipensiunkan / tidak tersedia / akses ditolak.
// Beda dengan error sementara (network/5xx) yang di-handle Retry.
func IsModelError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, kw := range []string{
		"404",
		"403",
		"model_not_found",
		"not_found",
		"not available",
		"no longer available",
		"does not exist",
		"invalid model",
		"permission_denied",
		"permission denied",
		"denied access",
		"forbidden",
	} {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

// IsFallbackable: error boleh lanjut ke model/provider berikutnya.
// Kuota habis ATAU model tidak bisa dipakai → coba yang lain.
// Error lain (network, format, dsb) → langsung gagal supaya cepat ketahuan.
func IsFallbackable(err error) bool {
	return IsQuotaError(err) || IsModelError(err)
}

// dedupModels: buang duplikat + string kosong, pertahankan urutan.
func dedupModels(models []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out
}
