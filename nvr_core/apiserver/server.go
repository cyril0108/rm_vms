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

func Initiate(ctx context.Context, cfg *utils.Config, pm *process.Manager, svcs *service.Services) {

	log.Println("Initializing API server")

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

	// Debug Info
	mux.HandleFunc("GET /debug/db", api.GetDebugInfo)

	// =============================================
	// Login
	// =============================================
	mux.HandleFunc("POST /api/login", api.HandleLogin)

	// =============================================
	// User Management
	// =============================================
	mux.HandleFunc("PUT /api/users/{id}/permissions", api.HandleUpdateUserPermissions)


	// =============================================
	// Camera Discovery
	// =============================================
	mux.HandleFunc("GET /api/scan", api.HandleCameraScan)
	mux.HandleFunc("GET /api/scansweep", api.HandleCameraSweep)
	mux.HandleFunc("POST /api/scansweep/detail", api.HandleBulkONVIFScan)

	mux.HandleFunc("GET /api/scan/{ip}", api.HandleCameraProbe)
	mux.HandleFunc("POST /api/scan/{ip}/onvif", api.HandleFetchCameraONVIF)

	// =============================================
	// Camera stream
	// =============================================
	mux.HandleFunc("GET /ws/stream/{id}", api.GetStream)
	mux.HandleFunc("GET /live/camera/{id}", api.HandleLiveTransmuxTS)

	mux.HandleFunc("GET /health", api.GetHealth)
	mux.HandleFunc("GET /health/shm/metrics", api.HandleGetSHMMetrics)

	// =============================================
	// Camera Management
	// =============================================
	mux.HandleFunc("GET /api/cameras", api.GetCameras)
	mux.HandleFunc("GET /api/cameras/db", api.GetDBCameras)
	mux.HandleFunc("GET /api/cameras/{id}/onvif", api.HandleFetchSystemCameraONVIF)

	mux.HandleFunc("POST /api/cameras/add", api.AddCamera)
	mux.HandleFunc("POST /api/cameras/add/onvif", api.AddONVIFCamera)
	mux.HandleFunc("PUT /api/cameras/{cam_id}/update", api.UpdateCamera)
	mux.HandleFunc("PUT /api/cameras/{cam_id}/auth", api.UpdateCameraAuth)
	mux.HandleFunc("POST /api/cameras/{cam_id}/update", api.UpdateCamera)
	mux.HandleFunc("POST /api/cameras/{cam_id}/stop", api.DeactivateCamera)
	mux.HandleFunc("POST /api/cameras/{cam_id}/start", api.ActivateCamera)

	// Becareful with this
	mux.HandleFunc("DELETE /api/cameras/{cam_id}", api.DeleteCamera)


	// =============================================
	// Timeline and Playback
	// =============================================
	mux.HandleFunc("GET /api/cameras/{cam_id}/timeline/{start}/{end}", api.GetTimeline)
	mux.HandleFunc("GET /api/cameras/{cam_id}/play", api.HandlePlayVideo)
	mux.HandleFunc("GET /api/cameras/{cam_id}/play/ts", api.HandleTransmuxTS)

	// =============================================
	// Calendar
	// =============================================
	mux.HandleFunc("GET /api/cameras/{cam_id}/summary/{start}/{end}", api.HandleGetDailySummary)


	// =============================================
	// Playlist
	// =============================================
	mux.HandleFunc("GET /api/cameras/{cam_id}/playlist.m3u8", api.HandleGetPlaylist)
	mux.HandleFunc("GET /api/cameras/{cam_id}/playlist/ts.m3u8", api.HandleGetTSPlaylist)


	// =============================================
	// Serve Web
	// =============================================
	webserver.ServeWeb(mux, cfg)

	// =============================================
	// =============================================
	// =============================================
	addr := fmt.Sprintf(":%d", cfg.Server.Port);

	handlerWithCORS := middleware.CORSMiddleware(mux)

	// Server configuration
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlerWithCORS,
		ReadTimeout:  5 * time.Second,  // Max time to read request headers/body
		WriteTimeout: 10 * time.Second, // Max time to write response
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}

	// certfile, keyfile := api.InitCert()

	// Start the server in a background goroutine
	go func() {
		log.Printf("[API] Server listening on %s", addr)
		// ErrServerClosed is expected when we call srv.Shutdown()
		// if err := srv.ListenAndServeTLS(certfile, keyfile); err != nil && err != http.ErrServerClosed {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] Server critically failed: %v", err)
		}
	}()

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