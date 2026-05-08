package service

import (
	"database/sql"
	"nvr_core/db/repository"
	"time"
)

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

type debugServiceBase struct {
	db *sql.DB
	repo repository.SegmentRepository
}
