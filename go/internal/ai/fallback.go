package ai

import (
	"strings"
)

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

func IsFallbackable(err error) bool {
	return IsQuotaError(err) || IsModelError(err)
}

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
