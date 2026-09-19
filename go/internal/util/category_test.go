package util

import (
	"testing"
	"time"
)

func atAge(age int) *time.Time {
	now := time.Now()
	t := time.Date(now.Year()-age, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return &t
}

func TestGetMemberCategory(t *testing.T) {
	cases := []struct {
		name    string
		age     int
		hasTTL  bool
		jenjang string
		nikah   bool
		want    string
	}{
		// ---- Nikah ----
		{"nikah muda -> Dewasa", 25, true, "SMA", true, KatDewasa},
		{"nikah lansia -> Istimewa", 70, true, "", true, KatIstimewa},
		{"nikah tanpa TTL -> Dewasa", -1, false, "", true, KatDewasa},

		// ---- Jenjang-primary (fix umur 4 PAUD) ----
		{"PAUD umur 4 -> Caberawit", 4, true, "PAUD", false, KatCaberawit},
		{"TK umur 5 -> Caberawit", 5, true, "TK", false, KatCaberawit},
		{"SD umur 10 -> Caberawit", 10, true, "SD", false, KatCaberawit},
		{"SMP umur 14 -> PraRemaja", 14, true, "SMP", false, KatPraRemaja},
		{"SMA umur 17 -> Remaja", 17, true, "SMA", false, KatRemaja},
		{"SMK umur 17 -> Remaja", 17, true, "smk", false, KatRemaja},

		// ---- Jenjang = pendidikan terakhir (umur melewati jendela) ----
		{"SMA terakhir umur 30 -> PraNikah", 30, true, "SMA", false, KatPraNikah},
		{"SMP terakhir umur 25 -> PraNikah", 25, true, "SMP", false, KatPraNikah},
		{"SD terakhir umur 20 -> PraNikah", 20, true, "SD", false, KatPraNikah},

		// ---- Fallback umur (tanpa jenjang) ----
		{"balita tanpa jenjang", 4, true, "", false, KatBalita},
		{"caberawit tanpa jenjang", 10, true, "", false, KatCaberawit},
		{"pra remaja tanpa jenjang", 14, true, "", false, KatPraRemaja},
		{"remaja tanpa jenjang", 17, true, "", false, KatRemaja},
		{"lansia tanpa jenjang -> Istimewa", 65, true, "", false, KatIstimewa},
		{"dewa dewasa tanpa jenjang -> PraNikah", 25, true, "", false, KatPraNikah},

		// ---- Tanpa TTL ----
		{"tanpa TTL tanpa jenjang -> PraNikah", -1, false, "", false, KatPraNikah},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var tgl *time.Time
			if c.hasTTL {
				tgl = atAge(c.age)
			}
			got := GetMemberCategory(tgl, c.jenjang, c.nikah)
			if got != c.want {
				t.Fatalf("GetMemberCategory(age=%d, hasTTL=%v, jenjang=%q, nikah=%v) = %q, want %q",
					c.age, c.hasTTL, c.jenjang, c.nikah, got, c.want)
			}
		})
	}
}
