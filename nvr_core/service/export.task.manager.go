package service

import (
	"sync"
	"time"

	"nvr_core/utils"
)


type ExportTask struct {
	ID          string     `json:"id"`
	Status      utils.TaskStatus `json:"status"`
	Progress    float64    `json:"progress"`
	MIME        string     `json:"-"`
	OutputPath  string     `json:"-"`
	ErrorMsg    string     `json:"error_msg,omitempty"` // Populated if failed
	CreatedAt   time.Time  `json:"created_at"`
}

type ExportTaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*ExportTask
}

func NewExportTaskManager() *ExportTaskManager {
	return &ExportTaskManager{
		tasks: make(map[string]*ExportTask),
	}
}

// CreateTask generates the ID and registers the initial state
func (m *ExportTaskManager) CreateTask() *ExportTask {
	task := &ExportTask{
		ID:        utils.GenerateTaskID(), // From step 1
		Status:    utils.TaskStatusPending,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.tasks[task.ID] = task
	m.mu.Unlock()

	return task
}

func (m *ExportTaskManager) UpdateTaskProgress(id string, progr float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, exists := m.tasks[id]; exists {
		task.Progress = progr

		// Give the frontend explicit feedback during the silent +faststart second pass
		if progr >= 99.0 && task.Status == utils.TaskStatusPending {
			task.Status = utils.TaskStatusFinalizing
		}
	}
}

func (m *ExportTaskManager) UpdateTaskStatus(id string, status utils.TaskStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, exists := m.tasks[id]; exists {
		task.Status = status
	}
}


func (m *ExportTaskManager) UpdateTaskSuccess(id, filePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, exists := m.tasks[id]; exists {
		task.Status = utils.TaskStatusCompleted
		task.OutputPath = filePath
	}
}

func (m *ExportTaskManager) UpdateTaskFailed(id, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, exists := m.tasks[id]; exists {
		task.Status = utils.TaskStatusFailed
		task.ErrorMsg = errorMsg
	}
}

func (m *ExportTaskManager) GetTask(id string) (*ExportTask, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, exists := m.tasks[id]
	return task, exists
}