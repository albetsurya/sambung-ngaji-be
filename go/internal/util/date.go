package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
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

// jakartaLoc adalah zona waktu acuan aplikasi (sheet + user di WIB).
// Semua kolom DATE di sheet maupun DB bermakna "date kalender WIB",
// bukan instant UTC. Gagal mengkonversi sebelum ambil YYYY-MM-DD
// menyebabkan sheet 1 Agustus (00:00 WIB = 31 Jul 17:00 UTC)
// tersimpan sebagai 31 Juli.
func jakartaLoc() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*3600)
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.In(jakartaLoc()).Format("2006-01-02")
}

// DateOnlyFromInstant mengubah instant (mis. hasil parse RFC3339 dari
// sheet/GAS: "2026-07-31T17:00:00Z") menjadi date kalender WIB
// (2026-08-01) sebagai time.Time midnight UTC, aman disimpan ke kolom DATE.
func DateOnlyFromInstant(t time.Time) time.Time {
	jt := t.In(jakartaLoc())
	return time.Date(jt.Year(), jt.Month(), jt.Day(), 0, 0, 0, 0, time.UTC)
}

// ParseRpNumber mem-parse nominal uang Indonesia: "Rp1.286.000",
// "Rp300.000", "100000", "Rp -", "-", "" → 0. Titik = ribuan,
// koma = desimal ("1.500,50" → 1500.5).
func ParseRpNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\u00a0", "")
	if s == "" || s == "-" || s == "Rp" || s == "Rp -" || s == "Rp-" {
		return 0
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "Rp"))
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, " ", "")
		s = strings.Replace(s, ",", ".", 1)
		s = strings.ReplaceAll(s, ",", "")
	} else {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		s = strings.ReplaceAll(s, " ", "")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ParseSheetMonth mem-parse token bulan susulan dari sheet menjadi "YYYY-MM".
// Menerima: "2025-08", "2026-1" (tak ber-nol), "08/2025", "8/2025",
// "Agu 2025", "Agustus 2025", "Aug 2025", "August 2025" (case-insensitive),
// bahkan date penuh ("01/08/2026" → "2026-08", pakai bulannya).
func ParseSheetMonth(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("bulan kosong")
	}
	if m := regexp.MustCompile(`^(\d{4})-(\d{1,2})$`).FindStringSubmatch(s); m != nil {
		mon := m[2]
		if len(mon) == 1 {
			mon = "0" + mon
		}
		if mon < "01" || mon > "12" {
			return "", fmt.Errorf("bulan tidak dikenal: %s", s)
		}
		return m[1] + "-" + mon, nil
	}
	if m := regexp.MustCompile(`^(\d{1,2})[/-](\d{4})$`).FindStringSubmatch(s); m != nil {
		mon := m[1]
		if len(mon) == 1 {
			mon = "0" + mon
		}
		if mon < "01" || mon > "12" {
			return "", fmt.Errorf("bulan tidak dikenal: %s", s)
		}
		return m[2] + "-" + mon, nil
	}
	lower := strings.ToLower(s)
	monthNames := map[string]string{
		"jan": "01", "januari": "01", "january": "01",
		"feb": "02", "februari": "02", "february": "02",
		"mar": "03", "maret": "03", "march": "03",
		"apr": "04", "april": "04",
		"mei": "05", "may": "05",
		"jun": "06", "juni": "06", "june": "06",
		"jul": "07", "juli": "07", "july": "07",
		"agu": "08", "agustus": "08", "aug": "08", "august": "08",
		"sep": "09", "september": "09", "sept": "09",
		"okt": "10", "oktober": "10", "oct": "10", "october": "10",
		"nov": "11", "november": "11",
		"des": "12", "desember": "12", "dec": "12", "december": "12",
	}
	parts := strings.Fields(lower)
	if len(parts) == 2 {
		if mon, ok := monthNames[parts[0]]; ok && regexp.MustCompile(`^\d{4}$`).MatchString(parts[1]) {
			return parts[1] + "-" + mon, nil
		}
	}
	// Fallback: token berupa date penuh ("01/08/2026") → pakai bulannya.
	if t, err := ParseSheetDate(s); err == nil {
		return t.Format("2006-01"), nil
	}
	return "", fmt.Errorf("bulan tidak dikenal: %s", s)
}

// SplitSheetMonths memecah teks bulan susulan sheet ("2025-06, Agu 2025; 08/2025")
// menjadi daftar "YYYY-MM" unik dan terurut. Token tak dikenal diabaikan
// (dikembalikan di unknowns agar bisa dicatat, bukan dihilangkan diam-diam).
func SplitSheetMonths(s string) (months []string, unknowns []string) {
	s = strings.ReplaceAll(s, ";", ",")
	seen := map[string]bool{}
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		m, err := ParseSheetMonth(tok)
		if err != nil {
			unknowns = append(unknowns, tok)
			continue
		}
		if !seen[m] {
			seen[m] = true
			months = append(months, m)
		}
	}
	sort.Strings(months)
	return months, unknowns
}

// ParseCarryoverBreakdown mem-parse kolom rincian susulan sheet.
// Bentuk asli: JSON map {"2025-08":83333,"2025-09":83333} (ditulis app lama).
// Nilai "{}" / kosong = tidak ada rincian (ok=false).
// Mengembalikan rincian terurut + totalnya. Kunci bulan dinormalisasi
// (terima "2026-8" dsb); amount negatif ditolak.
func ParseCarryoverBreakdown(s string) ([]DueMonthAmount, float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return nil, 0, false
	}
	var raw map[string]float64
	dec := json.NewDecoder(strings.NewReader(s))
	if err := dec.Decode(&raw); err != nil || len(raw) == 0 {
		return nil, 0, false
	}
	type kv struct {
		m string
		a float64
	}
	var list []kv
	var sum float64
	for k, v := range raw {
		m, err := ParseSheetMonth(k)
		if err != nil || v < 0 {
			return nil, 0, false
		}
		list = append(list, kv{m, v})
		sum += v
	}
	sort.Slice(list, func(i, j int) bool { return list[i].m < list[j].m })
	out := make([]DueMonthAmount, 0, len(list))
	for _, e := range list {
		out = append(out, DueMonthAmount{Month: e.m, Amount: e.a})
	}
	return out, sum, true
}

// DueMonthAmount = satu baris rincian susulan (bulan + nominal).
type DueMonthAmount struct {
	Month  string
	Amount float64
}

// ParseSheetDate mem-parse date dari sheet/GAS ke date kalender WIB.
// Menerima: "2006-01-02", "02/01/2006", "2/1/2006", "02-01-2006",
// "2 Jan 2006", RFC3339/RFC3339Nano (instant → dikonversi ke WIB dulu),
// dan serial number Excel/Sheets (day sejak 1899-12-30).
func ParseSheetDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("date kosong")
	}
	// Serial number Google Sheets/Excel, mis. "45885" (= 1 Agu 2026).
	if f, err := strconv.ParseFloat(s, 64); err == nil && f > 20000 && f < 80000 && !strings.ContainsAny(s, "-/:T") {
		days := int(f)
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		t := base.AddDate(0, 0, days)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	// Instant ber-timezone (dari JSON.stringify Date sheet) → konversi ke WIB.
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return DateOnlyFromInstant(t), nil
		}
	}
	// Date kalender tanpa time → langsung midnight UTC.
	for _, f := range []string{
		"2006-01-02",
		"2006/01/02",
		"02-01-2006",
		"02/01/2006",
		"2/1/2006",
		"2-1-2006",
		"02-Jan-2006",
		"02 Jan 2006",
		"2 January 2006",
		"02 January 2006",
		"January 2, 2006",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(f, s); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("format tanggal tidak dikenal: %s", s)
}

func ParseFlexibleDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("date kosong")
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
