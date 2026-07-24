package process

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"nvr_core/db/models"
	"nvr_core/license"
	"nvr_core/logger"
	"nvr_core/service"
	"nvr_core/utils"
)

// Manager handles the pool of workers
type Manager struct {
	cfg           *utils.Config
	ctx           context.Context
	LicManage     *license.LicenseManager
	workers       []*Worker
	EstWorker     *EstimationWorker
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

	// The assignment logic will fail if 
	if count < 2 {
		count = 2
	}

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

	// Create EstimationWorker
	eWorker := NewWorker(99999999999999, binaryPath, nil)

	// Start without hooking IPC
	if err := eWorker.Start(ctx, false); err != nil {
		LOG.Error("[NewManager] failed to start EstimationWorker", "error", err)
	}
	mgr.EstWorker = NewEstimationWorker(eWorker)

	return mgr
}

func (m *Manager) SetLicenseManager(lm *license.LicenseManager) {
	m.LicManage = lm
}

func (m *Manager) AllCameras() []*Camera {
	return utils.CopyMapValues(m.cams, &m.mu)
}

func (m *Manager) ActiveCameraCount() int {
	cams := m.AllCameras()
	cnt := 0
	for _, cam := range cams {
		if cam.Active {
			cnt++
		}
	}
	return cnt
}

// Do not allow adding camera when existing number is larger then
// licensed number
func (m *Manager) CanAddNewCamera() bool {
	return len(m.cams) >= m.LicManage.MaxCamera()
}

func (m *Manager) ReachMaxLicenseNumber() bool {
	return m.ActiveCameraCount() >= m.LicManage.MaxCamera()
}

// StartAll launches all worker processes
func (m *Manager) StartAllWorkers() error {
	for _, w := range m.workers {
		// fmt.Printf("[Process Manager] Starting Worker %d...\n", w.ID)
		m.log.Info("Starting Worker", "worker", w.ID)
		if err := w.Start(m.ctx, true); err != nil {
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

func (m *Manager) GetCamera(camID int) *Camera {
	return m.cams[camID]
}

func (m *Manager) AssignNewCamera(newCam *models.Camera) (error, error) {
	cam := NewCameraRuntime(newCam, m.cfg.Server.MasterKey())
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

	wId, subId := decideWorkerAssignIDs(camID, len(m.workers));

	var err1, err2 error
	mainWorker := m.workers[wId];
	err1 = mainWorker.AssignCam(cam, utils.SegmentMainProfile)
	if err1 == nil {
		m.camMainWorker[camID] = wId;
	} else {
		wId = -1
	}

	subWorker := m.workers[subId];
	err2 = subWorker.AssignCam(cam, utils.SegmentSubProfile)
	if err2 == nil {
		m.camSubWorker[camID] = subId;
	} else {
		subId = -1
	}

	m.log.Info("[AssignCamera]", "cam", camID, "worker1", wId, "worker2", subId);

	return err1, err2
}

// Assign given ID to two worker IDs.
func decideWorkerAssignIDs(id int, len int) (int, int) {
	// SHARDING ALGORITHM: Round Robin using Modulus
	wId := id % len;
	subId := (wId + 1) % len;
	return wId, subId
}

// Get main/sub workers for given camera
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

func (m *Manager) RemoveCamera(camID int) {


	mw, err := m.CameraWorker(camID, utils.SegmentMainProfile)
	if err != nil {
		LOG.Error("[RemoveCamera] Failed to find main stream worker", "cam", camID, "error", err)
	} else {
		if mw != nil {
			LOG.Info("Removing camera main from worker", "cam", camID, "worker", mw.ID)
			mw.StopCam(camID, utils.SegmentMainProfile)
		}
	}

	sw, err := m.CameraWorker(camID, utils.SegmentSubProfile)
	if err != nil {
		LOG.Error("[RemoveCamera] Failed to find sub stream worker", "cam", camID, "error", err)
	} else {
		if sw != nil {
			LOG.Info("Removing camera sub from worker", "cam", camID, "worker", sw.ID)
			sw.StopCam(camID, utils.SegmentSubProfile)
		}
	}

	m.mu.Lock()
	delete(m.cams, camID)
	m.mu.Unlock()

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

func (m *Manager) StartCameraRecording(camID int) error {

    LOG.Info("[m][StartCameraRecording]")

	wId, subId := m.camWorkerIDs(camID)
	var err1, err2 error

    LOG.Info("[m][StartCameraRecording]", "w1", wId, "w2", subId)

	if wId >= 0 {
		mainWorker := m.workers[wId];
		err1 = mainWorker.StartCamRecording(camID, utils.SegmentMainProfile)
	}

	if subId >= 0 {
		subWorker := m.workers[subId];
		err2 = subWorker.StartCamRecording(camID, utils.SegmentSubProfile)
	}

	if cam := m.cams[camID]; cam != nil {
		cam.Active = true
	}

	return errors.Join(err1, err2)
}


func (m *Manager) StopCameraRecording(camID int) error {

	wId, subId := m.camWorkerIDs(camID)
	var err1, err2 error

	if wId >= 0 {
		mainWorker := m.workers[wId];
		err1 = mainWorker.StopCamRecording(camID, utils.SegmentMainProfile)
	}

	if subId >= 0 {
		subWorker := m.workers[subId];
		err2 = subWorker.StopCamRecording(camID, utils.SegmentSubProfile)
	}

	if cam := m.cams[camID]; cam != nil {
		cam.Active = false
	}

	return errors.Join(err1, err2)
}

