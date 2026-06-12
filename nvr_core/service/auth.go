package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/security"
)

// const TokenExpireTime = 24 * time.Hour
// const RefreshTokenExpireTime = 7 * 24 * time.Hour
const TokenExpireTime = 1024 * time.Hour
const RefreshTokenExpireTime = 2 * 1024 * time.Hour

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
	RefreshToken(ctx context.Context, rawRefreshToken string) (string, []string, error)
	RevokeRefreshToken(ctx context.Context, rawRefreshToken string) (error)
	UpdateUserStatusForPermissionChange(userID int64)
	LogoutDeactivatedUser(ctx context.Context, userID int64) error
}


// Notice we align this with the struct in your services.base.go
func NewAuthService(userRepo repository.UserRepository, permRepo repository.PermissionRepository, tknRepo repository.RefreshTokenRepository, secretKey string) AuthService {
	return &authServiceBase{
		userRepo:   userRepo,
		reTokenRepo: tknRepo,
		permRepo:   permRepo,
		jwtSecret:  []byte(secretKey),
		tokenExpir: TokenExpireTime, // Standard session length for NVRs
		userStatus:   NewUserStatusMap(),
	}
}

func (s *authServiceBase) LogoutDeactivatedUser(ctx context.Context, userID int64) error {

	if err := s.reTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}

	// ==========================================
	// UPDATE IN-MEMORY STATUS
	// ==========================================
	// Register the user in the allowlist so their new JWT is instantly valid
	s.userStatus.LogoutAll(userID)

	return  nil
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
	accessToken, err := s.generateToken(user, permissions)
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

	// Hash the token before storing it in SQLite
	hashedToken:= security.HashRefreshToken(rawRefreshToken)

	// Build the database model
	refreshTokenRecord := &models.RefreshToken{
		ID:        uuid.New().String(), // Generate SQLite UUID manually
		UserID:    user.ID,
		TokenHash: hashedToken,
		ExpiresAt: time.Now().Add(RefreshTokenExpireTime), // 7-day session lifespan
	}

	// Save to database
	if err := s.reTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
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

/**
 * ------------------------------------
 * @param  {[type]} s *authServiceBase) RefreshToken(ctx context.Context, rawRefreshToken string) (string, []string, error [description]
 * @return {[type]}   [description]
 * ------------------------------------
 */
func (s *authServiceBase) RefreshToken(ctx context.Context, rawRefreshToken string) (string, []string, error) {
	// Hash the incoming raw token using our fast deterministic hash
	hash := security.HashRefreshToken(rawRefreshToken)

	// Fetch the token record from SQLite
	tokenRecord, err := s.reTokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return "", nil, ErrUnauthorized // Token not found or invalid
	}

	// Validate token state (ensure it isn't expired or manually revoked)
	if tokenRecord.IsRevoked || time.Now().After(tokenRecord.ExpiresAt) {
		return "", nil, ErrUnauthorized
	}

	// Fetch the user to ensure their account hasn't been disabled since they logged in
	user, err := s.userRepo.GetByID(ctx, int64(tokenRecord.UserID))
	if err != nil || !user.IsActive {
		return "", nil, ErrAccountDisabled
	}

	// Fetch the user's *newest* permissions directly from the database
	permissions, err := s.permRepo.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}

	// Generate the brand new JWT Access Token
	accessToken, err := s.generateToken(user, permissions)

	if err != nil {
		return "", nil, err
	}

	// Update the In-Memory Allowlist (Self-Healing)
	// This quietly re-registers the user if the server rebooted, or updates their 
	// timestamp if their permissions were just changed.
	s.userStatus.Login(user.ID)

	return accessToken, permissions, nil
}

func (s *authServiceBase) RevokeRefreshToken(ctx context.Context, rawRefreshToken string) (error) {
	// Hash the incoming raw token using our fast deterministic hash
	hash := security.HashRefreshToken(rawRefreshToken)
	return s.reTokenRepo.RevokeHashed(ctx, hash)
}


func (s *authServiceBase) UpdateUserStatusForPermissionChange(userID int64) {
	s.userStatus.UpdatePermissions(userID)
}


// =================================================================
// =================================================================
/**
 * Private function that generate our access token
 */
// =================================================================
// =================================================================
func (s *authServiceBase) generateToken(user *models.User, permissions []string) (string, error) {

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
	return token.SignedString(s.jwtSecret)
}