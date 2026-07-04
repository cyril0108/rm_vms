package apiserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/service"
	"nvr_core/utils"
)

// HandleLogin authenticates a user and provides session tokens.
//
//	@Summary		User login
//	@Description	Authenticates a user by username and password. Returns a short-lived JWT access token, a long-lived refresh token, and a list of user permissions.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		dto.LoginRequest	true	"User Login Credentials"
//	@Success		200			{object}	utils.APIResponse{ dto.LoginResponse }	"Successfully authenticated"
//	@Failure		400			{string}	string				"Invalid JSON payload"
//	@Failure		401			{string}	string				"Invalid credentials or account disabled"
//	@Failure		500			{string}	string				"Internal server error"
//	@Router			/api/login [post]
func (api *APIServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Limit body size and parse JSON
	r.Body = http.MaxBytesReader(w, r.Body, 1024*5) // 5KB limit

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Call the Auth Service
	user, token, refresh, perms, err := api.Services.Auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrAccountDisabled) {
			utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("[Auth API] Login error: %v", err)
		utils.RespondJSONHTTPStatus(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Formulate Response
	resp := dto.LoginResponse{
		User: *dto.NewLoginUser(user),
		Token:        token,
		RefreshToken: refresh,
		Permissions:  perms,
	}

	utils.RespondJSON(w, resp, "")
}


// HandleRefresh processes the silent refresh flow.
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
func (api *APIServer) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	// Protect against oversized payloads
	r.Body = http.MaxBytesReader(w, r.Body, 1024*5) 

	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Hit the service layer
	accessToken, perms, err := api.Services.Auth.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) || errors.Is(err, service.ErrAccountDisabled) {
			utils.RespondJSONHTTPStatus(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Printf("[Auth API] Refresh error: %v", err)
		utils.RespondJSONHTTPStatus(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Formulate and send response
	resp := dto.RefreshResponse{
		Token:       accessToken,
		Permissions: perms,
	}

	utils.RespondJSON(w, resp, "")
}

// HandleAppLogout expects: POST /api/logout
//
//	@Summary		App Logout
//	@Description	Invalidates a specific mobile/desktop session by revoking its refresh token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.RefreshRequest	true	"Refresh Token to revoke"
//	@Success		200		{string}	string					"Successfully logged out"
//	@Router			/api/logout [post]
func (api *APIServer) HandleLogout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*5)

	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Hit the service layer to revoke the token
	err := api.Services.Auth.RevokeRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Log the error but still return 200 OK to the client. 
		// If they are trying to log out, we shouldn't block them with server errors.
		log.Printf("[Auth API] App Logout error: %v", err)
	}

	utils.RespondJSON(w, "", "Logged out successfully")
}