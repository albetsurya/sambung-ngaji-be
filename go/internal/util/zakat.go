package util

import "strings"

// Kanonis kategori zakat: FITRAH, MAL, TIJAROH, ZURU, LIVESTOCK, OTHER.
// Dipakai semua write-path (API, sync spreadsheet, import) agar ejaan
// seragam dengan CHECK constraint di database.
func NormZakatCategory(v string) string {
	s := strings.ToUpper(strings.TrimSpace(v))
	s = strings.TrimPrefix(s, "ZAKAT ")
	s = strings.Trim(s, "' ")
	switch s {
	case "FITRAH", "FITR":
		return "FITRAH"
	case "MAL", "MAAL":
		return "MAL"
	case "TIJAROH", "TIJARAH", "DAGANG":
		return "TIJAROH"
	case "ZURU", "ZIRA'AH", "ZIRAAH", "PERTANIAN":
		return "ZURU"
	case "LIVESTOCK", "TERNAK":
		return "LIVESTOCK"
	case "OTHER", "LAINNYA", "LAIN-LAIN":
		return "OTHER"
	default:
		return "FITRAH"
	}
}
