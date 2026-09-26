package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func VerifyPassword(plain, hash string) (bool, bool) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return false, false
	}

	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
		return err == nil, false
	}

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
