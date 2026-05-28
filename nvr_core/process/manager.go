package process

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"nvr_core/db/models"
	"nvr_core/logger"
	"nvr_core/service"
	"nvr_core/utils"
)

// Manager handles the pool of workers
type Manager struct {
	cfg           *utils.Config
	ctx           context.Context
	workers       []*Worker
	binaryPath    string
	camMainWorker map[int]int
	camSubWorker  map[int]int
	cams          map[int]*Camera
	ingester      service.IngestService
	log           *logger.Logger
	mu            sync.Mutex
}

// NewManager initializes the pool (e.g., count=4)
func NewManager(ctx context.Context, cfg *utils.Config, count int, binaryPath string, ingester service.IngestService) *Manager {
	mgr := &Manager{
		ctx: ctx,
		cfg: cfg,
		workers:    make([]*Worker, count),
		binaryPath: binaryPath,
		camMainWorker: make(map[int]int),
		camSubWorker: make(map[int]int),
		cams: make(map[int]*Camera),
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

func (m *Manager) AllCameras() []*Camera {
	return utils.CopyMapValues(m.cams, &m.mu)
}

// StartAll launches all worker processes
func (m *Manager) StartAllWorkers() error {
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

func (m *Manager) AssignNewCamera(newCam *models.Camera) (error, error) {
	cam := NewCameraRuntime(newCam)
	return m.AssignCamera(cam)
}

// AssignCamera routes a camera to the correct worker (Sharding Logic)
func (m *Manager) AssignCamera(cam *Camera) (error, error) {
	if len(m.workers) == 0 {
		return fmt.Errorf("no workers available"), nil;
	}

	LOG.Info("[AssignCamera]", "main", cam.MainStream, "sub", cam.SubStream)

	camID := cam.ID
	m.cams[camID] = cam

	wId, subId := workerAssignIDs(camID, len(m.workers));

	mainWorker := m.workers[wId];
	m.camMainWorker[camID] = wId;

	subWorker := m.workers[subId];
	m.camSubWorker[camID] = subId;

	m.log.Info("[AssignCamera]", "cam", camID, "worker1", wId, "worker2", subId);

	err1 := mainWorker.AssignCam(cam, utils.SegmentMainProfile)
	err2 := subWorker.AssignCam(cam, utils.SegmentSubProfile)

	return err1, err2
}

// Assign given ID to two worker IDs.
func workerAssignIDs(id int, len int) (int, int) {
	// SHARDING ALGORITHM: Round Robin using Modulus
	wId := id % len;
	subId := (wId + 1) % len;
	return wId, subId
}

func (m *Manager) camWorkerIDs(camID int) (int, int) {

	wId, exists := m.camMainWorker[camID];
	subId, subExists := m.camSubWorker[camID];

	if !exists {
		wId = -1
	}

	if !subExists {
		subId = -1
	}

	return wId, subId
}

func (m *Manager) CameraWorker(camID int, profile string) (*Worker, error) {

	var index int
	switch profile {
	case "main":
		index = m.camMainWorker[camID];

	case "sub":
		index = m.camSubWorker[camID];

	default:
		return nil, fmt.Errorf("[Manager][CameraWorker] Unknown profile");
	}

	return m.workers[index], nil
}

func (m *Manager) StopCameraRecording(camID int) error {

	wId, subId := m.camWorkerIDs(camID)

	mainWorker := m.workers[wId];
	subWorker := m.workers[subId];

	err1 := mainWorker.StopCamRecording(camID, utils.SegmentMainProfile)
	err2 := subWorker.StopCamRecording(camID, utils.SegmentSubProfile)

	return errors.Join(err1, err2)
}

