package util

import "time"

const (
	KatBalita    = "BALITA"
	KatCaberawit = "CABERAWIT"
	KatPraRemaja = "PRA_REMAJA"
	KatRemaja    = "REMAJA"
	KatPraNikah  = "PRA_NIKAH"
	KatDewasa    = "DEWASA"
	KatIstimewa  = "ISTIMEWA"
)

// GetMemberCategory — port persis dari utils.js.
// Input: tanggalLahir, jenjangPendidikan (SD/SMP/SMA/SMK/PAUD/TK), isNikah.
func GetMemberCategory(tanggalLahir *time.Time, jenjangPendidikan string, isNikah bool) string {
	age := GetAge(tanggalLahir)

	if isNikah {
		if age >= 0 && age >= 60 {
			return KatIstimewa
		}
		return KatDewasa
	}

	jenjang := ""
	for _, c := range jenjangPendidikan {
		if c >= 'a' && c <= 'z' {
			jenjang += string(c - 32)
		} else {
			jenjang += string(c)
		}
	}

	switch jenjang {
	case "PAUD", "TK", "SD":
		return KatCaberawit
	case "SMP":
		return KatPraRemaja
	case "SMA", "SMK":
		return KatRemaja
	}

	if age >= 0 && age >= 60 {
		return KatIstimewa
	}
	if age >= 0 && age < 6 {
		return KatBalita
	}
	if age >= 0 && age < 13 {
		return KatCaberawit
	}
	if age >= 0 && age < 16 {
		return KatPraRemaja
	}
	if age >= 0 && age < 19 {
		return KatRemaja
	}
	return KatPraNikah
}
