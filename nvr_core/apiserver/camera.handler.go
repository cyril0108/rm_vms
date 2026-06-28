package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"

	"nvr_core/onvif"
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

	camList := s.PM.AllCameras()

	if err := utils.RespondJSON(w, camList, ""); err != nil {
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
		utils.RespondJSONHTTPStatus(w, "failed to get db cameras.", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetCameras] camList(%d)\n", len(camList))

	if err := utils.RespondJSON(w, camList, ""); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

// HandleFetchSystemCameraONVIF godoc
// @Summary      Fetch ONVIF data
// @Description  Fetch ONVIF camera detail data that is already in our system
// @Tags         Cameras, Fetch
// @Accept       json
// @Produce      json
// @Success      201     {object}  onvif.OnvifRecord
// @Failure      500     {string}  string "Failed to get ONVIF data or internal server error"
// @Router       /api/camera/{id}/onvif [get]
func (s *APIServer) HandleFetchSystemCameraONVIF(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}


	ctx := r.Context()
	cam, camErr := s.Services.Camera.GetByID(ctx, camID)
	if camErr != nil {

		utils.RespondJSONHTTPStatus(w, "No camera data with given ID", http.StatusNotFound)
		return

	}

	pwd, pwdErr := cam.DecryptPassword(s.CFG.Server.MasterKey())
	if pwdErr != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to decrypt camera password", http.StatusInternalServerError)
		return
	}

	port, err := GetRequestPort(r)
	if err != nil {
		port = DefaultScanPort
	}

	result, err := onvif.FetchCameraONVIFData(cam.IPAddress, port, cam.Username, pwd)
	if result==nil && err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		result.ErrorMSG = err.Error()
	}

	if err := utils.RespondJSON(w, result, ""); err != nil {
		log.Printf("Error fetching camera ONVIF data: %v", err)
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
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

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraConfig() {
		utils.RespondErrForbidden(w)
		return
	}

	// if s.PM.CanAddNewCamera() {
	// 	utils.RespondErrReachMaxLicense(w)
	// 	return
	// }

	var newCamera dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&newCamera); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	ctx := r.Context()
	cam := newCamera.MapToDBCamera()
	cam.EncryptPassword(newCamera.Password, s.CFG.Server.MasterKey())
	if _, err := s.Services.Camera.AddCamera(ctx, cam); err != nil {
		errstr := fmt.Sprintf("Failed to add camera: %v", err)
		log.Print(errstr)
		utils.RespondJSONHTTPStatus(w, errstr, http.StatusBadRequest)
		return
	}

	//
	// s.PM.AssignNewCamera(cam)

	theCamera := dto.MapCameraToDetail(*cam)

	w.WriteHeader(http.StatusCreated)
	if err := utils.RespondJSON(w, theCamera, ""); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
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
// @Router       /api/cameras/add/onvif [post]
func (s *APIServer) AddONVIFCamera(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraConfig() {
		utils.RespondErrForbidden(w)
		return
	}

	// if s.PM.CanAddNewCamera() {
	// 	utils.RespondErrReachMaxLicense(w)
	// 	return
	// }

	var camReq dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&camReq); err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	port, err := GetRequestPort(r)
	if err != nil {
		port = DefaultScanPort
	}

// LOG.Debug("[AddONVIFCamera] receive", "u", camReq.Username, "p", camReq.Password)

	cam, camErr := onvif.FetchCameraONVIFData(camReq.IPAddress, port, camReq.Username, camReq.Password)
	if camErr != nil {

		// utils.RespondJSONHTTPStatus(w, "Failed to get ONVIF data", http.StatusInternalServerError)
		errstr := fmt.Sprintf("Failed to get ONVIF data: %v", camErr)
		log.Print(errstr)
		utils.RespondJSONHTTPStatus(w, errstr, http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	newCam := cam.MapToDBCamera()
	newCam.EncryptPassword(camReq.Password, s.CFG.Server.MasterKey())
	if _, err := s.Services.Camera.AddCamera(ctx, newCam); err != nil {
		log.Printf("Failed to add camera: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to add camera", http.StatusInternalServerError)
		return
	}

	// s.PM.AssignNewCamera(newCam)

	w.WriteHeader(http.StatusCreated)
	if err := utils.RespondJSON(w, dto.MapCameraToDetail(*newCam), ""); err != nil {
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

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraConfig() {
		utils.RespondErrForbidden(w)
		return
	}

	if s.PM.ReachMaxLicenseNumber() {
		utils.RespondErrReachMaxLicense(w)
		return
	}


	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.ActivateCamera(ctx, int64(camID)); err != nil {
		log.Printf("Failed to deactivate camera: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to stop camera recording", http.StatusInternalServerError)
		return
	}

	if err := s.PM.StartCameraRecording(int(camID)); err != nil {
		log.Printf("Failed to activate camera runtime err: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to start camera recording", http.StatusInternalServerError)
	}

	if err := utils.RespondJSON(w, "", "started"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

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

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraConfig() {
		utils.RespondErrForbidden(w)
		return
	}

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.DeactivateCamera(ctx, camID); err != nil {
		log.Printf("Failed to deactivate camera: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to stop camera recording", http.StatusInternalServerError)
		return
	}

	if err := s.PM.StopCameraRecording(int(camID)); err != nil {
		log.Printf("Failed to deactivate camera runtime err: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to stop camera recording", http.StatusInternalServerError)
	}

	if err := utils.RespondJSON(w, "", "stopped"); err != nil {
		log.Printf("Error encoding response: %v", err)
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

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraConfig() {
		utils.RespondErrForbidden(w)
		return
	}

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	// if(camID <= 3 && camID>0) {
	// 	utils.RespondJSONHTTPStatus(w, "[dev] testing camera id should not be deleted.", http.StatusBadRequest)
	// 	return
	// }

	ctx := r.Context()
	if err := s.Services.Camera.DeleteCamera(ctx, camID); err != nil {
		log.Printf("Failed to delete camera: %v", err)
		utils.RespondJSONHTTPStatus(w, "Failed to delete camera", http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "deleted"); err != nil {
		log.Printf("Error encoding delete camera response: %v", err)
	}

}

