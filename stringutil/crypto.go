package stringutil

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken GenerateRandomString Generates a random token with a given length
func GenerateToken(size int) (string, error) {
	randomBytes := make([]byte, size)

	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
