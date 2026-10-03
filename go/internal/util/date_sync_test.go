package util

import (
	"testing"
	"time"
)

// Regresi: sheet 1 Agu 00:00 WIB diserial JSON sebagai
// "2026-07-31T17:00:00.000Z" - harus kembali menjadi 2026-08-01, bukan 31 Juli.
func TestParseSheetDate_RFC3339InstantKeWIB(t *testing.T) {
	got, err := ParseSheetDate("2026-07-31T17:00:00.000Z")
	if err != nil {
		t.Fatalf("parse gagal: %v", err)
	}
	if got.Format("2006-01-02") != "2026-08-01" {
		t.Fatalf("selisih 1 day: dapat %s, mau 2026-08-01", got.Format("2006-01-02"))
	}
}

func TestParseSheetDate_KalenderBiasa(t *testing.T) {
	for in, want := range map[string]string{
		"2026-08-01": "2026-08-01",
		"01/08/2026": "2026-08-01",
		"1/8/2026":   "2026-08-01",
		"01-08-2026": "2026-08-01",
	} {
		got, err := ParseSheetDate(in)
		if err != nil {
			t.Fatalf("parse %q gagal: %v", in, err)
		}
		if got.Format("2006-01-02") != want {
			t.Fatalf("parse %q: dapat %s, mau %s", in, got.Format("2006-01-02"), want)
		}
	}
}

func TestFormatDate_DariInstantUTC(t *testing.T) {
	// 1 Agu 00:00 WIB sebagai instant UTC.
	inst := time.Date(2026, 7, 31, 17, 0, 0, 0, time.UTC)
	if got := FormatDate(&inst); got != "2026-08-01" {
		t.Fatalf("FormatDate: dapat %s, mau 2026-08-01", got)
	}
	// DATE midnight UTC tetap day yang sama.
	mid := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if got := FormatDate(&mid); got != "2026-08-01" {
		t.Fatalf("FormatDate midnight: dapat %s, mau 2026-08-01", got)
	}
}

func TestDateOnlyFromInstant(t *testing.T) {
	inst := time.Date(2026, 7, 31, 17, 0, 0, 0, time.UTC)
	got := DateOnlyFromInstant(inst)
	if got.Format("2006-01-02") != "2026-08-01" {
		t.Fatalf("dapat %s, mau 2026-08-01", got.Format("2006-01-02"))
	}
}

func TestParseSheetMonth(t *testing.T) {
	cases := map[string]string{
		"2025-08":      "2025-08",
		"08/2025":      "2025-08",
		"8/2025":       "2025-08",
		"Agu 2025":     "2025-08",
		"Agustus 25":   "",
		"Agustus 2025": "2025-08",
		"January 2026": "2026-01",
	}
	for in, want := range cases {
		got, err := ParseSheetMonth(in)
		if want == "" {
			if err == nil {
				t.Fatalf("parse %q seharusnya gagal", in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parse %q gagal: %v", in, err)
		}
		if got != want {
			t.Fatalf("parse %q: dapat %s, mau %s", in, got, want)
		}
	}
}

func TestSplitSheetMonths(t *testing.T) {
	months, unknowns := SplitSheetMonths("2025-06, Agu 2025; 08/2025, ngawur")
	if len(unknowns) != 1 || unknowns[0] != "ngawur" {
		t.Fatalf("unknowns salah: %v", unknowns)
	}
	if len(months) != 2 || months[0] != "2025-06" || months[1] != "2025-08" {
		t.Fatalf("months salah: %v", months)
	}
}

func TestParseRpNumber(t *testing.T) {
	cases := map[string]float64{
		"Rp300.000":   300000,
		"Rp1.286.000": 1286000,
		"100000":      100000,
		"Rp -":        0,
		"-":           0,
		"":            0,
		"0":           0,
		"Rp7.000.000": 7000000,
	}
	for in, want := range cases {
		if got := ParseRpNumber(in); got != want {
			t.Fatalf("ParseRpNumber(%q): dapat %v, mau %v", in, got, want)
		}
	}
}

func TestParseCarryoverBreakdown(t *testing.T) {
	items, sum, ok := ParseCarryoverBreakdown(`{"2025-08":83333,"2025-09":83333,"2025-10":83334}`)
	if !ok || len(items) != 3 || sum != 250000 {
		t.Fatalf("breakdown salah: %v %v %v", items, sum, ok)
	}
	if items[0].Month != "2025-08" || items[2].Amount != 83334 {
		t.Fatalf("isi breakdown salah: %+v", items)
	}
	for _, s := range []string{"", "{}", "ngawur", `{"xx":1}`} {
		if _, _, ok := ParseCarryoverBreakdown(s); ok {
			t.Fatalf("breakdown %q seharusnya gagal", s)
		}
	}
}

func TestParseSheetMonthPadding(t *testing.T) {
	for in, want := range map[string]string{"2026-1": "2026-01", "2026-2": "2026-02", "01/08/2026": "2026-08"} {
		got, err := ParseSheetMonth(in)
		if err != nil || got != want {
			t.Fatalf("ParseSheetMonth(%q): dapat %q err=%v, mau %q", in, got, err, want)
		}
	}
}
