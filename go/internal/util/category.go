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

const (
	caberawitMaxAge = 13
	praRemajaMaxAge = 16
	remajaMaxAge    = 19
	istimewaMinAge  = 60
	balitaMaxAge    = 6
)

func GetMemberCategory(tanggalLahir *time.Time, jenjangPendidikan string, isNikah bool) string {
	age := GetAge(tanggalLahir)

	if isNikah {
		if age >= istimewaMinAge {
			return KatIstimewa
		}
		return KatDewasa
	}

	jenjang := upper(jenjangPendidikan)

	switch jenjang {
	case "PAUD", "TK", "SD":
		if age >= 0 && age < caberawitMaxAge {
			return KatCaberawit
		}
	case "SMP":
		if age >= 0 && age <= praRemajaMaxAge {
			return KatPraRemaja
		}
	case "SMA", "SMK":
		if age >= 0 && age <= remajaMaxAge {
			return KatRemaja
		}
	}

	if age >= istimewaMinAge {
		return KatIstimewa
	}
	if age >= 0 && age < balitaMaxAge {
		return KatBalita
	}
	if age >= 0 && age < caberawitMaxAge {
		return KatCaberawit
	}
	if age >= 0 && age <= praRemajaMaxAge {
		return KatPraRemaja
	}
	if age >= 0 && age <= remajaMaxAge {
		return KatRemaja
	}
	return KatPraNikah
}

func upper(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out[i] = c
	}
	return string(out)
}
