package dto

import "nvr_core/db/models"

// ===========================
// User Manage
// ===========================
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	RoleId   int64  `json:"role_id"`
}

func (cu *CreateUserRequest) MapToDBUser() *models.User {
	return  &models.User{
		Username: cu.Username,
		Password: cu.Password,
		Email: cu.Email,
		RoleID: cu.RoleId,
	}
}

// ===========================
// Permissions
// ===========================
type UpdatePermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}