package utils

import (
	"encoding/json"
	"errors"
	"net/http"
)


type APIResponse struct {
	Data     any     `json:"data"`
	Message  string  `json:"message"`
}

var (
	ErrorInvalidPayload = errors.New("Invalid Payload")
	ErrorFailedToEncodeResponse = errors.New("Failed to encode response")
)


// Source - https://stackoverflow.com/a/62734272
// Posted by kinshuk4
// Retrieved 2026-06-15, License - CC BY-SA 4.0
func RespondJSONHTTPStatus(w http.ResponseWriter, payload interface{}, code int) {

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.WriteHeader(code)

    if message, ok := payload.(string); ok {
    	json.NewEncoder(w).Encode(APIResponse{
    	    Message: message,
    	})
    } else 
    if err, ok := payload.(error); ok {
    	json.NewEncoder(w).Encode(APIResponse{
    	    Message: err.Error(),
    	})
    } else {
    	json.NewEncoder(w).Encode(APIResponse{
    	    Data: payload,
    	})
    }

}

func RespondJSON(w http.ResponseWriter, data any, message string) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(APIResponse{
		Data: data,
		Message: message,
	})
}

func RespondErrFailedToEncodeResponse(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorFailedToEncodeResponse, http.StatusInternalServerError)
}

func RespondErrInvalidPayload(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorInvalidPayload, http.StatusInternalServerError)
}
