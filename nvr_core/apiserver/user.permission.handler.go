package apiserver

import (
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
)

// HandleGetAllRoles retrieves all available user roles.
//
//	@Summary		Get all roles
//	@Description	Retrieves a list of all roles defined in the system. Requires MANAGE_USERS permission.
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{array}		models.Role	"List of roles"
//	@Failure		401		{string}	string		"Unauthorized"
//	@Failure		403		{string}	string		"Forbidden - Insufficient permissions"
//	@Failure		500		{string}	string		"Internal server error"
//	@Router			/api/roles [get]
func (api *APIServer) HandleGetAllRoles(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	if session, ok := middleware.GetSession(ctx);
		!ok || !session.HasPermissionUserManage() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	roles, err := api.Services.Perms.GetAllRoles(ctx)
	if err != nil {
		http.Error(w, "Failed to get all roles", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, roles)


}

// HandleGetAllPermissions retrieves all available system permissions.
//
//	@Summary		Get all permissions
//	@Description	Retrieves a list of all permissions that can be assigned to users or roles. Requires MANAGE_USERS permission.
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{array}		models.Permission	"List of permissions"
//	@Failure		401		{string}	string				"Unauthorized"
//	@Failure		403		{string}	string				"Forbidden - Insufficient permissions"
//	@Failure		500		{string}	string				"Internal server error"
//	@Router			/api/permissions [get]
func (api *APIServer) HandleGetAllPermissions(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	if session, ok := middleware.GetSession(ctx);
		!ok || !session.HasPermissionUserManage() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	perms, err := api.Services.Perms.GetAllPermissions(ctx)
	if err != nil {
		http.Error(w, "Failed to get all permissions", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, perms)

}

// HandleUpdateUserPermissions expects: PUT /api/users/{id}/permissions
//
//	@Summary		Update user explicit permissions
//	@Description	Overwrites a user's direct permission grants. Cannot be used to modify one's own permissions. Requires MANAGE_USERS permission.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Target User ID"
//	@Param			payload	body		dto.UpdatePermissionsRequest	true	"Array of permission IDs"
//	@Success		200		{object}	map[string]string				"Permissions updated successfully"
//	@Failure		400		{string}	string							"Invalid user ID or payload"
//	@Failure		401		{string}	string							"Unauthorized"
//	@Failure		403		{string}	string							"Forbidden - Cannot manage own permissions or insufficient access"
//	@Failure		500		{string}	string							"Internal server error"
//	@Router			/api/users/{id}/permissions [put]
func (api *APIServer) HandleUpdateUserPermissions(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req *dto.UpdatePermissionsRequest
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Call the service layer to perform the transactional swap
	err = api.Services.User.UpdateUserPermissions(r.Context(), session.UserID, targetUserID, req.PermissionIDs)
	if err != nil {
		// Log error internally, return generic 500 or 404 to client
		http.Error(w, "Failed to update permissions", http.StatusInternalServerError)
		return
	}

	api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)

	w.WriteHeader(http.StatusOK)
	RespondJSON(w, true)
}