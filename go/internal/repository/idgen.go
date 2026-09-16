package repository

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func newIDWithPrefix(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + strings.ToUpper(hex.EncodeToString(b))
}
