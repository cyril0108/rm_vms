package process

import (
	"context"

    "nvr_core/db/models"
	"nvr_core/service"
	"nvr_core/utils"
)

const CPP_WORKER_BIN = "./nvr_worker"

func Startup(ctx context.Context, cfg *utils.Config, ingester service.IngestService, cams []*models.Camera) (*Manager) {

	pm := NewManager(ctx, cfg, 4, CPP_WORKER_BIN, ingester)
	ll := LOG.Lin("fn", "Startup")

	if err := pm.StartAll(); err != nil {
		// log.Fatalf("Failed to start workers: %v", err)
		ll.Error("Failed to start workers", "err", err.Error())
	}

	// Distribute Cameras
	for _, cam := range cams {
		if err, err2 := pm.AssignCamera(NewCameraRuntime(cam)); err != nil || err2 != nil {
			// log.Printf("Failed to assign cam %d: %v", cam.ID, err)
			ll.Error("Failed to assign camera workers", "cam", cam.ID, "err", err, "err2", err2)
		}
	}

	return pm;

}
