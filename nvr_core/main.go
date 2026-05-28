package main

import (
	"fmt"
	"log"
	"time"

	// Internal Packages
	"nvr_core/apiserver"
	"nvr_core/db"
	// "nvr_core/security"
	// "nvr_core/security/cert"
	// "nvr_core/webserver"

	// "nvr_core/db/repository"
	"nvr_core/process"
	"nvr_core/service"
	"nvr_core/utils"
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

	// Load Configuration
	fmt.Println("================================================")
	fmt.Println("[Go Manager] v.0.0.1")
	fmt.Println("[Go Manager] Loading config...")

	cfg, err := utils.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dbPath := cfg.Server.DBPath+"/nvr_metadata.db"

	log.Println("Attempting to open DB at:", dbPath)
	dbConn, err := db.InitiateDB(ctx, dbPath)

	if err != nil {
		log.Fatalf("Error Initiate database: %v", err)
	}


	fmt.Printf("[Go Manager] Config Loaded. Storage: %s, Port: %d \n", 
			cfg.Server.StoragePath, cfg.Server.Port)

	servs := service.NewServices(dbConn)
	ingester := service.StartIngester(ctx, dbConn)

	// Load cameras from db
	cams, err := servs.Camera.GetAll(ctx)
	if err != nil {
		log.Fatalf("Error loading cameras: %v", err)
	}

	pm := process.Startup(ctx, cfg, ingester, cams)

	go apiserver.Initiate(ctx, cfg, pm, servs)

	// go webserver.ServeWeb(cfg)

	// Block until the context is canceled (SIGINT/SIGTERM received)
	<-ctx.Done()

	fmt.Println("\n[Signal] Shutdown signal received. Terminating in 5 seconds...")

	// TODO:
	// At this point, the context is canceled. 
	// The C++ workers are automatically receiving kill signals.
	// We can add a brief time.Sleep() here or use a sync.WaitGroup 
	// to give everything a second to flush buffers before main() finally exits.

	time.Sleep(5*time.Second)

}

