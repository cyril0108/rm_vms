package apiserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"nvr_core/apiserver/dto"
	"nvr_core/service"
)

const WebCookieNameRefreshToken = "nvr_refresh_token"
const WebCookiePathForRefreshToken = "/api/web/refresh"

func getWebRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(WebCookieNameRefreshToken)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func setWebRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {

	http.SetCookie(w, &http.Cookie{
		Name:     WebCookieNameRefreshToken,
		Value:    refreshToken,
		Path:     WebCookiePathForRefreshToken, // ONLY send this cookie to the refresh endpoint!
		Expires:  time.Now().Add(service.RefreshTokenExpireTime),
		HttpOnly: true,  // JavaScript CANNOT read this (XSS protection)
		Secure:   true,  // Requires HTTPS
		SameSite: http.SameSiteStrictMode, // CSRF protection
	})

}

func revokeWebRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     WebCookieNameRefreshToken,
		Value:    "",
		Path:     WebCookiePathForRefreshToken,
		MaxAge:   -1, // Instantly expires the cookie
		HttpOnly: true,
		Secure:   true,
	})
}

// HandleWebLogin expects: POST /api/web/login
//
//	@Summary		User login
//	@Description	Authenticates a user by username and password. Returns a short-lived JWT access token, a long-lived refresh token, and a list of user permissions.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		dto.LoginRequest	true	"User Login Credentials"
//	@Success		200			{object}	dto.LoginResponse	"Successfully authenticated"
//	@Failure		400			{string}	string				"Invalid JSON payload"
//	@Failure		401			{string}	string				"Invalid credentials or account disabled"
//	@Failure		500			{string}	string				"Internal server error"
//	@Router			/api/web/login [post]
func (api *APIServer) HandleWebLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*5)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Call the EXACT SAME Auth Service used by mobile apps
	token, refreshToken, perms, err := api.Services.Auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrAccountDisabled) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("[Web Auth] Login error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// === THE WEB DIFFERENCE ===
	// Set the Refresh Token as an HTTP-Only Secure Cookie
	setWebRefreshTokenCookie(w, refreshToken)

	// Formulate Response: ONLY return the short-lived JWT and permissions in the JSON body
	resp := dto.LoginResponse{
		Token:       token,
		Permissions: perms,
	}

	RespondJSON(w, resp)
}


// HandleWebRefresh expects: POST /api/web/refresh
//
//	@Summary		Refresh Access Token
//	@Description	Exchanges a valid refresh token for a new short-lived JWT access token and updated permissions.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.RefreshRequest	true	"Refresh Token Payload"
//	@Success		200		{object}	dto.RefreshResponse	"Successfully refreshed"
//	@Failure		400		{string}	string				"Invalid JSON payload"
//	@Failure		401		{string}	string				"Invalid or expired refresh token"
//	@Failure		500		{string}	string				"Internal server error"
//	@Router			/api/refresh [post]
func (api *APIServer) HandleWebRefreshOrLogout(w http.ResponseWriter, r *http.Request) {

	// Extract the refresh token directly from the Secure Cookie
	rawRefreshToken, err := getWebRefreshTokenCookie(r)
	if err != nil {
		http.Error(w, "No refresh token found", http.StatusUnauthorized)
		return
	}

	var req dto.RefreshLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Logout {
		api.doWebLogout(w, rawRefreshToken)
		return
	}

	// Check refresh token
	accessToken, perms, err := api.Services.Auth.RefreshToken(r.Context(), rawRefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) || errors.Is(err, service.ErrAccountDisabled) {
			// If the refresh token is dead, clear the cookie so the browser forgets it
			revokeWebRefreshTokenCookie(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Printf("[Web Auth] Refresh error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the new short-lived JWT
	resp := dto.RefreshResponse{
		Token:       accessToken,
		Permissions: perms,
	}

	RespondJSON(w, resp)
}


// HandleWebLogout expects: POST /api/web/logout
// func (api *APIServer) HandleWebLogout(w http.ResponseWriter, r *http.Request) {

// 	// (Optional but recommended) Extract the cookie, hash it, and mark it as Revoked in your SQLite DB
// 	rawRefreshToken, err := getWebRefreshTokenCookie(r)
// 	if err != nil {
// 		log.Printf("[HandleWebLogout] err %v", err)
// 		http.Error(w, "No refresh token found", http.StatusUnauthorized)
// 		return
// 	}

// 	revokeResult := ""
// 	err = api.Services.Auth.RevokeRefreshToken(api.Context, rawRefreshToken)
// 	if err != nil {
// 		// http.Error(w, "No refresh token found", http.StatusUnauthorized)
// 		// return
// 		log.Printf("[Web Auth] Failed to revoke db refresh token: %v", err)
// 		revokeResult = " Warning: Failed to revoke in db. Check system log for detail."
// 	}

// 	// Tell the browser to delete the cookie
// 	revokeWebRefreshTokenCookie(w)

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message": "Logged out successfully.` + revokeResult + `"}`))
// }

func (api *APIServer) doWebLogout(w http.ResponseWriter, rawRefreshToken string) {

	revokeResult := ""
	err := api.Services.Auth.RevokeRefreshToken(api.Context, rawRefreshToken)
	if err != nil {
		// http.Error(w, "No refresh token found", http.StatusUnauthorized)
		// return
		log.Printf("[Web Auth] Failed to revoke db refresh token: %v", err)
		revokeResult = " Warning: Failed to revoke in db. Check system log for detail."
	}

	// Tell the browser to delete the cookie
	revokeWebRefreshTokenCookie(w)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out successfully.` + revokeResult + `"}`))
}