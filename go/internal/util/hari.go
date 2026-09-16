package util

import "time"

var hariID = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

func GetHariFromDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return hariID[int(t.Weekday())]
}
