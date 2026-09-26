package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func dataDir2(dir, name string) string {
	return filepath.Join(dir, name+".csv")
}

func readCSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}

	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		if i == 0 {
			h = strings.TrimPrefix(h, "\ufeff")
		}
		headers[i] = h
	}

	var out []map[string]string
	for _, row := range records[1:] {
		empty := true
		for _, v := range row {
			if strings.TrimSpace(v) != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		m := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func str(s string) string { return strings.TrimSpace(s) }

func strDef(s, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "true" || s == "1" || s == "ya" || s == "yes" || s == "t"
}

func parseGender(s string) *string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil
	}
	switch s {
	case "l", "laki-laki", "laki laki", "pria", "male":
		r := "L"
		return &r
	case "p", "perempuan", "wanita", "female":
		r := "P"
		return &r
	}
	return nil
}

func parseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		wib := t.In(time.FixedZone("WIB", 7*3600))
		d := time.Date(wib.Year(), wib.Month(), wib.Day(), 0, 0, 0, 0, time.UTC)
		return &d
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		return &d
	}
	return nil
}

func parseTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

func tPtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func sPtr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
