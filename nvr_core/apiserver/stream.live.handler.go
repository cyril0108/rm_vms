package apiserver

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"nvr_core/stream"
	"nvr_core/utils"
)

// =====================================================================
//  HTTP HANDLER: Manages Request, Headers, and Hub Subscription
// =====================================================================

func (api *APIServer) HandleLiveTransmuxTS(w http.ResponseWriter, r *http.Request) {
	// --- HTTP/Connection Setup ---
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		log.Printf("[TS Handler] Warning: Failed to clear write deadline: %v", err)
	}

	camID, err := strconv.Atoi(r.PathValue("cam_id"))
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid camera ID", http.StatusBadRequest)
		return
	}

	profile := utils.SanitizeProfile(r.PathValue("profile"))

	worker, err := api.PM.CameraWorker(camID, profile)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "No worker for camera/profile", http.StatusNotFound)
		return
	}

	hub := worker.StreamHubForCam(camID, profile)
	if hub == nil {
		utils.RespondJSONHTTPStatus(w, "Stream not running", http.StatusNotFound)
		return
	}

	// Setup Endless HTTP Streaming Headers
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// --- Hub Subscription ---
	sub := &stream.Subscriber{
		Send:               make(chan stream.StreamPacket, 256),
		WaitingForKeyframe: true,
	}
	hub.Register <- sub
	defer func() {
		hub.Unregister <- sub
	}()

	// --- Stream Processing ---
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[TS Handler] Client disconnected from Cam %d", camID)
			return

		case packet, ok := <-sub.Send:
			if !ok {
				return // Hub channel closed
			}
			w.Write(packet.Payload)

		}
	}
}

