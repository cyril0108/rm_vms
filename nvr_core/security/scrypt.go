package security

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

/**
 * scrypt support is only for migrating from scrypt
 * to latter algorithm.
 * Check scrypt format:
 * scrypt$N=16384,r=8,p=1$salt_base64$hash_base64
 */

// compareScrypt parses the legacy string and securely compares the bytes.
func compareScrypt(password, encodedHash string) (bool, error) {
	// Expected format: $scrypt$N=16384,r=8,p=1$salt_base64$hash_base64
	vals := strings.Split(encodedHash, "$")

	// vals[0] is empty (before the first $)
	// vals[1] is "scrypt"
	// vals[2] is the parameters "N=16384,r=8,p=1"
	// vals[3] is the salt
	// vals[4] is the hash
	if len(vals) != 5 {
		return false, ErrInvalidHash
	}

	var N, r, p int
	// Parse the N, r, and p parameters from the string
	_, err := fmt.Sscanf(vals[2], "N=%d,r=%d,p=%d", &N, &r, &p)
	if err != nil {
		return false, fmt.Errorf("failed to parse scrypt parameters: %w", err)
	}

	// Decode the salt. 
	// (Change to base64.StdEncoding if your legacy system used '=' padding)
	salt, err := base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return false, err
	}

	// Decode the expected hash
	expectedHash, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	// Scrypt allows variable key lengths. We determine the key length 
	// by checking how long the decoded expected hash is.
	keyLen := len(expectedHash)

	// Hash the plaintext password using the exact parameters extracted from the legacy string
	actualHash, err := scrypt.Key([]byte(password), salt, N, r, p, keyLen)
	if err != nil {
		return false, err
	}

	// Use constant-time comparison to prevent timing side-channel attacks
	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}
	
	return false, nil
}