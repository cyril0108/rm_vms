package service

import (
	"database/sql"
	"time"

	"nvr_core/db/repository"
	"nvr_core/logger"
)

var LOG = logger.NewLogger("\033[35m[nvr_core][service]\033[0m")

type authServiceBase struct {
	userRepo    repository.UserRepository
	reTokenRepo repository.RefreshTokenRepository
	permRepo    repository.PermissionRepository
	jwtSecret   []byte
	tokenExpir  time.Duration
	userStatus  *UserStatusMap
}

type userServiceBase struct {
	userRepo repository.UserRepository
	permRepo repository.PermissionRepository
}

type bookmarkServiceBase struct {
	repo repository.BookmarkRepository
}

type layoutServiceBase struct {
	repo repository.LayoutRepository
}

type eventServiceBase struct {
	repo repository.EventRepository
}

type cameraServiceBase struct {
	repo repository.CameraRepository
	segRepo repository.SegmentRepository
}

type permissionsServiceBase struct {
	repo repository.PermissionRepository
}

type segmentServiceBase struct {
	repo repository.SegmentRepository
}

type systemServiceBase struct {
	db *sql.DB
	repo repository.UserRepository
}

type sysSettingServiceBase struct {
	repo repository.SystemSettingsRepository
}

type licenseServiceBase struct {
	repo repository.LicenseRepository
}

type maintainServiceBase struct {
	repo repository.SegmentRepository
}
