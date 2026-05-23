package apiserver

import (
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/utils"
	"strconv"
)

// SearchRecords queries the SQLite database for video segments within a time range
func (api *APIServer) GetTimeline(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	// Parse Query Parameters
	startStr := r.PathValue("start")
	endStr := r.PathValue("end")

	start, err2 := strconv.ParseInt(startStr, 10, 64)
	end, err3 := strconv.ParseInt(endStr, 10, 64)

	if err2 != nil || err3 != nil {
		http.Error(w, "Invalid query parameters. Requires cam_id, start, and end (Unix ms)", http.StatusBadRequest)
		return
	}

	// Convert to milliseconds
	start = start*1000
	end = end*1000

	svcs := api.Services
	blocks, errTimeline := svcs.Timeline.GetContiguousBlocks(api.Context, camID, start, end)

	if(errTimeline != nil) {
		log.Println("[GetTimeline] failed to get timeline blocks", errTimeline.Error())
		return
	}

	// hostAddr := r.Host // Dynamically grab the server IP/Port

	response := dto.TimelineResponse{
		CameraID: int(camID),
		Timelines: blocks,
	}

	// Send the JSON payload
	if err := RespondJSON(w, response); err != nil {
		log.Printf("[API] Error encoding search response: %v", err)
	}
}