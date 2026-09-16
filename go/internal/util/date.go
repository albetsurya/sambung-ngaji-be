package util

import "time"

// GetAge: hitung usia dari tanggal lahir.
// Return -1 kalau tanggal invalid/nil.
func GetAge(tglLahir *time.Time) int {
	if tglLahir == nil {
		return -1
	}
	now := time.Now()
	age := now.Year() - tglLahir.Year()
	m := int(now.Month()) - int(tglLahir.Month())
	if m < 0 || (m == 0 && now.Day() < tglLahir.Day()) {
		age--
	}
	return age
}

// FormatDate: return "YYYY-MM-DD" atau "" kalau nil.
func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
