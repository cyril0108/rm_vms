package middleware

import (
	"slices"
	"context"
	"net/http"
	"strings"

	"nvr_core/service"
	"nvr_core/utils"
)

// contextKey is an unexported type to prevent collisions in the context map
type contextKey string

const sessionKey = contextKey("session_data")

type PERMISSION string

const (
	PERMSystem PERMISSION ="system"

	PERMUserManage PERMISSION = "user_manage"
	PERMUserNoSelfManage PERMISSION = "user_no_self_manage"

	PERMLayout_all PERMISSION = "layout_all"

	PERMViewAllDevice PERMISSION = "view_all_device"
	PERMCameraConfigure PERMISSION = "camera_configure"
	PERMCameraPtz PERMISSION = "camera_ptz"
	PERMCameraPlayback PERMISSION = "camera_playback"
	PERMCameraLive PERMISSION = "camera_live"

	PERMRecordingExport PERMISSION = "recording_export"
)

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
				utils.RespondJSONHTTPStatus(w, "Unauthorized: missing or invalid Bearer token", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate the token using your Auth Service
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				utils.RespondJSONHTTPStatus(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
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
// func (s SessionData) HasPermission(requiredPerm string) bool {
// 	return slices.Contains(s.Permissions, requiredPerm)
// }

func (s SessionData) HasPermission(requiredPerm PERMISSION) bool {
	return slices.Contains(s.Permissions, string(requiredPerm))
}


// ---------------------------
func (s SessionData) HasPermissionUserManage() bool {
	return s.HasPermission(PERMUserManage) || s.HasPermission(PERMUserNoSelfManage)
}

func (s SessionData) HasPermissionUserNoSelfManage(targetId int64) bool {
	return s.HasPermission(PERMUserManage) || (s.HasPermission(PERMUserNoSelfManage) && s.UserID != targetId)
}
