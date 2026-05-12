package apiserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/service"
)

// HandleLogin expects: POST /api/login
func (api *APIServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Limit body size and parse JSON
	r.Body = http.MaxBytesReader(w, r.Body, 1024*5) // 5KB limit

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Call the Auth Service
	token, perms, err := api.Services.Auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrAccountDisabled) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("[Auth API] Login error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}



	// Formulate Response
	resp := dto.LoginResponse{
		Token:       token,
		Permissions: perms,
	}

	RespondJSON(w, resp)
}