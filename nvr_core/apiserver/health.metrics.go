package apiserver

import (
	"net/http"
	"nvr_core/shm"
	"nvr_core/utils"
)

// HandleGetSHMMetrics aggregates the ring buffer stats from all running workers
func (api *APIServer) HandleGetSHMMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics := make([]*shm.WorkerMetrics, 0)

	// Iterate through your Process Manager's workers
	for _, worker := range api.PM.GetWorkers() {

		// Assuming you expose a getter for the shmReader on the Worker struct:
		if reader := worker.GetSHMReader(); reader != nil {
			allMetrics = append(allMetrics, reader.GetWorkerMetrics())
		}
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	utils.RespondJSON(w, allMetrics, "")
}