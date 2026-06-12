package middleware

import (
	"slices"
	"context"
	"net/http"
	"strings"

	"nvr_core/service"
)

// contextKey is an unexported type to prevent collisions in the context map
type contextKey string

const sessionKey = contextKey("session_data")

// SessionData holds the extracted JWT claims for easy access in handlers
type SessionData struct {
	UserID      int64
	Username    string
	RoleID      int64
	Permissions []string
}

// RequireAuth is a middleware that enforces JWT authentication
func RequireAuth(authService service.AuthService) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: missing or invalid Bearer token", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate the token using your Auth Service
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extract data from claims (safely handling type assertions)
			// Note: jwt-go parses JSON numbers as float64
			userID, _ := claims["sub"].(float64)
			roleID, _ := claims["role"].(float64)
			username, _ := claims["name"].(string)
			
			// Safely extract the string array of permissions
			var permissions []string
			if rawPerms, ok := claims["perms"].([]interface{}); ok {
				for _, p := range rawPerms {
					if strPerm, ok := p.(string); ok {
						permissions = append(permissions, strPerm)
					}
				}
			}

			// Build the SessionData struct
			session := SessionData{
				UserID:      int64(userID),
				Username:    username,
				RoleID:      int64(roleID),
				Permissions: permissions,
			}

			// Inject the session into the request context
			ctx := context.WithValue(r.Context(), sessionKey, session)

			// Pass the request to the next handler with the new context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSession extracts the SessionData from the request context
func GetSession(ctx context.Context) (SessionData, bool) {
	session, ok := ctx.Value(sessionKey).(SessionData)
	return session, ok
}

// HasPermission is a quick utility to check if a user has a specific permission
func (s SessionData) HasPermission(requiredPerm string) bool {
	return slices.Contains(s.Permissions, requiredPerm)
}