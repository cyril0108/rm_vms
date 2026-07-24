package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)


type APIResponse struct {
	Data     any     `json:"data"`
	Message  string  `json:"message"`
}

var (
	ErrorForbidden = errors.New("Forbidden")
	ErrorInvalidPayload = errors.New("Invalid Payload")
	ErrorFailedToEncodeResponse = errors.New("Failed to encode response")
	ErrorReachMaxLicense = errors.New("Reached max license number")
)

// DisableHTTPTimeouts removes the read and write deadlines for the current HTTP response.
// This is strictly required for endpoints that stream large files (like video exports or DB backups)
// to prevent the global http.Server timeouts from killing the connection prematurely.
func DisableHTTPTimeouts(w http.ResponseWriter) error {
	rc := http.NewResponseController(w)

	// Passing a zero-value time.Time{} disables the timeout
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to disable write deadline: %w", err)
	}

	if err := rc.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to disable read deadline: %w", err)
	}

	return nil
}

// Source - https://stackoverflow.com/a/62734272
// Posted by kinshuk4
// Retrieved 2026-06-15, License - CC BY-SA 4.0
func RespondJSONHTTPStatus(w http.ResponseWriter, payload interface{}, code int) {

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.WriteHeader(code)

    if message, ok := payload.(string); ok {
    	json.NewEncoder(w).Encode(APIResponse{
    	    Data: "",
    	    Message: message,
    	})
    } else 
    if err, ok := payload.(error); ok {
    	json.NewEncoder(w).Encode(APIResponse{
    		Data: "",
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

func RespondErrForbidden(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorForbidden, http.StatusForbidden)
}

func RespondErrReachMaxLicense(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorReachMaxLicense, http.StatusBadRequest)
}

func RespondErrFailedToEncodeResponse(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorFailedToEncodeResponse, http.StatusInternalServerError)
}

func RespondErrInvalidPayload(w http.ResponseWriter) {
	RespondJSONHTTPStatus(w, ErrorInvalidPayload, http.StatusInternalServerError)
}
