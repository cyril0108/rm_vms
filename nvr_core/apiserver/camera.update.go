package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"

	// "nvr_core/db/models"
	"nvr_core/onvif"
	// "nvr_core/process"
	"nvr_core/utils"
)


// UpdateCamera godoc
// @Summary      Add a new manual camera
// @Description  Creates a new IP camera in the NVR database using the provided manual data.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.UpdateCameraRequest  true  "Camera creation payload"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid JSON or missing fields"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/{cam_id}/update [put]
func (s *APIServer) UpdateCamera(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	var newCamera dto.UpdateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&newCamera); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}


	ctx := r.Context()
	cam := newCamera.ToMapInterface(s.CFG.Server.MasterKey())
	if err := s.Services.Camera.UpdateCamera(ctx, camID, cam); err != nil {
		errstr := fmt.Sprintf("Failed to update camera: %v", err)
		log.Print(errstr)
		http.Error(w, errstr, http.StatusBadRequest)
		return
	}

	// newCam.Status = "initializing"
	// s.State.Cameras.Store(cam.ID, cam)

	// TODO: Send camera start up command to the target C++ Worker Subprocess

	// theCamera := dto.MapCameraToDetail(*cam)

	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, "success"); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}


// UpdateCameraAuth godoc
// @Summary      Update camera credentials
// @Description  Check given credentials, and update if there is no issue.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.UpdateCameraRequest  true  "Camera creation payload"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid JSON or missing fields"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/cameras/{cam_id}/auth [put]
func (s *APIServer) UpdateCameraAuth(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	var newCamera dto.UpdateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&newCamera); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check authenticity of user/pwd combination
	// before save
	if cam, err := s.Services.Camera.GetByID(ctx, camID); err != nil {
		http.Error(w, "Failed to get camera data or it does not exist.", http.StatusBadRequest)
		return
	} else {

		good, err := onvif.VerifyCredentials(cam.IPAddress, *newCamera.Username, *newCamera.Password)
		if err != nil {
			errstr := fmt.Sprintf("Failed to check credentials: %v", err)
			http.Error(w, errstr, http.StatusBadRequest)
			return
		}
		if !good {
			http.Error(w, "Camera credentials was wrong", http.StatusForbidden)
			return
		}

	}

	// Cast only user/password for update
	cam := newCamera.ToUserPWDMapInterface(s.CFG.Server.MasterKey())
	if err := s.Services.Camera.UpdateCamera(ctx, camID, cam); err != nil {
		errstr := fmt.Sprintf("Failed to update camera: %v", err)
		log.Print(errstr)
		http.Error(w, errstr, http.StatusBadRequest)
		return
	}


	// TODO: Send camera start up command to the target C++ Worker Subprocess

	// theCamera := dto.MapCameraToDetail(*cam)


	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, "success"); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

// UpdateCameraONVIF godoc
// @Summary      Add a new ONVIF camera
// @Description  Creates a camera by dynamically fetching ONVIF profile tokens and streams using the provided IP and credentials.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.CreateCameraRequest  true  "ONVIF credentials and payload"
// @Success      201     {object}  dto.CameraDetailResponse
// @Failure      400     {string}  string "Invalid JSON payload"
// @Failure      500     {string}  string "Failed to get ONVIF data or internal server error"
// @Router       /api/cameras/{cam_id}/onvif [put]
func (s *APIServer) UpdateCameraONVIF(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	var camReq dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&camReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check authenticity of user/pwd combination
	// before save
	cam, err := s.Services.Camera.GetByID(ctx, camID)
	if err != nil {
		http.Error(w, "Failed to get camera data or it does not exist.", http.StatusBadRequest)
		return
	}

	user := cam.Username
	pwd, err := cam.DecryptPassword(s.CFG.Server.MasterKey())

	if camReq.Username != "" {
		user = camReq.Username
		pwd = camReq.Password
	}

	camOnvif, err := onvif.FetchCameraONVIFData(cam.IPAddress, user, pwd)
	if err != nil {

		// http.Error(w, "Failed to get ONVIF data", http.StatusInternalServerError)
		errstr := fmt.Sprintf("Failed to get ONVIF data: %v", err)
		log.Print(errstr)
		http.Error(w, errstr, http.StatusBadRequest)

		return
	}

	updateCam := dto.Onvif2UpdateCameraRequest(camOnvif)
	data := updateCam.ToMapInterface(s.CFG.Server.MasterKey())

	if err := s.Services.Camera.UpdateCamera(ctx, camID, data); err != nil {
		log.Printf("Failed to update camera: %v", err)
		http.Error(w, "Failed to update camera", http.StatusInternalServerError)
		return
	}

	// We need to restart camera worker
	if cam.StreamURL != *updateCam.StreamURL || cam.SubStreamURL != *updateCam.SubStreamURL {

	}

	w.WriteHeader(http.StatusCreated)
	if err := RespondJSON(w, updateCam); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

