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
func (s *APIServer) getCameraPTZController(w http.ResponseWriter, r *http.Request) (*onvif.PTZController) {

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

	address := onvif.ONVIFAddress(cam.IPAddress, port)
	pc, err := onvif.NewPTZController(address, cam.Username, pwd, cam.OnvifProfileToken)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to initialize ptz controller", http.StatusInternalServerError)
		return nil
	}

	return pc
}

func (s *APIServer) CameraProfile(w http.ResponseWriter, r *http.Request) {


	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraCapabilities]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	caps, err := pc.Client.GetCapabilities(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	probe, err := pc.ProbeCapabilities(ctx)
	if err != nil {
		ll.Info("failed to probe capabilities", "err", err.Error())
	}

	if err := utils.RespondJSON(w, map[string]any{
		"profile": pc.Profile,
		"caps": caps,
		"probe": probe,
	}, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraCapabilities(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraCapabilities]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	caps, err := pc.Client.GetCapabilities(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	probe, err := pc.ProbeCapabilities(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, map[string]any{
		"caps": caps,
		"probe": probe,
	}, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraServices(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraServices]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	services, err := pc.Client.GetServices(ctx, true)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, services, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZStatus(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	profile := GetQueryProfile(r)

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	status, err := pc.GetStatus(ctx, profile)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, status, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZContinuous(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZContinuous]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	pan, tilt, zoom, _, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pc.MoveContinuous(ctx, pan, tilt, zoom)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZStop(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZStop]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	err := pc.Stop(ctx, true, true)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZAbsolute(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZAbsolute]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	pan, tilt, zoom, speed, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pc.MoveAbsolute(ctx, pan, tilt, zoom, speed)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := pc.GetStatus(ctx, "")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "MoveAbsolute success sent, but failed to get status", http.StatusMultiStatus)
		return
	}

	if err := utils.RespondJSON(w, status, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZRelative(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZRelative]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	pan, tilt, zoom, speed, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pc.MoveRelative(ctx, pan, tilt, zoom, speed)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := pc.GetStatus(ctx, "")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "MoveRelative success sent, but failed to get status", http.StatusMultiStatus)
		return
	}

	if err := utils.RespondJSON(w, status, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) CameraConfigurations(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	data, err := pc.GetConfigurations(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, data, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZCampatibleConfigurations(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	data, err := pc.GetCompatiblePTZConfigurations(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, data, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZConfiguration(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	data, err := pc.GetConfiguration(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, data, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZPresets(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	data, err := pc.GetPresets(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, data, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) CameraPTZToHome(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	err := pc.GotoHomePosition(ctx, pc.Speed(1.0))
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func (s *APIServer) CameraPTZToPreset(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	ll := LOG.Prefix("[CameraPTZ]")

	token := r.URL.Query().Get("token")
	if token == "" {
		utils.RespondErrInvalidPayload(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}


	ctx := r.Context()

	ll.Info("Endpoint", "Endpoint", pc.Client.Endpoint())

	err := pc.GotoPreset(ctx, token, pc.Speed(1.0))
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


// HandlePTZStepLeft godoc
// @Summary      Send PTZ step left command to camera
// @Description  Creates a new IP camera in the NVR database using the provided manual data.
// @Tags         Cameras
// @Accept       json
// @Produce      json
// @Success      201     {object}
// @Failure      400     {string}  string "Invalid JSON or missing fields"
// @Failure      500     {string}  string "Internal server error"
// @Router       /api/ptz/{cam_id}/step/left [post]
func (s *APIServer) HandleCameraPTZStepLeft(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()
	err := pc.StepLeft(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *APIServer) HandleCameraPTZStepRight(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.StepRight(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *APIServer) HandleCameraPTZStepUp(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.StepUp(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *APIServer) HandleCameraPTZStepDown(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.StepDown(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *APIServer) HandleCameraPTZStepZoomIn(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.StepZoomIn(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *APIServer) HandleCameraPTZStepZoomOut(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.StepZoomOut(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}


func (s *APIServer) HandleCameraPTZCenter(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pc := s.getCameraPTZController(w, r)
	if pc == nil {
		return
	}

	ctx := r.Context()
	err := pc.MoveAbsoluteCenter(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

