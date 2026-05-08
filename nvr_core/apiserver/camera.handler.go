package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/onvif"
	"nvr_core/process"
	"nvr_core/utils"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetCameras(w http.ResponseWriter, r *http.Request) {
	var camList []*process.Camera

	workers := s.PM.GetWorkers()

	log.Printf("[GetCameras] workers(%d)\n", len(workers))

	for _, worker := range workers {

		cams := worker.GetCameras()
		log.Printf("[GetCameras] cams (%d)\n", len(cams))
		for _, cam := range cams {
			// cam.rtsp = ""
			camList = append(camList, cam)
		}

	}

	log.Printf("[GetCameras] camList(%d)\n", len(camList))

	if err := RespondJSON(w, camList); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

func (s *APIServer) GetDBCameras(w http.ResponseWriter, r *http.Request) {

	camList, err := s.Services.Camera.GetAll(r.Context())

	if(err != nil) {
		log.Printf("Error getting database camera list: %v", err)
		http.Error(w, "failed to get db cameras.", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetCameras] camList(%d)\n", len(camList))

	if err := RespondJSON(w, camList); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

/**
JSON Payload {
	name: string
	manufacturer: string
	model: string
	serial_number: string

	ip_address: string
	http_port: int
	type: string

	username: string
	password: string

	stream_url: string
	sub_stream_url: string

	retention_gb_limit: int
	is_active: bool
}
 */
// AddCamera stores the camera and start up
func (s *APIServer) AddCamera(w http.ResponseWriter, r *http.Request) {

	var newCamera dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&newCamera); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.AddCamera(ctx, newCamera.MapToDBCamera()); err != nil {
		log.Printf("Failed to add camera: %v", err)
		http.Error(w, "Failed to add camera", http.StatusInternalServerError)
		return
	}

	// newCam.Status = "initializing"
	// s.State.Cameras.Store(newCam.ID, newCam)

	// TODO: Send camera start up command to the target C++ Worker Subprocess


	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, newCamera); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

func (s *APIServer) DeleteCamera(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.DeleteCamera(ctx, int64(camID)); err != nil {
		log.Printf("Failed to delete camera: %v", err)
		http.Error(w, "Failed to delete camera", http.StatusInternalServerError)
		return
	}

}


/*
JSON Payload {
	username: string
	password: string
}
*/
// AddONVIFCamera stores the camera and start up
func (s *APIServer) AddONVIFCamera(w http.ResponseWriter, r *http.Request) {

	var camReq dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&camReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	cam, camErr := onvif.FetchCameraONVIFData(camReq.IPAddress, camReq.Username, camReq.Password)
	if camErr != nil {
		log.Printf("Failed to get ONVIF data: %v", camErr)
		http.Error(w, "Failed to get ONVIF data", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.AddCamera(ctx, cam.MapToDBCamera()); err != nil {
		log.Printf("Failed to add camera: %v", err)
		http.Error(w, "Failed to add camera", http.StatusInternalServerError)
		return
	}

	// newCam.Status = "initializing"
	// s.State.Cameras.Store(newCam.ID, newCam)

	// TODO: Send camera start up command to the target C++ Worker Subprocess


	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, cam); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

func (s *APIServer) ActivateCamera(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.ActivateCamera(ctx, int64(camID)); err != nil {
		log.Printf("Failed to deactivate camera: %v", err)
		http.Error(w, "Failed to stop camera recording", http.StatusInternalServerError)
		return
	}

	// TODO: Start camera recording

}

func (s *APIServer) DeactivateCamera(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.DeactivateCamera(ctx, int64(camID)); err != nil {
		log.Printf("Failed to deactivate camera: %v", err)
		http.Error(w, "Failed to stop camera recording", http.StatusInternalServerError)
		return
	}

	// TODO: Stop camera recording

}
