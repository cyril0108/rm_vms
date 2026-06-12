package apiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"nvr_core/logger"
)

var LOG = logger.NewLogger("[nvr_core][apiserver]")

type APIResponse struct {
	Data     any
}

func decodeRequest(r *http.Request, req *any) error {
	return json.NewDecoder(r.Body).Decode(&req)
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

// Get given alternative names from url query.
// msName is prefered when both exists.
// The returned int64 will be millisecond-base that is
// proper for database search.
func GetMSFromTime(r *http.Request, msName, sName string) (int64, error) {
	if mstimeStr := r.URL.Query().Get(msName); mstimeStr != "" {
		mstime, err := strconv.ParseInt(mstimeStr, 10, 64)
		if err != nil {
			return 0, errors.New("invalid " + msName + " format")
		}
		return mstime, nil
	}

	if timeStr := r.URL.Query().Get(sName); timeStr != "" {
		timeSec, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return 0, errors.New("invalid " + sName + " format")
		}
		// Shift one second so it should be within star/end time range
		// of sql search condition.
		return timeSec * 1000, nil
	}

	return -1, errors.New("missing " + sName + " parameter")
}

// Get search time from url query.
// The returned int64 will be millisecond-base that is
// proper for database search.
func GetSearchAtTime(r *http.Request) (int64, error) {
	return GetMSFromTime(r, "mstime", "time")
}

func GetDurationTime(r *http.Request) (int64, error) {
	return GetMSFromTime(r, "msduration", "duration")

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
