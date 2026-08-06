package apiserver

import (
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/db/models"
	"nvr_core/utils"
)

// GetCameraBookmarks
// @Router       /api/bookmark/{cam_id}?start=123&end=123 [get]
func (s *APIServer) GetCameraBookmarks(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPlayback() {
		utils.RespondErrForbidden(w)
		return
	}

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	start, end, err := GetMSTimeRange(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	ctx := s.Context

	data, err := s.Services.Bookmark.GetCameraBookmarksBetween(ctx, camID, start, end)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get bookmark.", http.StatusInternalServerError)
		return
	}

	list := utils.ConvertSlice(data, func(m *models.Bookmark) *dto.BookmarkResult {
		br := dto.BookmarkResult{}
		br.LoadFrom(m)
		return &br
	})

	if err := utils.RespondJSON(w, list, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}

// @Router       /api/bookmark?start=123&end=123 [get]
func (s *APIServer) GetBookmarks(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPlayback() {
		utils.RespondErrForbidden(w)
		return
	}

	start, end, err := GetMSTimeRange(r)
	if err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	ctx := s.Context

	data, err := s.Services.Bookmark.GetBookmarksBetween(ctx, start, end)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get bookmark.", http.StatusInternalServerError)
		return
	}

	list := utils.ConvertSlice(data, func(m *models.Bookmark) *dto.BookmarkResult {
		br := dto.BookmarkResult{}
		br.LoadFrom(m)
		return &br
	})

	if err := utils.RespondJSON(w, list, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}

// @Router       /api/bookmark [post]
func (s *APIServer) HandleCreateBookmark(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionCameraPlayback() {
		utils.RespondErrForbidden(w)
		return
	}

	var bookmarkReq dto.BookmarkRequest
	err := decodeRequest(r, &bookmarkReq)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] payload error", "err", err)
		utils.RespondErrInvalidPayloadWError(w, err)
		return
	}

	ctx := s.Context
	model := bookmarkReq.AsNewModel(session.UserID)

	LOG.Info("[HandleCreateBookmark] ", "model", model)

	data, err := s.Services.Bookmark.AddBookmark(ctx, model)
	if err != nil {
		LOG.Info("[HandleCreateBookmark] error", "err", err)
		utils.RespondJSONHTTPStatus(w, "failed to add bookmark.", http.StatusInternalServerError)
		return
	}

	result := &dto.BookmarkResult{}
	result.LoadFrom(data)

	if err := utils.RespondJSON(w, result, ""); err != nil {
		log.Printf("Failed to encode JSON format: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}
