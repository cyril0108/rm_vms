package apiserver

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data     any
}

func RespondJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(APIResponse{
		Data: data,
	})
}

