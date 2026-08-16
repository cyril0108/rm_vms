package apiserver

import (
	"log"
	"net/http"

	"nvr_core/apiserver/middleware"
	"nvr_core/onvif"
	"nvr_core/utils"
)

// Get data from request and return PTZ controller.
// nil means there is an error and is already dealt with
func (s *APIServer) getOnvifManager(w http.ResponseWriter, r *http.Request) (*onvif.ONVIFManager) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))

	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return nil
	}

	ctx := r.Context()
	cam, camErr := s.Services.Camera.GetByID(ctx, camID)
	if camErr != nil {

		utils.RespondJSONHTTPStatus(w, "No camera data with given ID", http.StatusNotFound)
		return nil

	}

	pwd, pwdErr := cam.DecryptPassword(s.CFG.Server.MasterKey())
	if pwdErr != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to decrypt camera password", http.StatusInternalServerError)
		return nil
	}

	port, err := GetRequestPort(r)
	if err != nil {
		port = cam.HTTPPort
	}

	cfg := onvif.DeviceConfig {
		IP: cam.IPAddress,
		Port: port,
		Username: cam.Username,
		Password: pwd,
	}

	om, err := onvif.NewONVIFManager(cfg, cam)
	if err != nil {
		LOG.Info("[getOnvifManager]", "err", err, "msg", err.Error())
		utils.RespondJSONHTTPStatus(w, "Failed to initialize ONVIF manager", http.StatusInternalServerError)
		return nil
	}

	return om
}

func (s *APIServer) OMCameraProfiles(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[OMCameraProfiles]")

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", om.Device.GetEndpoint("media"))

	profiles, err := om.GetProfiles(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, profiles, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) OMCameraProfile(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[OMCameraProfile]")

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", om.Device.GetEndpoint("media"))

	pf := GetQueryProfile(r)

	switch pf {
	case "":
		fallthrough
	case utils.SegmentMainProfile:
		pf = om.Camera.OnvifProfileToken

	case utils.SegmentSubProfile:
		pf = om.Camera.SubStreamProfileToken
	}
	// if pf == utils.SegmentMainProfile {
	// 	pf = om.Camera.OnvifProfileToken
	// } else if pf == utils.SegmentSubProfile {
	// 	pf = om.Camera.SubStreamProfileToken
	// }

	ll.Info("Query profile", "profile", pf)

	profile, err := om.GetProfile(ctx, pf)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, profile, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) OMCameraCapabilities(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[OMCameraCapabilities]")

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("")

	caps, err := om.GetCapabilities(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, caps, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) OMCameraDeviceInfo(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	info := om.GetDeviceInfo()

	if err := utils.RespondJSON(w, info, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) OMCameraVideoSourceConfig(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[OMCameraVideoSourceConfig]")

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	tkn := r.URL.Query().Get("token")

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", om.Device.GetEndpoint("media"))

	video, err := om.GetMainVideoSource(ctx)

	if tkn != "" {
		video, err = om.GetVideoSource(ctx, tkn)
	}

	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, video, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


// func (s *APIServer) OMCameraServices(w http.ResponseWriter, r *http.Request) {

// 	session, ok := middleware.GetSession(r.Context())
// 	if !ok || !session.HasPermissionCameraPTZ() {
// 		utils.RespondErrForbidden(w)
// 		return
// 	}

// 	ll := LOG.Prefix("[CameraServices]")

// 	om := s.getOnvifManager(w, r)
// 	if om == nil {
// 		// getCameraPTZController will send respond, so here we just return
// 		return
// 	}

// 	ctx := r.Context()

// 	ll.Info("Endpoint", "Endpoint", om)

// 	services, err := pc.Client.GetServices(ctx, true)
// 	if err != nil {
// 		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if err := utils.RespondJSON(w, services, "success"); err != nil {
// 		log.Printf("Error encoding response: %v", err)
// 	}

// }
