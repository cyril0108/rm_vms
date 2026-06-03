package apiserver

import (
	"net/http"
)

// HandleUpdateUserPermissions expects: GET /api/permissions
func (api *APIServer) HandleGetPermissions(w http.ResponseWriter, r *http.Request) {

	ctx := api.Context

	// Call the service layer to perform the transactional swap
	list, err := api.Services.Perms.GetAllPermissions(ctx)
	if err != nil {
		// Log error internally, return generic 500 or 404 to client
		http.Error(w, "Failed to update permissions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	RespondJSON(w, list)
}