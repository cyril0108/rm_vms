package process

import (
	"context"
	"errors"
	"fmt"

	"nvr_core/logger"
	"nvr_core/service"
	"nvr_core/utils"
)

// Manager handles the pool of workers
type Manager struct {
	cfg        *utils.Config
	ctx        context.Context
	workers    []*Worker
	binaryPath string
	camWorker  map[int]int
	ingester   service.IngestService
	log        *logger.Logger
}

// NewManager initializes the pool (e.g., count=4)
func NewManager(ctx context.Context, cfg *utils.Config, count int, binaryPath string, ingester service.IngestService) *Manager {
	mgr := &Manager{
		ctx: ctx,
		cfg: cfg,
		workers:    make([]*Worker, count),
		binaryPath: binaryPath,
		camWorker: make(map[int]int),
		ingester:   ingester,
		log: LOG.Lin("sub","[manager]"),
	}

	// Initialize workers
	for i := 0; i < count; i++ {
		w := NewWorker(i, binaryPath, ingester)
		mgr.workers[i] = w
		w.SetStoragePath(cfg.Server.StoragePath)
	}

	return mgr
}

// StartAll launches all worker processes
func (m *Manager) StartAll() error {
	for _, w := range m.workers {
		// fmt.Printf("[Process Manager] Starting Worker %d...\n", w.ID)
		m.log.Info("Starting Worker", "worker", w.ID)
		if err := w.Start(m.ctx); err != nil {
			return err
		}
	}
	return nil
}

// StopAll shuts them down
func (m *Manager) StopAll() {
	m.log.Info("[StopAll]")
	for _, w := range m.workers {
		w.Stop()
	}
}

func (m *Manager) GetWorkers() []*Worker {
	return m.workers
}

// AssignCamera routes a camera to the correct worker (Sharding Logic)
func (m *Manager) AssignCamera(camID int, url string) error {
	if len(m.workers) == 0 {
		return fmt.Errorf("no workers available");
	}

	// SHARDING ALGORITHM: Round Robin using Modulus
	workerIndex := camID % len(m.workers);
	targetWorker := m.workers[workerIndex];

	m.camWorker[camID] = workerIndex;

	m.log.Info("[AssignCamera]", "cam", camID, "worker", workerIndex);

	return targetWorker.AssignCam(NewCamera(camID, url));
}

func (m *Manager) CameraWorker(camID int) *Worker {

	index := m.camWorker[camID];
	return m.workers[index];
}

func (m *Manager) StopCameraRecording(camID int) error {

	if worker := m.CameraWorker(camID); worker == nil {

		return errors.New("No worker for given camera")

	} else {

		worker.StopCam(camID);

	}

	return nil
}
