package dto

import (
	"nvr_core/db/models"
)

// ===========================
// User Manage
// ===========================
type CreateUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	RoleID   int64  `json:"role_id"`
	IsActive *bool  `json:"is_active"`
}

func (cu *CreateUserRequest) MapToDBUser() *models.User {
	u := &models.User{
		Username: cu.Username,
		Name: cu.Name,
		Password: cu.Password,
		Email: cu.Email,
		RoleID: cu.RoleID,
	}
	if cu.IsActive != nil {
		u.IsActive = *cu.IsActive
	} else {
		u.IsActive = true
	}
	return u
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
	RoleId   *int64  `json:"role_id,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

func (ur *UpdateUserRequest) MapToPartialInterface() models.PartialUpdateInterfaces {

	updates := make(models.PartialUpdateInterfaces)

	if ur.Name      != nil { updates["name"] = *ur.Name }
	if ur.Password  != nil { updates["password"] = *ur.Password }
	if ur.Email     != nil { updates["email"] = *ur.Email }
	if ur.RoleId    != nil { updates["role_id"] = *ur.RoleId }
	if ur.IsActive  != nil { updates["is_active"] = *ur.IsActive }

	return updates

}


// ===========================
// Permissions
// ===========================
type UpdatePermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}