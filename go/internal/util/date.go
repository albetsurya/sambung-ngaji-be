package util

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

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

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func ParseFlexibleDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("tanggal kosong")
	}

	formats := []string{
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"2 January 2006",
		"02 January 2006",
		"January 2, 2006",
		"2006/01/02",
		"02-Jan-2006",
		"02 Jan 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("format tanggal tidak dikenal: %s", s)
}
