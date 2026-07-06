package apiserver

import (
	"net/http"
	"nvr_core/apiserver/middleware"
	"nvr_core/hardware"
	"nvr_core/network"
	"nvr_core/shm"
	"nvr_core/utils"
)

// HandleGetSHMMetrics aggregates the ring buffer stats from all running workers
func (api *APIServer) HandleGetSHMMetrics(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondErrForbidden(w)
		return
	}

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

func (api *APIServer) HandleGetNetworkInfo(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondErrForbidden(w)
		return
	}

	primaryIP, _ := network.GetPrimaryIP()
	allNICs := network.GetAvailableNICs()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	utils.RespondJSON(w, map[string]any{
		"primaryIP": primaryIP,
		"NICs": allNICs,
	}, "")
}

func (api *APIServer) HandleGetSystemMetrics(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondErrForbidden(w)
		return
	}

	nic := network.GetPrimaryNIC()
	allMetrics := hardware.NewTelemetryWatchdog(api.CFG.Server.StoragePath, nic)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	utils.RespondJSON(w, allMetrics, "")
}