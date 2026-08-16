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
func (s *APIServer) getOnvifPTZManager(w http.ResponseWriter, r *http.Request) (*onvif.PTZPMController) {

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

	profile := GetQueryProfile(r)

	port, err := GetRequestPort(r)
	if err != nil {
		port = cam.HTTPPort
	}

	pm, err := onvif.NewPTZPMController(cam, profile, port, s.CFG.Server.MasterKey())
	if err != nil {
		LOG.Info("[getOnvifManager]", "err", err, "msg", err.Error())
		utils.RespondJSONHTTPStatus(w, "Failed to initialize ONVIF manager", http.StatusInternalServerError)
		return nil
	}

	return pm
}

// =============================================================================
// PTZ
// =============================================================================

func (s *APIServer) OMCameraPTZStatus(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	profile := GetQueryProfile(r)

	// ll := LOG.Prefix("[CameraPTZ]")

	om := s.getOnvifManager(w, r)
	if om == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	status, err := om.GetPTZStatus(ctx, profile)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, status, "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) OMCameraPTZContinuousMove(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	// ll := LOG.Prefix("[CameraPTZ]")

	pm := s.getOnvifPTZManager(w, r)
	if pm == nil {
		// getOnvifPTZManager will send respond, so here we just return
		return
	}

	ctx := r.Context()

	pan, tilt, zoom, _, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pm.MoveContinuous(ctx, pan, tilt, zoom)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) OMCameraPTZRelativeMove(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	// ll := LOG.Prefix("[CameraPTZ]")

	pm := s.getOnvifPTZManager(w, r)
	if pm == nil {
		// getOnvifPTZManager will send respond, so here we just return
		return
	}

	ctx := r.Context()

	pan, tilt, zoom, speed, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pm.MoveRelative(ctx, pan, tilt, zoom, speed)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) OMCameraPTZAbsoluteMove(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	// ll := LOG.Prefix("[CameraPTZ]")

	pm := s.getOnvifPTZManager(w, r)
	if pm == nil {
		// getOnvifPTZManager will send respond, so here we just return
		return
	}

	ctx := r.Context()

	pan, tilt, zoom, speed, err := getQueryPTZ(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = pm.MoveAbsolute(ctx, pan, tilt, zoom, speed)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}


func (s *APIServer) OMCameraPTZAbsoluteCanter(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	// ll := LOG.Prefix("[CameraPTZ]")

	pm := s.getOnvifPTZManager(w, r)
	if pm == nil {
		// getOnvifPTZManager will send respond, so here we just return
		return
	}

	ctx := r.Context()

	err := pm.MoveAbsoluteCenter(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}




func (s *APIServer) OMCameraPTZStep(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPTZ() {
		utils.RespondErrForbidden(w)
		return
	}

	pm := s.getOnvifPTZManager(w, r)
	if pm == nil {
		// getCameraPTZController will send respond, so here we just return
		return
	}

	ctx := r.Context()

	dir := r.URL.Query().Get("dir")

	var err error
	switch dir {
	case "up":
		err = pm.StepUp(ctx)
	case "down":
		err = pm.StepDown(ctx)
	case "left":
		err = pm.StepLeft(ctx)
	case "right":
		err = pm.StepRight(ctx)
	case "in":
		err = pm.StepZoomIn(ctx)
	case "out":
		err = pm.StepZoomOut(ctx)
	}

	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "", "success"); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

// func (s *APIServer) OMCameraPTZ(w http.ResponseWriter, r *http.Request) {

// 	session, ok := middleware.GetSession(r.Context())
// 	if !ok || !session.HasPermissionCameraPTZ() {
// 		utils.RespondErrForbidden(w)
// 		return
// 	}

// 	profile := GetQueryProfile(r)

// 	// ll := LOG.Prefix("[CameraPTZ]")

// 	om := s.getOnvifManager(w, r)
// 	if om == nil {
// 		// getCameraPTZController will send respond, so here we just return
// 		return
// 	}

// 	ctx := r.Context()

// 	status, err := om.GetPTZStatus(ctx, profile)
// 	if err != nil {
// 		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if err := utils.RespondJSON(w, status, "success"); err != nil {
// 		log.Printf("Error encoding response: %v", err)
// 	}

// }

// func (s *APIServer) OMCameraPTZ(w http.ResponseWriter, r *http.Request) {

// 	session, ok := middleware.GetSession(r.Context())
// 	if !ok || !session.HasPermissionCameraPTZ() {
// 		utils.RespondErrForbidden(w)
// 		return
// 	}

// 	profile := GetQueryProfile(r)

// 	// ll := LOG.Prefix("[CameraPTZ]")

// 	om := s.getOnvifManager(w, r)
// 	if om == nil {
// 		// getCameraPTZController will send respond, so here we just return
// 		return
// 	}

// 	ctx := r.Context()

// 	status, err := om.GetPTZStatus(ctx, profile)
// 	if err != nil {
// 		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if err := utils.RespondJSON(w, status, "success"); err != nil {
// 		log.Printf("Error encoding response: %v", err)
// 	}

// }

