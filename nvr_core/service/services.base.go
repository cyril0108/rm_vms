package service

import (
	"database/sql"
	"nvr_core/db/repository"
	"nvr_core/logger"
	"time"
)

var LOG = logger.NewLogger("[nvr_core]","service")

type authServiceBase struct {
	userRepo   repository.UserRepository
	permRepo   repository.PermissionRepository
	jwtSecret  []byte
	tokenExpir time.Duration
}

type userServiceBase struct {
	userRepo repository.UserRepository
	permRepo repository.PermissionRepository
}

type cameraServiceBase struct {
	repo repository.CameraRepository
	segRepo repository.SegmentRepository
}

type segmentServiceBase struct {
	repo repository.SegmentRepository
}

type systemServiceBase struct {
	db *sql.DB
	repo repository.UserRepository
}
