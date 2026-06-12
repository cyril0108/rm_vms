package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSecureToken creates a random hex string for the refresh token
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}