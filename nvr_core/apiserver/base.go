package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/logger"
)

// var ll = LOG.WithPrefix("sub/")
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

