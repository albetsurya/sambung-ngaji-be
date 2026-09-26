package util

import (
	"strings"
	"unicode"
)

func TitleCaseID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
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
