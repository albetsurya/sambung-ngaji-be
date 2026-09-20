package util

import "testing"

func TestTitleCaseID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"   ", ""},
		{"sasa", "Sasa"},
		{"galih rahmat hidayat", "Galih Rahmat Hidayat"},
		{"ALBET SURYA KEMBARA", "Albet Surya Kembara"},
		{"Alecia Keisyafani trihapsari", "Alecia Keisyafani Trihapsari"},
		{"Aisya Permata Ainy", "Aisya Permata Ainy"},
		{"  medokan   semampir  ", "Medokan Semampir"},
		{"muhammad o'brien", "Muhammad O'Brien"},
		{"abdul-rahman", "Abdul-Rahman"},
		{"a", "A"},
		{"JALAN-SUDIRMAN", "Jalan-Sudirman"},
		{"Ummi Kalsum binti Abdullah", "Ummi Kalsum Binti Abdullah"},
	}
	for _, c := range cases {
		got := TitleCaseID(c.in)
		if got != c.want {
			t.Errorf("TitleCaseID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
