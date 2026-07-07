package service

import (
	"context"
	"database/sql"
	"nvr_core/db/repository"
	"os"
	"path/filepath"
)

// Services acts as a dependency injection container for the API layer.
// The API layer knows NOTHING about SQLite or Repositories, only these interfaces.
// Even though, it's a bridge between API process and Repositories.
type Services struct {
	Auth       AuthService
	License    LicenseService
	Perms      PermsService
	User       UserManagementService
	Camera     CameraManagementService
	Timeline   TimelineService
	Playback   PlaybackService
	Playlist   PlaylistService
	System     SystemService
	SysSetting SystemSettingService
	Export     ExportService
	ExportTM   *ExportTaskManager
	Maintain   MaintainService
}



func NewServices(dbConn *sql.DB) *Services {

	segRepo  := repository.NewSegmentRepository(dbConn)
	userRepo  := repository.NewUserRepository(dbConn)
	retknRepo := repository.NewRefreshTokenRepository(dbConn)
	permRepo  := repository.NewPermissionRepository(dbConn)
	cameraRepo := repository.NewCameraRepository(dbConn)
	sssRepo    := repository.NewSystemSettingsRepository(dbConn)
	licRepo    := repository.NewLicenseRepository(dbConn)
	timelineSvc := NewTimelineService(segRepo)
	playbackSvc := NewPlaybackService(segRepo)
	playlistSvc := NewPlaylistService(segRepo)
	systemSvc := NewSystemService(dbConn, userRepo)
	// Some random secret key for now
	authSvc := NewAuthService(userRepo, permRepo, retknRepo, ")($#YHdsJdsx")
	userSvc := NewUserManagementService(userRepo, permRepo)
	permSvc := NewPermsService(permRepo)
	camSvc := NewCameraManagementService(cameraRepo, segRepo)
	sssSvc := NewSysSettingService(sssRepo)
	licSvc := NewLicenseService(licRepo)
	exportSvc := NewExportService(segRepo)
	exportTMSvc := NewExportTaskManager()
	maintainSvc := NewMaintainService(segRepo)

	return &Services{
		Auth:     authSvc,
		License:  licSvc,
		Perms:    permSvc,
		User:     userSvc,
		Camera:   camSvc,
		Timeline: timelineSvc,
		Playback: playbackSvc,
		Playlist: playlistSvc,
		System:   systemSvc,
		SysSetting: sssSvc,
		Export: exportSvc,
		ExportTM: exportTMSvc,
		Maintain: maintainSvc,
	}
}

func StartIngester(ctx context.Context, dbConn *sql.DB) IngestService {

	repo := repository.NewSegmentRepository(dbConn)

	// Initialize the Global BatchIngester
	// Buffer 200 segments, flush to disk in batches of 50
	ingester := NewBatchIngester(repo, 200, 50)
	go ingester.Start(ctx)

	return ingester

}

// Run this exactly once when initializing your service
func CleanExportsOnBoot(rootPath string) error {
	exportDir := filepath.Join(rootPath, "export")

	// os.RemoveAll is extremely fast and handles non-existent directories gracefully
	if err := os.RemoveAll(exportDir); err != nil {
		return err
	}

	// Recreate the empty directory so it's ready for new tasks
	return os.MkdirAll(exportDir, 0755)
}


func StartRetentionWatcher(ctx context.Context, dbConn *sql.DB, path string, cameraProvider ActiveCameraProvider) {

	LOG.Info("[StartRetentionWatcher]", "path", path)

	repo := repository.NewSegmentRepository(dbConn)

	retention := NewRetentionService(repo, path, cameraProvider)

	go retention.StartDiskWatchdog(ctx)

}
