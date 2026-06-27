package main

import (
	"time"

	// Internal Packages
	"nvr_core/apiserver"
	"nvr_core/db"
	"nvr_core/hardware"
	"nvr_core/license"
	"nvr_core/logger"
	"nvr_core/process"
	"nvr_core/service"
	"nvr_core/utils"
)

// Version Info with default values
// Should be set by GO_LDFLAGS
var (
    Version   = "dev"
    CommitSHA = "none"
    BuildTime = "unknown"
)

// @title           NVR Core API
// @version         0.1
// @description     API for NVR system.
// @host            localhost:9080
// @BasePath        /
func main() {

	// application-wide context
	ctx, cancel := utils.SetupSignalContext()
	defer cancel() // Ensures resources are freed when main exits

	ll := logger.NewLogger("")

	// Load Configuration
	ll.Info("=======================================================================")
	ll = ll.Prefix("[Go]")

	ll.Info("", "ver", Version, "commit", CommitSHA, "build_time", BuildTime)

	ll.Info("Loading config...")

	cfg, err := utils.LoadConfig("config.json")
	if err != nil {
		// log.Fatalf("Error loading config: %v", err)
		ll.Error("Error loading config", "error", err)
		return
	}

	dbPath := cfg.Server.DBPath+"/nvr_metadata.db"

	ll.Info("Attempting to open DB at:", "path", dbPath)
	dbConn, err := db.InitiateDB(ctx, dbPath)

	if err != nil {
		// log.Fatalf("Error Initiate database: %v", err)
		ll.Error("Error Initiate database", "error", err)
		return
	}


	ll.Info("Config Loaded", "Storage", cfg.Server.StoragePath, "Port", cfg.Server.Port)

	servs := service.NewServices(dbConn)
	ingester := service.StartIngester(ctx, dbConn)

	service.StartRetentionWatcher(ctx, dbConn, cfg.Server.StoragePath)


	//--------------------------
	// Boot Up Licenses
	//--------------------------
	machineID := hardware.GetPersistentMachineID()
	lm := license.NewLicenseManager()
	lics, err := servs.License.GetValidLicenses(ctx)
	if err == nil {
		lm.InitWithLicenses(lics, machineID)
	} else {
		lm.InitWithLicenses(nil, machineID)
		ll.Error("Error loading valid licenses", "error", err)
	}


	// Load cameras from db
	cams, err := servs.Camera.StartUpCameras(ctx, lm.MaxCamera())
	if err != nil {
		// log.Fatalf("Error loading cameras: %v", err)
		ll.Error("Error loading cameras", "error", err)
		return
	}

	pm := process.Startup(ctx, cfg, ingester, cams)
	pm.SetLicenseManager(lm)

	go apiserver.Initiate(ctx, cfg, pm, servs)

	// Block until the context is canceled (SIGINT/SIGTERM received)
	<-ctx.Done()

	// fmt.Println("\n[Signal] Shutdown signal received. Terminating in 5 seconds...")
	ll.Info("\n[Signal] Shutdown signal received. Terminating in 5 seconds...")

	// TODO:
	// At this point, the context is canceled. 
	// The C++ workers are automatically receiving kill signals.
	// We can add a brief time.Sleep() here or use a sync.WaitGroup 
	// to give everything a second to flush buffers before main() finally exits.

	time.Sleep(5*time.Second)

}

