package apiserver

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"nvr_core/apiserver/middleware"
	"nvr_core/service"
	"nvr_core/utils"
)

// HandleExportRequest godoc
// @Summary      Initiate a video export task
// @Description  Asynchronously exports video for a specific camera and time range. Returns a task ID for status polling and downloading.
// @Tags         Export
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cam_id   path      int     true  "Camera ID"
// @Param        start    query     int     true  "Start Unix timestamp (milliseconds)"
// @Param        end      query     int     true  "End Unix timestamp (milliseconds)"
// @Param        profile  query     string  false "Stream profile (e.g., main, sub)"
// @Success      200      {object}  map[string]string "Returns task_id and status (pending)"
// @Failure      400      {object}  map[string]string "Invalid camera ID or timestamps"
// @Failure      403      {string}  string            "Forbidden - Missing export permissions"
// @Router       /api/export/{cam_id} [get]
// GET /api/export/{cam_id}?profile=main&start=123&end=123
func (api *APIServer) HandleExportRequest(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionRecordExport() { // Strictly enforce permissions
		utils.RespondErrForbidden(w)
		return
	}

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	start, end, err := GetMSTimeRange(r)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid start or end timestamps", http.StatusBadRequest)
		return
	}

	profile := GetQueryProfile(r)

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "mp4"
	}


	TM := api.Services.ExportTM

	// Generate ID and register the task synchronously
	task := TM.CreateTask()
	task.MIME = utils.GetVideoMimeType(format)

	// Spin up the background worker
	go func(taskID string) {
		// Update status to running
		// (You could add an UpdateTaskRunning method to the manager)

		rootPath := api.CFG.Server.StoragePath

		// Run the FFmpeg logic we built earlier
		outputPath, err := api.Services.Export.ExportTimeRange(context.Background(), rootPath, service.ExportParams{
			TM: api.Services.ExportTM,
			TaskID: taskID,
			CamID: camID,
			Profile: profile,
			Format: format,
			Start: start,
			End: end,
		})

		if err != nil {
			TM.UpdateTaskFailed(taskID, err.Error())
			return
		}

		// Convert physical path (/var/nvr/exports/...) to a downloadable URL path (/api/downloads/...)
		// downloadUrl := "/api/downloads/" + filepath.Base(outputPath)

		TM.UpdateTaskSuccess(taskID, outputPath)

	}(task.ID)


	utils.RespondJSON(w,map[string]string{
		"task_id": task.ID,
		"status":  "pending",
	}, "success")

}

// HandleExportTaskStatus godoc
// @Summary      Get export task status
// @Description  Get the status and progress of the exporting task.
// @Tags         Export
// @Security     BearerAuth
// @Param        task_id  path      string  true  "Export Task ID (UUID)"
// @Success      200      {file}    file    "The exported video file"
// @Failure      400      {string}  string  "Missing task ID"
// @Failure      403      {string}  string  "Forbidden - Missing export permissions"
// @Failure      404      {string}  string  "Export task not found"
// @Failure      410      {string}  string  "Exported file has expired or been deleted"
// @Router       /api/export/{task_id}/status [get]
func (api *APIServer) HandleExportTaskStatus(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionRecordExport() { // Strictly enforce permissions
		utils.RespondErrForbidden(w)
		return
	}

	// Extract the TaskID from the URL parameters
	taskID := r.PathValue("task_id")

	if taskID == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	// Look up the task securely in server memory
	task, exists := api.Services.ExportTM.GetTask(taskID)
	if !exists {
		http.Error(w, "Export task not found", http.StatusNotFound)
		return
	}

	utils.RespondJSON(w, task, string(task.Status))

}


// HandleDownloadExport godoc
// @Summary      Download exported video
// @Description  Downloads the compiled .mp4 file for a completed export task.
// @Tags         Export
// @Security     BearerAuth
// @Param        task_id  path      string  true  "Export Task ID (UUID)"
// @Success      200      {file}    file    "The exported video file"
// @Failure      400      {string}  string  "Missing task ID"
// @Failure      403      {string}  string  "Forbidden - Missing export permissions"
// @Failure      404      {string}  string  "Export task not found"
// @Failure      409      {string}  string  "Export is not ready for download"
// @Failure      410      {string}  string  "Exported file has expired or been deleted"
// @Router       /api/export/{task_id}/download [get]
func (api *APIServer) HandleDownloadExport(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionRecordExport() { // Strictly enforce permissions
		utils.RespondErrForbidden(w)
		return
	}


	// Extract the TaskID from the URL parameters
	taskID := r.PathValue("task_id")

	if taskID == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	// Look up the task securely in server memory
	task, exists := api.Services.ExportTM.GetTask(taskID)
	if !exists {
		http.Error(w, "Export task not found", http.StatusNotFound)
		return
	}

	if task.Status != utils.TaskStatusCompleted {
		http.Error(w, "Export is not ready for download", http.StatusConflict)
		return
	}

	// Verify the file actually exists on disk (defense in depth)
	if _, err := os.Stat(task.OutputPath); os.IsNotExist(err) {
		http.Error(w, "Exported file has expired or been deleted", http.StatusGone)
		return
	}

	utils.DisableHTTPTimeouts(w)

	// Force the browser to download the file rather than play it in-browser
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(task.OutputPath))
	w.Header().Set("Content-Type", task.MIME)

	// Safely serve the file.
	// http.ServeFile automatically handles range requests (for pausing/resuming downloads)
	http.ServeFile(w, r, task.OutputPath)
}