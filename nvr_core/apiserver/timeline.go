package apiserver

import (
	"errors"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/service"
	"nvr_core/utils"
	"strconv"
)

// GET /api/cameras/{cam_id}/timeline/segs?profile=main&start=123&end=123
func (api *APIServer) GetProfileSegments(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	profile := GetQueryProfile(r)

	start, end, err := GetMSTimeRange(r)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid start or end timestamps", http.StatusBadRequest)
		return
	}

	items, err := api.Services.Timeline.GetProfileSegmentItems(api.Context, camID, profile, start, end)

	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get profile segments", http.StatusInternalServerError)
		return
	}

	response := dto.TimelineResponse{
		CameraID: int(camID),
		Segments: items,
	}

	// Send the JSON payload
	if err := utils.RespondJSON(w, response, ""); err != nil {
		log.Printf("[API] Error encoding response: %v", err)
	}

}

// GET /api/cameras/{cam_id}/snapshot/range?start=123&end=123
func (api *APIServer) GetSegmentSnapshots(w http.ResponseWriter, r *http.Request) {

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

	items, err := api.Services.Timeline.GetCameraSnapshots(api.Context, camID, start, end)

	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get profile segments", http.StatusInternalServerError)
		return
	}

	response := dto.TimelineResponse{
		CameraID: int(camID),
		Snapshots: items,
	}

	// Send the JSON payload
	if err := utils.RespondJSON(w, response, ""); err != nil {
		log.Printf("[API] Error encoding response: %v", err)
	}

}

// GetTimeline godoc
// @Summary      
// @Description  GetTimeline queries the SQLite database for video segments within a time range. And returns continuous recording blocks.
// @Tags         Timeline
// @Produce      json
// @Param        cam_id  path      string  true  "Camera ID"
// @Param        start   path      string  true  "start time in seconds"
// @Param        end     path      string  true  "end time in seconds"
// @Success      201     {object}  dto.TimelineResponse
// @Failure      400     {string}  string "Invalid cam id"
// @Failure      400     {string}  string "Invalid start or end timestamps"
// @Failure      500     {string}  string "failed to get timeline blocks"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/{cam_id}/timeline/{start}/{end} [get]
func (api *APIServer) GetTimeline(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	// Parse Query Parameters
	startStr := r.PathValue("start")
	endStr := r.PathValue("end")

	start, err2 := strconv.ParseInt(startStr, 10, 64)
	end, err3 := strconv.ParseInt(endStr, 10, 64)

	if err2 != nil || err3 != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid query parameters. Requires cam_id, start, and end (Unix ms)", http.StatusBadRequest)
		return
	}

	// Convert to milliseconds
	start = start*1000
	end = end*1000

	svcs := api.Services
	blocks, errTimeline := svcs.Timeline.GetContiguousBlocks(api.Context, camID, start, end)

	if(errTimeline != nil) {
		log.Println("[GetTimeline] failed to get timeline blocks", errTimeline.Error())
		utils.RespondJSONHTTPStatus(w, "failed to get timeline blocks", http.StatusInternalServerError)
		return
	}

	response := dto.TimelineResponse{
		CameraID: int(camID),
		Timelines: blocks,
	}

	// Send the JSON payload
	if err := utils.RespondJSON(w, response, ""); err != nil {
		log.Printf("[API] Error encoding search response: %v", err)
	}
}

// HandleSegmentSnapshot expects: GET /api/cameras/{id}/snapshot?time=1711000050
func (api *APIServer) HandleSegmentSnapshot(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	// Shift one second so it should be within star/end time range
	// of sql search condition.
	timestamp, err := GetSearchAtTime(r)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the validated physical path from the Service
	filePath, err := api.Services.Playback.GetVideoSnapshotFilePath(r.Context(), camID, timestamp)
	if err != nil {
		if errors.Is(err, service.ErrVideoSegmentNotFound) || errors.Is(err, service.ErrFileMissing) {
			utils.RespondJSONHTTPStatus(w, "Snapshot not found", http.StatusNotFound)
			return
		}
		utils.RespondJSONHTTPStatus(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// http.ServeFile automatically reads the file from disk in chunks, 
	http.ServeFile(w, r, filePath)
}