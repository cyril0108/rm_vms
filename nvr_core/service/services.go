package service

import (
	"context"
	"database/sql"
	"nvr_core/db/repository"
)

// Services acts as a dependency injection container for the API layer.
// The API layer knows NOTHING about SQLite or Repositories, only these interfaces.
// Even though, it's a bridge between API process and Repositories.
type Services struct {
	Auth       AuthService
	User       UserManagementService
	Camera     CameraManagementService
	Timeline   TimelineService
	Playback   PlaybackService
	Playlist   PlaylistService
	System     SystemService
}



func NewServices(dbConn *sql.DB) *Services {

	segRepo := repository.NewSegmentRepository(dbConn)
	userRepo := repository.NewUserRepository(dbConn)
	permRepo := repository.NewPermissionRepository(dbConn)
	cameraRepo := repository.NewCameraRepository(dbConn)
	timelineSvc := NewTimelineService(segRepo)
	playbackSvc := NewPlaybackService(segRepo)
	playlistSvc := NewPlaylistService(segRepo)
	systemSvc := NewSystemService(dbConn, userRepo)
	// Some random secret key for now
	authSvc := NewAuthService(userRepo, permRepo, ")($#YHdsJdsx")
	userSvc := NewUserManagementService(userRepo, permRepo)
	camSvc := NewCameraManagementService(cameraRepo, segRepo)

	return &Services{
		Auth:     authSvc,
		User:     userSvc,
		Camera:   camSvc,
		Timeline: timelineSvc,
		Playback: playbackSvc,
		Playlist: playlistSvc,
		System: systemSvc,
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
