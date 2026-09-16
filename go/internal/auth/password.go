package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// VerifyPassword mendukung 2 format hash:
//   - SHA-256 hex 64 karakter (legacy dari Apps Script)
//   - bcrypt ($2a$ / $2b$ / $2y$)
//
// Return: (match, needsRehash).
// needsRehash=true → hash lama SHA-256, harus di-upgrade ke bcrypt.
func VerifyPassword(plain, hash string) (bool, bool) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return false, false
	}

	// bcrypt
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
		return err == nil, false
	}

	// SHA-256 legacy
	if len(hash) == 64 {
		got := sha256.Sum256([]byte(plain))
		if hex.EncodeToString(got[:]) == strings.ToLower(hash) {
			return true, true
		}
	}

	return false, false
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
