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
	// Login validates credentials and returns a signed JWT and the permission list
	Login(ctx context.Context, username string, password string) (string, []string, error)
	// ValidateToken parses a JWT and returns its claims if valid
	ValidateToken(tokenString string) (jwt.MapClaims, error)
	RevokeToken(tokenString string) error
}

// Notice we align this with the struct in your services.base.go
func NewAuthService(userRepo repository.UserRepository, permRepo repository.PermissionRepository, secretKey string) AuthService {
	return &authServiceBase{
		userRepo:   userRepo,
		permRepo:   permRepo,
		jwtSecret:  []byte(secretKey),
		tokenExpir: 24 * time.Hour, // Standard session length for NVRs
		denylist:   NewInMemoryDenylist(1 * time.Hour),
	}
}

func (s *authServiceBase) Login(ctx context.Context, username string, password string) (string, []string, error) {
	// Fetch user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", nil, ErrInvalidCredentials // Generic error to prevent username enumeration
		}
		return "", nil, err
	}

	// Check active status
	if !user.IsActive {
		return "", nil, ErrAccountDisabled
	}

	// Verify bcrypt password hash
	match, err := security.CheckPasswordHash(password, user.Password)
	if err != nil || !match {
		return "", nil, ErrInvalidCredentials
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
		return "", nil, err
	}

	tokenID := uuid.New().String()

	// Construct the JWT Claims
	claims := jwt.MapClaims{
		"jti":   tokenID,
		"sub":   user.ID,
		"name":  user.Username,
		"role":  user.RoleID,
		"perms": permissions, // Embed permissions directly in the token
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.tokenExpir).Unix(),
	}

	// Sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return signedToken, permissions, nil
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

	// Check if the token's JTI is on the denylist
	if jti, ok := claims["jti"].(string); ok {
		if s.denylist.IsRevoked(jti) {
			return nil, ErrUnauthorized // Instantly reject disabled users
		}
	}

	return claims, nil
}

func (s *authServiceBase) RevokeToken(tokenString string) error {
	// We parse and validate it first to ensure malicious users can't fill our memory cache with garbage data
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// If it's already invalid or naturally expired, we don't need to do anything
		return nil 
	}

	jti, jtiOk := claims["jti"].(string)
	expFloat, expOk := claims["exp"].(float64) // jwt-go parses JSON numbers as float64

	if jtiOk && expOk {
		expTime := time.Unix(int64(expFloat), 0)
		s.denylist.Revoke(jti, expTime)
	}

	return nil
}