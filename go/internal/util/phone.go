package util

import "strings"

// NormalizePhone: port dari normalizePhoneNumber di utils.js.
func NormalizePhone(raw string) string {
	if raw == "" {
		return ""
	}
	var digits strings.Builder
	for _, c := range raw {
		if c >= '0' && c <= '9' {
			digits.WriteRune(c)
		}
	}
	s := digits.String()
	if s == "" {
		return ""
	}
	if s[0] == '0' {
		s = "62" + s[1:]
	}
	if !strings.HasPrefix(s, "62") {
		s = "62" + s
	}
	return s
}

// BuildSapaan: sapaan berdasarkan usia + jenis kelamin.
func BuildSapaan(jenisKelamin string, tanggalLahir interface{}) string {
	usia := -1
	if t, ok := tanggalLahir.(interface{ Year() int }); ok {
		_ = t
	}
	// Caller pakai GetAge, ini stub
	jk := strings.ToUpper(jenisKelamin)
	if usia >= 0 && usia < 13 {
		return "Adik"
	}
	if usia >= 40 {
		if jk == "P" {
			return "Ibu"
		}
		return "Bapak"
	}
	if jk == "P" {
		return "Saudari"
	}
	return "Saudara"
}

func BuildDoa(jenisKelamin string) string {
	if strings.ToUpper(jenisKelamin) == "P" {
		return "Jazaakillahu khoiro"
	}
	return "Jazaakallahu khoiro"
}
