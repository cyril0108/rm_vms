package apiserver

import (
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/utils"
)

// GetUserLayouts
// @Router       /api/layout [get]
func (s *APIServer) GetUserLayouts(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	userID := session.UserID

	ctx := s.Context

	data, err := s.Services.Layout.GetUserLayouts(ctx, userID)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get user layouts.", http.StatusInternalServerError)
		return
	}

	list := make([]*dto.LayoutResult, len(data))
	for _, lo := range data {
		lr := &dto.LayoutResult{}
		lr.LoadFrom(lo)
		list = append(list, lr)
	}

	if err := utils.RespondJSON(w, list, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}

// @Router       /api/layout/{id} [get]
func (s *APIServer) GetLayout(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	id, err := getPathID(r, "id")
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	userID := session.UserID

	ctx := s.Context

	data, err := s.Services.Layout.GetLayout(ctx, userID, id)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get bookmark.", http.StatusInternalServerError)
		return
	}

	result := dto.LayoutResult{}
	result.LoadFrom(data)

	if err := utils.RespondJSON(w, result, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}

// @Router       /api/layout [post]
func (s *APIServer) HandleCreateLayout(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPlayback() {
		utils.RespondErrForbidden(w)
		return
	}

	var layoutReq dto.LayoutRequest
	err := decodeRequest(r, &layoutReq)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] payload error", "err", err)
		utils.RespondErrInvalidPayloadWError(w, err)
		return
	}

	userID := session.UserID

	ctx := s.Context
	model := layoutReq.AsModel(userID)

	LOG.Info("[HandleCreateBookmark] ", "model", model)

	data, err := s.Services.Layout.AddLayout(ctx, userID, model)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] error", "err", err)
		utils.RespondJSONHTTPStatus(w, "failed to add bookmark.", http.StatusInternalServerError)
		return
	}

	result := &dto.LayoutResult{}
	result.LoadFrom(data)

	if err := utils.RespondJSON(w, result, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}



// @Router       /api/layout [put]
func (s *APIServer) HandleUpdateLayout(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPlayback() {
		utils.RespondErrForbidden(w)
		return
	}

	var layoutReq dto.LayoutRequest
	err := decodeRequest(r, &layoutReq)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] payload error", "err", err)
		utils.RespondErrInvalidPayloadWError(w, err)
		return
	}

	userID := session.UserID

	ctx := s.Context
	model := layoutReq.AsModel(userID)

	LOG.Info("[HandleCreateBookmark] ", "model", model)

	err = s.Services.Layout.UpdateLayout(ctx, userID, model)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] error", "err", err)
		utils.RespondJSONHTTPStatus(w, "failed to add bookmark.", http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "updated", ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}


// @Router       /api/layout/{id} [delete]
func (s *APIServer) DeleteLayout(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	id, err := getPathID(r, "id")
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	userID := session.UserID

	ctx := s.Context

	err = s.Services.Layout.Delete(ctx, userID, id)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get bookmark.", http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, "deleted", "success"); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}