package service

import (
	"context"
	// "database/sql"
	// "errors"

	// "nvr_core/apiserver/dto"
	// "nvr_core/db/models"
	"nvr_core/db/models"
	"nvr_core/db/repository"
)


type PermsService interface {
	GetAllPermissions(ctx context.Context) ([]*models.Permission, error)
	GetAllRoles(ctx context.Context) ([]*models.Role, error)
}

func NewPermsService(repo repository.PermissionRepository) PermsService {
	return &permissionsServiceBase{repo: repo}
}

func (p *permissionsServiceBase) GetAllPermissions(ctx context.Context) ([]*models.Permission, error) {
	return p.repo.GetAll(ctx)
}

func (p *permissionsServiceBase) GetAllRoles(ctx context.Context) ([]*models.Role, error) {
	return p.repo.GetAllRoles(ctx)
}
