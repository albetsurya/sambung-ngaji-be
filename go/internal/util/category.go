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

// Ambang jendela umur: jenjang valid selama umur masih masuk rentang wajar.
// Kalau umur melewati ambang, jenjang dianggap "pendidikan terakhir" →
// jatuh ke klasifikasi umur.
const (
	caberawitMaxAge = 13 // PAUD - Kelas 6 SD, tipikal 4-12
	praRemajaMaxAge = 16 // SMP, tipikal 12-15
	remajaMaxAge    = 19 // SMA/SMK, tipikal 15-18
	istimewaMinAge  = 60
	balitaMaxAge    = 6 // 0-5, belum sekolah
)

// GetMemberCategory — klasifikasi kohort member.
// Input: tanggalLahir, jenjangPendidikan (PAUD/TK/SD/SMP/SMA/SMK), isNikah.
// Strategi: jenjang jadi penentu utama; umur jadi fallback + pengaman.
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

	// Tanpa jenjang (atau jenjang = "pendidikan terakhir" di luar jendela) → pakai umur.
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
	// Termasuk umur -1 (tanpa TTL) → default Pra Nikah.
	return KatPraNikah
}

// upper mengubah string ke uppercase (pendukung jenjang, ASCII-safe).
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
