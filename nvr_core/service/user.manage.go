package service

import (
	"context"
	"fmt"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/security"
)

type UserManagementService interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, name string) (*models.User, error)

	// User Management
	GetAllUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error)
	CreateUser(ctx context.Context, adminID int64, user *models.User) error
	UpdateUserPassword(ctx context.Context, adminID, userID int64, encrypt string) error
	UpdatePartial(ctx context.Context, adminID, userID int64, updates models.PartialUpdateInterfaces) error
	DeactivateUser(ctx context.Context, adminID, targetUserID int64) error

	// Permissions
	UpdateUserRole(ctx context.Context, adminID, targetUserID, newRoleID int64) error
	GrantPermission(ctx context.Context, adminID, targetUserID, permID int64) error
	RevokePermission(ctx context.Context, adminID, targetUserID, permID int64) error
	UpdateUserPermissions(ctx context.Context, adminID, targetUserID int64, permIDs []int64) error
}

func NewUserManagementService(uRepo repository.UserRepository, pRepo repository.PermissionRepository) UserManagementService {
	return &userServiceBase{userRepo: uRepo, permRepo: pRepo}
}

func (s *userServiceBase) CreateUser(ctx context.Context, adminID int64, user *models.User) error {

	hashed, err := security.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}
	user.Password = hashed

	return s.userRepo.Create(ctx, user)
}

func (s *userServiceBase) UpdatePartial(ctx context.Context, adminID, userID int64, updates models.PartialUpdateInterfaces) error {
	return s.userRepo.UpdatePartial(ctx, userID, updates)
}

func (s *userServiceBase) UpdateUserPassword(ctx context.Context, adminID, userID int64, password string) error {

	hashed, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}

	// Business Rule: Ensure target user actually exists before modifying
	if err := s.userRepo.UpdatePassword(ctx, userID, hashed); err != nil {
		return fmt.Errorf("update user password failed: %w", err)
	}
	return nil
}

func (s *userServiceBase) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userServiceBase) GetByUsername(ctx context.Context, name string) (*models.User, error) {
	return s.userRepo.GetByUsername(ctx, name)
}


// ==============================================================
// User Management
// ==============================================================
func (s *userServiceBase) GetAllUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default reasonable limit
	}

	offset := (page - 1) * pageSize
	
	users, err := s.userRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *userServiceBase) DeactivateUser(ctx context.Context, adminID, targetUserID int64) error {
	// Business Rule: Prevent admins from deactivating themselves
	if adminID == targetUserID {
		return fmt.Errorf("cannot deactivate your own account")
	}
	return s.userRepo.Deactivate(ctx, targetUserID)
}


// ==============================================================
// User Permissions
// ==============================================================
func (s *userServiceBase) UpdateUserRole(ctx context.Context, adminID, targetUserID, newRoleID int64) error {
	// Business Rule: Ensure target user actually exists before modifying
	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return fmt.Errorf("target user verification failed: %w", err)
	}
	return s.userRepo.UpdateRole(ctx, targetUserID, newRoleID)
}

func (s *userServiceBase) GrantPermission(ctx context.Context, adminID, targetUserID, permID int64) error {
	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return err
	}
	return s.permRepo.GrantUserPermission(ctx, targetUserID, permID)
}

func (s *userServiceBase) RevokePermission(ctx context.Context, adminID, targetUserID, permID int64) error {
	return s.permRepo.RevokeUserPermission(ctx, targetUserID, permID)
}

func (s *userServiceBase) UpdateUserPermissions(ctx context.Context, adminID, targetUserID int64, permIDs []int64) error {
	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return err
	}
	// Delegate the bulk replacement to the transactional repository method
	return s.permRepo.ReplaceUserPermissions(ctx, targetUserID, permIDs)
}