package service

import (
	"context"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)

type CameraManagementService interface {
	// UpdateUserPermissions(ctx context.Context, adminID, targetUserID int64, permIDs []int64) error
	GetByID(ctx context.Context, id string) (*models.Camera, error)
	GetAll(ctx context.Context) ([]*models.Camera, error)
	AddCamera(ctx context.Context, cam *models.Camera) error
	UpdateCamera(ctx context.Context, cam *models.Camera) error
	ActivateCamera(ctx context.Context, id string) error
	DeactivateCamera(ctx context.Context, id string) error
}


func NewCameraManagementService(cRepo repository.CameraRepository) CameraManagementService {
	return &cameraServiceBase{repo: cRepo}
}

func (s *cameraServiceBase) GetByID(ctx context.Context, id string) (*models.Camera, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *cameraServiceBase) GetAll(ctx context.Context) ([]*models.Camera, error) {
	return s.repo.GetAll(ctx)
}

func (s *cameraServiceBase) AddCamera(ctx context.Context, cam *models.Camera) error {
	return s.repo.Create(ctx, cam)
}

func (s *cameraServiceBase) UpdateCamera(ctx context.Context, cam *models.Camera) error {
	return s.repo.Update(ctx, cam)
}

func (s *cameraServiceBase) ActivateCamera(ctx context.Context, id string) error {
	return s.repo.Activate(ctx, id)
}

func (s *cameraServiceBase) DeactivateCamera(ctx context.Context, id string) error {
	return s.repo.Deactivate(ctx, id)
}
