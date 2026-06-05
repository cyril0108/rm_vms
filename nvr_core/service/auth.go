package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"nvr_core/db/repository"
	"nvr_core/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrUnauthorized       = errors.New("unauthorized token")
)

type AuthService interface {
	// Login validates credentials and returns (accessToken, refreshToken, permissions, error)
	Login(ctx context.Context, username string, password string) (string, string, []string, error)
	// ValidateToken parses a JWT and returns its claims if valid
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// Notice we align this with the struct in your services.base.go
func NewAuthService(userRepo repository.UserRepository, permRepo repository.PermissionRepository, tknRepo repository.RefreshTokenRepository, secretKey string) AuthService {
	return &authServiceBase{
		userRepo:   userRepo,
		reTokenRepo: tknRepo,
		permRepo:   permRepo,
		jwtSecret:  []byte(secretKey),
		tokenExpir: 1024 * time.Hour, // Standard session length for NVRs
		userStatus:   NewUserStatusMap(),
	}
}

func (s *authServiceBase) Login(ctx context.Context, username string, password string) (string, string, []string, error) {
	// Fetch user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", "", nil, ErrInvalidCredentials // Generic error to prevent username enumeration
		}
		return "", "", nil, err
	}

	// Check active status
	if !user.IsActive {
		return "", "", nil, ErrAccountDisabled
	}

	// Verify bcrypt password hash
	match, err := security.CheckPasswordHash(password, user.Password)
	if err != nil || !match {
		return "", "", nil, ErrInvalidCredentials
	}

	// Try to migrate to new hash algorithm automatically
	if security.IsLegacyHash(user.Password) {

		newHash, hashErr := security.HashPassword(password)
		if hashErr == nil {
			_ = s.userRepo.UpdatePassword(ctx, user.ID, newHash)
		}

	}

	// Fetch the aggregated permissions (Role + Direct Grants)
	permissions, err := s.permRepo.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		return "", "", nil, err
	}

	// ==========================================
	// GENERATE ACCESS TOKEN (JWT)
	// ==========================================
	tokenID := uuid.New().String()
	claims := jwt.MapClaims{
		"jti":   tokenID,
		"sub":   user.ID,
		"name":  user.Username,
		"role":  user.RoleID,
		"perms": permissions,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.tokenExpir).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", nil, err
	}

	// ==========================================
	// GENERATE & SAVE REFRESH TOKEN
	// ==========================================
	rawRefreshToken, err := security.GenerateSecureToken(32) // Generates a 64-character hex string
	if err != nil {
		return "", "", nil, err
	}

	// ==========================================
	// UPDATE IN-MEMORY STATUS
	// ==========================================
	// Register the user in the allowlist so their new JWT is instantly valid
	s.userStatus.Login(user.ID)

	return accessToken, rawRefreshToken, permissions, nil
}

func (s *authServiceBase) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Extract the sub (UserID) and iat (Issued At) from the claims
	userIDFloat, subOk := claims["sub"].(float64)
	iatFloat, iatOk := claims["iat"].(float64)

	if !subOk || !iatOk {
	    return nil, ErrUnauthorized
	}

	userID := int64(userIDFloat)
	tokenIssuedAt := int64(iatFloat)

	// Check the in-memory allowlist
	if !s.userStatus.IsValid(userID, tokenIssuedAt) {
	    return nil, ErrUnauthorized // Instantly reject stale or logged-out users
	}

	return claims, nil
}