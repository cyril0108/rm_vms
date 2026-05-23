package apiserver

import (
	"net/http"
	"nvr_core/utils"
	"strconv"
)

// HandleGetDailySummary expects: GET /api/cameras/{cam_id}/summary?start=1714521600&end=1717200000
func (s *APIServer) HandleGetDailySummary(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	startUnix, err1 := strconv.ParseInt(startStr, 10, 64)
	endUnix, err2 := strconv.ParseInt(endStr, 10, 64)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid start or end timestamps", http.StatusBadRequest)
		return
	}

	// Default to main as that we don't have sub profile now
	profile := "main"

	summaries, err := s.Services.Timeline.GetDailySummary(r.Context(), camID, profile, startUnix, endUnix)
	if err != nil {
		http.Error(w, "Failed to fetch summary", http.StatusInternalServerError)
		return
	}

	if err := RespondJSON(w, summaries); err!=nil {
		http.Error(w, "Failed to encode summary data", http.StatusInternalServerError)
		return
	}

}