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

// GetCameras godoc
// @Summary      Get active camera runtime list
// @Description  Retrieves a list of cameras currently loaded in memory by the C++ workers, including their live runtime status.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Success      200     {array}   process.Camera
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras [get]
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

// GetDBCameras godoc
// @Summary      Get database camera list
// @Description  Retrieves the complete list of saved cameras directly from the database configuration.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Success      200     {array}   models.Camera
// @Failure      500     {string}  string "Failed to get db cameras"
// @Router       /api/cameras/db [get]
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

// AddCamera godoc
// @Summary      Add a new manual camera
// @Description  Creates a new IP camera in the NVR database using the provided manual data.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.CreateCameraRequest  true  "Camera creation payload"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid JSON or missing fields"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/add [post]
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

	theCamera := dto.MapCameraToDetail(*newCamera.MapToDBCamera())

	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, theCamera); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

// DeleteCamera godoc
// @Summary      Delete camera
// @Description  Delete a camera by its ID. Note that a camera with existing recording data will be rejected.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        cam_id  path      string  true  "Camera ID"
// @Success      200     {string}  string  "deleted"
// @Failure      400     {string}  string  "Invalid cam id or development lock"
// @Failure      500     {string}  string  "Internal server error"
// @Router       /api/cameras/{cam_id} [delete]
func (s *APIServer) DeleteCamera(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	if(camID <= 3) {
		http.Error(w, "[dev] testing camera id should not be deleted.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.DeleteCamera(ctx, int64(camID)); err != nil {
		log.Printf("Failed to delete camera: %v", err)
		http.Error(w, "Failed to delete camera", http.StatusInternalServerError)
		return
	}

	if err := RespondJSON(w, "deleted"); err != nil {
		log.Printf("Error encoding delete camera response: %v", err)
	}

}


/*
JSON Payload {
	username: string
	password: string
}
*/
// AddONVIFCamera godoc
// @Summary      Add a new ONVIF camera
// @Description  Creates a camera by dynamically fetching ONVIF profile tokens and streams using the provided IP and credentials.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.CreateCameraRequest  true  "ONVIF credentials and payload"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid JSON payload"
// @Failure      500     {string}  string "Failed to get ONVIF data or internal server error"
// @Router       /api/cameras/onvif [post]
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

// ActivateCamera godoc
// @Summary      Activate camera
// @Description  Start camera recording and processing.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        cam_id  path      string  true  "Camera ID"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid cam id"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/{cam_id}/start [post]
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

// DeactivateCamera godoc
// @Summary      Deactivate camera
// @Description  Stop camera recording and processing.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        cam_id  path      string  true  "Camera ID"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid cam id"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/{cam_id}/stop [post]
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
