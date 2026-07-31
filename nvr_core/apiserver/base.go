package apiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"nvr_core/logger"
)

var LOG = logger.NewLogger("[nvr_core][apiserver]")

const DefaultScanPort int = 80

func getPathID(r *http.Request, id string) (int64, error) {
	targetIDStr := r.PathValue(id)
	return strconv.ParseInt(targetIDStr, 10, 64)
}

func decodeRequest[T any](r *http.Request, req *T) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(req)
}

func getQueryInt(r *http.Request, name string) (int, error) {
	targetStr := r.URL.Query().Get(name)
	val, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func getQueryFloat(r *http.Request, name string, def float64) (float64, error) {
	targetStr := r.URL.Query().Get(name)
	val, err := strconv.ParseFloat(targetStr, 64)
	if err != nil {
		return def, err
	}
	return val, nil
}

func GetQueryProfile(r *http.Request) string {
	profile := r.URL.Query().Get("profile")
	if profile == "" {
		profile = "main"
	}
	return profile
}


// pan, tilt, zoom, speed
func getQueryPTZ(r *http.Request) (float64,float64,float64,float64, error) {

	pan, errP := getQueryFloat(r, "pan", 0)
	tilt, errT := getQueryFloat(r, "tilt", 0)
	zoom, errZ := getQueryFloat(r, "zoom", 0)
	speed, _ := getQueryFloat(r, "speed", 0.5)

	if errP != nil && errT != nil && errZ != nil {
		return 0,0,0,0, errors.New("Lack argument: pan, tilt, zoom. At least one should be present.")
	}

	return pan, tilt, zoom, speed, nil
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
	start, err := GetMSFromTime(r, "msstart", "start")
	if err != nil {
		return 0, 0, err
	}

	end, err := GetMSFromTime(r, "msend", "end")
	if err != nil {
		return 0, 0, err
	}

	LOG.Info("[GetMSTimeRange] ms: ", "s", start, "e", end)

	return  start, end, nil

}

func GetRequestPort(r *http.Request) (int, error) {

	// Parse timestamps
	p := r.URL.Query().Get("port")

	port, err := strconv.ParseInt(p, 10, 64)

	return  int(port), err

}

// isValidResolution strictly matches formats like "1280x720" or "3840x2160".
// It forces the width and height to be between 2 and 4 digits long, starting with a non-zero.
// This prevents command injection and stops bad actors from requesting absurdly huge resolutions.
var isValidResolution = regexp.MustCompile(`^[1-9][0-9]{1,3}x[1-9][0-9]{1,3}$`).MatchString
