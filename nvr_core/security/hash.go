package security

import "golang.org/x/crypto/argon2"
import "golang.org/x/crypto/bcrypt"

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible variant of argon2")
)

// Standard OWASP recommended parameters for Argon2id
type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var currentParams = &argonParams{
	memory:      64 * 1024, // 64 MB
	iterations:  3,         
	parallelism: 2,         
	saltLength:  16,        
	keyLength:   32,        
}

// HashPassword now generates an Argon2id hash in the standard PHC string format.
// Example: $argon2id$v=19$m=65536,t=3,p=2$saltbase64$keybase64
func HashPassword(password string) (string, error) {
	salt := make([]byte, currentParams.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, currentParams.iterations, currentParams.memory, currentParams.parallelism, currentParams.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, currentParams.memory, currentParams.iterations, currentParams.parallelism, b64Salt, b64Hash,
	)

	return encodedHash, nil
}

// CheckPasswordHash is now a smart router. It checks the prefix to determine 
// if it should verify using legacy bcrypt or modern Argon2id.
func CheckPasswordHash(password, encodedHash string) (bool, error) {
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		return compareArgon2id(password, encodedHash)
	}

	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		return err == nil, nil
	}

	return false, ErrInvalidHash
}

// IsLegacyHash helps the login service know if the user needs an automatic upgrade.
func IsLegacyHash(encodedHash string) bool {
	return !strings.HasPrefix(encodedHash, "$argon2id$")
}

// compareArgon2id parses the PHC string and securely compares the bytes.
func compareArgon2id(password, encodedHash string) (bool, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	p := &argonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	p.saltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, err
	}
	p.keyLength = uint32(len(hash))

	comparisonHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Use constant-time comparison to prevent timing side-channel attacks
	if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
		return true, nil
	}
	return false, nil
}
