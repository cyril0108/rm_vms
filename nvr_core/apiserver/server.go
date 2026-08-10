package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"nvr_core/logger"
	"nvr_core/apiserver/middleware"
	"nvr_core/apiserver/webserver"
	"nvr_core/network"
	"nvr_core/process"
	"nvr_core/security/cert"
	"nvr_core/service"
	"nvr_core/utils"
)

// NVRState uses sync.Map for highly concurrent, lock-free (mostly) reads/writes
type NVRState struct {
	Cameras sync.Map
}

func NewNVRState() *NVRState {
	return &NVRState{}
}

type APIServer struct {
	logger *logger.Logger
	Context context.Context
	CFG   *utils.Config
	State *NVRState
	PM    *process.Manager
	Services *service.Services
}

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func Initiate(ctx context.Context, cfg *utils.Config, pm *process.Manager, svcs *service.Services) {

	LOG.Info("Initializing API server")

	state := NewNVRState()

	api := &APIServer{
		logger: LOG,
		Context: ctx,
		State: state,
		CFG: cfg,
		PM: pm,
		Services: svcs,
	}

	mux := http.NewServeMux()

	authMid := middleware.RequireAuth(api.Services.Auth)

	// Mount configure
	// if health, err := svcs.System.GetHealthData(ctx);
	//   err == nil && !health.Configured {
	//   LOG.Info("NVR account not configured. Init with configure API.")
	// }

	// =============================================
	// System
	// =============================================
	mux.HandleFunc("GET /health", api.GetHealth)
	mux.HandleFunc("POST /api/configure", api.HandleAdminInitConfigure)


	// Debug Info
	mux.HandleFunc("GET /debug/db", authMid(http.HandlerFunc(api.GetDebugInfo)))

	// =============================================
	// Server Set
	// =============================================
	mux.HandleFunc("PUT /api/config/servername", authMid(http.HandlerFunc(api.HandleSetServerName)))


	// =============================================
	// Hardware and License
	// =============================================
	// Return machine id and server name
	mux.HandleFunc("GET /api/machine", authMid(http.HandlerFunc(api.HandleGetMachineInfo)))
	mux.HandleFunc("POST /api/license", authMid(http.HandlerFunc(api.HandleReceiveLicense)))
	mux.HandleFunc("POST /api/license/key", authMid(http.HandlerFunc(api.HandleReceiveLicenseKey)))

	mux.HandleFunc("GET /api/license/all", authMid(http.HandlerFunc(api.HandleGetLicenseList)))
	mux.HandleFunc("GET /api/license/status", authMid(http.HandlerFunc(api.HandleGetActiveLicenseStatus)))

	mux.HandleFunc("GET /api/estimate/recording", authMid(http.HandlerFunc(api.HandleGetRecordingEstimation)))

	mux.HandleFunc("GET /api/shmmetrics", authMid(http.HandlerFunc(api.HandleGetSHMMetrics)))
	mux.HandleFunc("GET /api/usage", authMid(http.HandlerFunc(api.HandleGetSystemMetrics)))
	mux.HandleFunc("GET /api/network", authMid(http.HandlerFunc(api.HandleGetNetworkInfo)))


	// =============================================
	// Maintenance
	// =============================================
	mux.HandleFunc("GET /api/maintain/segments", authMid(http.HandlerFunc(api.GetAbnormalSegments)))
	mux.HandleFunc("PUT /api/maintain/segments", authMid(http.HandlerFunc(api.FixAbnormalSegments)))


	// =============================================
	// Email Settings
	// =============================================
	mux.HandleFunc("GET /api/email/smtp", authMid(http.HandlerFunc(api.HandleGetSMTPSettings)))
	mux.HandleFunc("PUT /api/email/smtp", authMid(http.HandlerFunc(api.HandleUpdateSMTPSettings)))
	mux.HandleFunc("POST /api/email/test", authMid(http.HandlerFunc(api.HandleTestEmail)))

	mux.HandleFunc("GET /api/email/groups", authMid(http.HandlerFunc(api.HandleListEmailGroups)))
	mux.HandleFunc("POST /api/email/groups", authMid(http.HandlerFunc(api.HandleCreateEmailGroup)))
	mux.HandleFunc("PUT /api/email/groups/{id}", authMid(http.HandlerFunc(api.HandleUpdateEmailGroup)))
	mux.HandleFunc("DELETE /api/email/groups/{id}", authMid(http.HandlerFunc(api.HandleDeleteEmailGroup)))


	// =============================================
	// Login
	// =============================================
	mux.HandleFunc("POST /api/login", api.HandleLogin)
	mux.HandleFunc("POST /api/logout", authMid(http.HandlerFunc(api.HandleLogout)))
	mux.HandleFunc("POST /api/refresh", api.HandleRefresh)

	mux.HandleFunc("POST /api/web/login", api.HandleWebLogin)
	// mux.HandleFunc("POST /api/web/logout", api.HandleWebLogout)
	mux.HandleFunc("POST /api/web/refresh", api.HandleWebRefreshOrLogout)

	// =============================================
	// Events
	// =============================================
	mux.HandleFunc("GET /api/events", authMid(http.HandlerFunc(api.HandleGetEvents)))


	// =============================================
	// User Management
	// =============================================
	mux.HandleFunc("GET /api/admin/users", authMid(http.HandlerFunc(api.HandleListUsers)))

	mux.HandleFunc("GET /api/users/me", authMid(http.HandlerFunc(api.HandleGetLoginUser)))

	mux.HandleFunc("GET /api/admin/permissions", authMid(http.HandlerFunc(api.HandleGetAllPermissions)))
	mux.HandleFunc("GET /api/admin/roles", authMid(http.HandlerFunc(api.HandleGetAllRoles)))

	mux.HandleFunc("POST /api/admin/users/create", authMid(http.HandlerFunc(api.HandleCreateUser)))
	mux.HandleFunc("PUT /api/admin/users/{id}", authMid(http.HandlerFunc(api.HandleUpdateUser)))
	mux.HandleFunc("PUT /api/admin/users/{id}/password", authMid(http.HandlerFunc(api.HandleUpdateUserPassword)))
	mux.HandleFunc("PUT /api/admin/users/{id}/permissions", authMid(http.HandlerFunc(api.HandleUpdateUserPermissions)))

	mux.HandleFunc("DELETE /api/users/{id}", authMid(http.HandlerFunc(api.HandleDeactivateUser)))



	// =============================================
	// Camera Discovery
	// =============================================
	mux.HandleFunc("GET /api/scan", (http.HandlerFunc(api.HandleCameraScan)))
	mux.HandleFunc("GET /api/scansweep", (http.HandlerFunc(api.HandleCameraSweep)))
	mux.HandleFunc("POST /api/scansweep/detail", (http.HandlerFunc(api.HandleBulkONVIFScan)))

	mux.HandleFunc("GET /api/scan/{ip}", (http.HandlerFunc(api.HandleCameraProbe)))
	mux.HandleFunc("POST /api/scan/{ip}/onvif", (http.HandlerFunc(api.HandleFetchCameraONVIF)))


	// =============================================
	// Camera stream
	// =============================================
	// mux.HandleFunc("GET /ws/stream/{cam_id}", authMid(http.HandlerFunc(api.GetStream)))
	// mux.HandleFunc("GET /ws/stream/{cam_id}/{profile}", authMid(http.HandlerFunc(api.GetStream)))

	// NO LOGIN FOR TEST
	// mux.HandleFunc("GET /live/camera/{cam_id}", authMid(http.HandlerFunc(api.HandleLiveTransmuxTS)))
	// mux.HandleFunc("GET /live/camera/{cam_id}/{profile}", authMid(http.HandlerFunc(api.HandleLiveTransmuxTS)))
	mux.HandleFunc("GET /live/camera/{cam_id}", api.HandleLiveTransmuxTS)
	mux.HandleFunc("GET /live/camera/{cam_id}/{profile}", api.HandleLiveTransmuxTS)

	// =============================================
	// Camera Bookmark
	// =============================================
	mux.HandleFunc("GET /api/bookmark", authMid(http.HandlerFunc(api.GetBookmarks)))
	mux.HandleFunc("GET /api/bookmark/{cam_id}", authMid(http.HandlerFunc(api.GetCameraBookmarks)))
	mux.HandleFunc("POST /api/bookmark", authMid(http.HandlerFunc(api.HandleCreateBookmark)))


	// =============================================
	// Camera PTZ
	// =============================================
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/left", authMid(http.HandlerFunc(api.HandleCameraPTZStepLeft)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/right", authMid(http.HandlerFunc(api.HandleCameraPTZStepRight)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/up", authMid(http.HandlerFunc(api.HandleCameraPTZStepUp)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/down", authMid(http.HandlerFunc(api.HandleCameraPTZStepDown)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/in", authMid(http.HandlerFunc(api.HandleCameraPTZStepZoomIn)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/step/out", authMid(http.HandlerFunc(api.HandleCameraPTZStepZoomOut)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/center", authMid(http.HandlerFunc(api.HandleCameraPTZCenter)))

	mux.HandleFunc("GET /api/ptz/{cam_id}/home", authMid(http.HandlerFunc(api.CameraPTZToHome)))

	mux.HandleFunc("GET /api/ptz/{cam_id}/profile", authMid(http.HandlerFunc(api.CameraProfile)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/caps", authMid(http.HandlerFunc(api.CameraCapabilities)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/services", authMid(http.HandlerFunc(api.CameraServices)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/status", authMid(http.HandlerFunc(api.CameraPTZStatus)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/config", authMid(http.HandlerFunc(api.CameraConfigurations)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/configs", authMid(http.HandlerFunc(api.CameraPTZConfiguration)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/presets", authMid(http.HandlerFunc(api.CameraPTZPresets)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/preset/{token}", authMid(http.HandlerFunc(api.CameraPTZToPreset)))

	mux.HandleFunc("GET /api/ptz/{cam_id}/abs", authMid(http.HandlerFunc(api.CameraPTZAbsolute)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/rel", authMid(http.HandlerFunc(api.CameraPTZRelative)))

	mux.HandleFunc("GET /api/ptz/{cam_id}/cnt", authMid(http.HandlerFunc(api.CameraPTZContinuous)))
	mux.HandleFunc("GET /api/ptz/{cam_id}/stp", authMid(http.HandlerFunc(api.CameraPTZStop)))


	// =============================================
	// Camera Management
	// =============================================
	mux.HandleFunc("GET /api/cameras", authMid(http.HandlerFunc(api.GetCameras)))
	mux.HandleFunc("GET /api/cameras/db", authMid(http.HandlerFunc(api.GetDBCameras)))
	mux.HandleFunc("GET /api/cameras/{id}/onvif", authMid(http.HandlerFunc(api.HandleFetchSystemCameraONVIF)))

	mux.HandleFunc("POST /api/cameras/add", authMid(http.HandlerFunc(api.AddCamera)))
	mux.HandleFunc("POST /api/cameras/add/onvif", authMid(http.HandlerFunc(api.AddONVIFCamera)))
	mux.HandleFunc("PUT /api/cameras/{cam_id}/update", authMid(http.HandlerFunc(api.UpdateCamera)))
	mux.HandleFunc("PUT /api/cameras/{cam_id}/onvif", authMid(http.HandlerFunc(api.UpdateCameraONVIF)))
	mux.HandleFunc("PUT /api/cameras/{cam_id}/auth", authMid(http.HandlerFunc(api.UpdateCameraAuth)))
	mux.HandleFunc("POST /api/cameras/{cam_id}/update", authMid(http.HandlerFunc(api.UpdateCamera)))
	mux.HandleFunc("POST /api/cameras/{cam_id}/stop", authMid(http.HandlerFunc(api.DeactivateCamera)))
	mux.HandleFunc("POST /api/cameras/{cam_id}/start", authMid(http.HandlerFunc(api.ActivateCamera)))

	mux.HandleFunc("DELETE /api/cameras/{cam_id}", authMid(http.HandlerFunc(api.DeleteCamera)))


	// =============================================
	// Timeline and Playback
	// =============================================
	mux.HandleFunc("GET /api/cameras/{cam_id}/timeline/{start}/{end}", authMid(http.HandlerFunc(api.GetTimeline)))
	mux.HandleFunc("GET /api/cameras/{cam_id}/snapshot", authMid(http.HandlerFunc(api.HandleSegmentSnapshot)))

	mux.HandleFunc("GET /api/cameras/{cam_id}/timeline/segs", authMid(http.HandlerFunc(api.GetProfileSegments)))
	mux.HandleFunc("GET /api/cameras/{cam_id}/snapshot/range", authMid(http.HandlerFunc(api.GetSegmentSnapshots)))

	// mux.HandleFunc("GET /api/cameras/{cam_id}/play", authMid(http.HandlerFunc(api.HandlePlayVideo)))
	// mux.HandleFunc("GET /api/cameras/{cam_id}/play/ts", authMid(http.HandlerFunc(api.HandleTransmuxTS)))
	mux.HandleFunc("GET /api/cameras/{cam_id}/play", api.HandlePlayVideo)
	mux.HandleFunc("GET /api/cameras/{cam_id}/play/ts", api.HandleTransmuxTS)
	// GET /api/cameras/{id}/play/gap?duration=5000
	mux.HandleFunc("GET /api/cameras/{cam_id}/play/gap", api.HandleGapFillerTS)

	// =============================================
	// Calendar
	// =============================================
	mux.HandleFunc("GET /api/cameras/{cam_id}/summary", authMid(http.HandlerFunc(api.HandleGetDailySummary)))


	// =============================================
	// Playlist
	// =============================================
	// mux.HandleFunc("GET /api/cameras/{cam_id}/playlist.m3u8", authMid(http.HandlerFunc(api.HandleGetPlaylist)))
	// mux.HandleFunc("GET /api/cameras/{cam_id}/playlist/ts.m3u8", authMid(http.HandlerFunc(api.HandleGetTSPlaylist)))
	mux.HandleFunc("GET /api/cameras/{cam_id}/playlist.m3u8", api.HandleGetPlaylist)
	mux.HandleFunc("GET /api/cameras/{cam_id}/playlist/ts.m3u8", api.HandleGetTSPlaylist)


	// =============================================
	// Export
	// =============================================
	mux.HandleFunc("GET /api/export/{cam_id}", authMid(http.HandlerFunc(api.HandleExportRequest)))
	mux.HandleFunc("GET /api/export/{task_id}/status", authMid(http.HandlerFunc(api.HandleExportTaskStatus)))
	mux.HandleFunc("GET /api/export/{task_id}/download", authMid(http.HandlerFunc(api.HandleDownloadExport)))
	mux.HandleFunc("POST /api/export/{cam_id}/watermark", authMid(http.HandlerFunc(api.HandleExportWatermarkRequest)))



	// =============================================
	// Serve Web
	// =============================================
	webserver.ServeWeb(mux, cfg)

	// =============================================
	// =============================================
	// =============================================

	srv := api.initServer(mux, cfg.Server.Port)
	srvTLS := api.initServerTLS(mux, cfg.Server.Port+1)


	// =============================================
	// Shutting Down
	// =============================================
	// Block this function until the parent context is canceled (SIGTERM/SIGINT)
	<-ctx.Done()

	log.Println("[API] Shutdown signal received. Finishing active requests...")


	// Create a secondary timeout context specifically for the shutdown process.
	// This ensures a malicious or ultra-slow client can't hold the server open forever.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute the graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[API] Server forced to shutdown due to timeout: %v", err)
	} else {
		log.Println("[API] Server gracefully stopped.")
	}

	if err := srvTLS.Shutdown(shutdownCtx); err != nil {
		log.Printf("[API] Server HTTPS forced to shutdown due to timeout: %v", err)
	} else {
		log.Println("[API] Server HTTPS gracefully stopped.")
	}

}

func (s *APIServer) initServer(mux *http.ServeMux, port int) *http.Server {

	addr := fmt.Sprintf(":%d", port);

	handlerWithCORS := middleware.CORSMiddleware(mux)

	// Server configuration
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlerWithCORS,
		ReadTimeout:  5 * time.Second,  // Max time to read request headers/body
		WriteTimeout: 10 * time.Second, // Max time to write response
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}

	// Start the server in a background goroutine
	go func() {
		log.Printf("[API] Server listening on %s", addr)
		// ErrServerClosed is expected when we call srv.Shutdown()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] Server critically failed: %v", err)
		}
	}()

	return srv

}

func (s *APIServer) initServerTLS(mux *http.ServeMux, port int) *http.Server {

	addr := fmt.Sprintf(":%d", port);

	handlerWithCORS := middleware.CORSMiddleware(mux)

	// Server configuration
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlerWithCORS,
		ReadTimeout:  5 * time.Second,  // Max time to read request headers/body
		WriteTimeout: 10 * time.Second, // Max time to write response
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}

	certfile, keyfile := s.InitCert()

	// Start the server in a background goroutine
	go func() {
		log.Printf("[API] Server HTTPS listening on %s", addr)
		// ErrServerClosed is expected when we call srv.Shutdown()
		if err := srv.ListenAndServeTLS(certfile, keyfile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] Server HTTPS critically failed: %v", err)
		}
	}()

	return srv

}

func (s *APIServer) InitCert() (string, string) {

	mainIP, err := network.GetPrimaryIP()
	if err != nil {
		log.Printf("failed to get primary ip, use localhost as CERT address.")
		mainIP = "localhost"
	}

	certfile, keyfile := CertKeyPathsForAddress(mainIP, "./")
	cert.EnsureExists(mainIP, certfile, keyfile)

	return certfile, keyfile

}

func CertKeyPathsForAddress(addr string, rootPath string) (string, string) {

	certfile := fmt.Sprintf("%s.cer", addr)
	cert := path.Join(rootPath, certfile)

	keyfile := fmt.Sprintf("%s.key", addr)
	key := path.Join(rootPath, keyfile)

	return cert, key

}