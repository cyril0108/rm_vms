package service

import (
	"context"
	"errors"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)

var ErrCameraHasSegments = errors.New("cannot delete camera: existing video segments must be cleared first")

type CameraManagementService interface {
	// UpdateUserPermissions(ctx context.Context, adminID, targetUserID int64, permIDs []int64) error
	GetByID(ctx context.Context, id int64) (*models.Camera, error)
	GetAll(ctx context.Context) ([]*models.Camera, error)
	GetAllForInSystemCheck(ctx context.Context) ([]*models.Camera, error)
	AddCamera(ctx context.Context, cam *models.Camera) (int64, error)
	UpdateCamera(ctx context.Context, id int64, cam models.PartialUpdateInterfaces) error
	DeleteCamera(ctx context.Context, id int64) error
	ActivateCamera(ctx context.Context, id int64) error
	DeactivateCamera(ctx context.Context, id int64) error
}


func NewCameraManagementService(cRepo repository.CameraRepository, segRepo repository.SegmentRepository) CameraManagementService {
	return &cameraServiceBase{repo: cRepo, segRepo: segRepo}
}

func (s *cameraServiceBase) GetByID(ctx context.Context, id int64) (*models.Camera, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *cameraServiceBase) GetAll(ctx context.Context) ([]*models.Camera, error) {
	return s.repo.GetAll(ctx)
}

func (s *cameraServiceBase) GetAllForInSystemCheck(ctx context.Context) ([]*models.Camera, error) {
	return s.repo.GetAllForInSystemCheck(ctx)
}

func (s *cameraServiceBase) AddCamera(ctx context.Context, cam *models.Camera) (int64, error) {
	return s.repo.Create(ctx, cam)
}

func (s *cameraServiceBase) UpdateCamera(ctx context.Context, id int64, cam models.PartialUpdateInterfaces) error {
	return s.repo.UpdatePartial(ctx, id, cam)
}

func (s *cameraServiceBase) DeleteCamera(ctx context.Context, id int64) error {
	hasSegments, err := s.segRepo.HasSegments(ctx, id)
	if err != nil {
		return err // Database error during check
	}

	if hasSegments {
		return ErrCameraHasSegments
	}

	return s.repo.Delete(ctx, id)
}

func (s *cameraServiceBase) ActivateCamera(ctx context.Context, id int64) error {
	return s.repo.Activate(ctx, id)
}

func (s *cameraServiceBase) DeactivateCamera(ctx context.Context, id int64) error {
	return s.repo.Deactivate(ctx, id)
}
