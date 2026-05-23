package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"nvr_core/logger"
)

var LOG = logger.NewLogger("[nvr_core]","[apiserver]")


type APIResponse struct {
	Data     any
}

func RespondJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(APIResponse{
		Data: data,
	})
}

func GetQueryProfile(r *http.Request) string {
	profile := r.URL.Query().Get("profile")
	if profile == "" {
		profile = "main"
	}
	return profile
}

func GetMSTimeRange(r *http.Request) (int64, int64, error) {

	// Parse timestamps
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, errStart := strconv.ParseInt(startStr, 10, 64)
	end, errEnd := strconv.ParseInt(endStr, 10, 64)

	if errStart != nil {
		return start, end, errStart
	}

	if errEnd != nil {
		return start, end, errEnd
	}

	LOG.Info("[GetMSTimeRange] seconds: ", "s", start, "e", end)

	start = start*1000
	end = end*1000

	LOG.Info("[GetMSTimeRange] ms: ", "s", start, "e", end)

	return  start, end, nil

}
