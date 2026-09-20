package util

import (
	"strings"
	"unicode"
)

// TitleCaseID mengubah string nama jadi Title Case sesuai konvensi Indonesia.
// Menangani: spasi ganda, ALL CAPS, huruf kecil semua, nama hyphenated, apostrof.
// Input kosong / whitespace-only mengembalikan string kosong.
//
//	"galih rahmat hidayat"  -> "Galih Rahmat Hidayat"
//	"ALBET SURYA KEMBARA"   -> "Albet Surya Kembara"
//	"muhammad o'brien"      -> "Muhammad O'Brien"
//	"abdul-rahman"          -> "Abdul-Rahman"
func TitleCaseID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Collapse multiple whitespace jadi single space
	s = strings.Join(strings.Fields(s), " ")

	var b strings.Builder
	b.Grow(len(s))

	capNext := true
	for _, r := range s {
		switch {
		case r == ' ' || r == '-' || r == '\'':
			b.WriteRune(r)
			capNext = true
		case capNext:
			b.WriteRune(unicode.ToUpper(r))
			capNext = false
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}
