package security

import (
	"crypto/rsa"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"nvr_core/logger"
)

var LOG = logger.NewLogger("[license]")

//go:embed key/public.key.pem
var thePublicKey embed.FS


var (
	ErrInvalidLicense      = errors.New("invalid license")
	ErrInappropriateSigned = errors.New("inappropriate signed")
	ErrMissingClaims       = errors.New("missing required claims")
)


func LoadPublicKey() (*rsa.PublicKey, error) {
	// Read the file directly from the embedded filesystem
	keyBytes, err := thePublicKey.ReadFile("key/public.key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded public key: %w", err)
	}

	// Parse the raw PEM bytes into a usable RSA public key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return publicKey, nil
}

func GetLicenseInfo(tokenString string) (jwt.MapClaims, error) {
	publicKey, err := LoadPublicKey()
	if err != nil {
		// If the key can't load, the system cannot validate licenses safely
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is part of the RSA family (handles RS256, RS384, RS512)
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInappropriateSigned
		}
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		LOG.Info("token invalid", "token", token, "err", err)
		return nil, ErrInvalidLicense
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		LOG.Info("Claims not okay", "token", token)
		return nil, ErrInvalidLicense
	}

	// Extract the aud (Audience) and iat (Issued At) from the claims
	// JWT standard: 'aud' is usually a string, 'iat' is a numeric date (float64 in Go JSON)
	_, audOk := claims["aud"].(string)
	_, iatOk := claims["iat"].(float64)

	if !audOk || !iatOk {
		return nil, ErrMissingClaims
	}

	return claims, nil
}