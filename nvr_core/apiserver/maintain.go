package apiserver

import (
	"net/http"
	"nvr_core/apiserver/middleware"
	"nvr_core/utils"
)

// GET /api/maintain/segments
func (api *APIServer) GetAbnormalSegments(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondErrForbidden(w)
		return
	}

	ctx := api.Context

	segs, err := api.Services.Maintain.GetAbnormalSegments(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "", http.StatusInternalServerError)
	}

	utils.RespondJSON(w, segs, "success")
	return

}


// PUT /api/maintain/segments
func (api *APIServer) FixAbnormalSegments(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondErrForbidden(w)
		return
	}

	ctx := api.Context

	segs, err := api.Services.Maintain.GetAbnormalSegments(ctx)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "", http.StatusInternalServerError)
	}

	fixed := api.Services.Maintain.FixSegmentsEndTime(ctx, segs)

	utils.RespondJSON(w, map[string]int{
		"fixed_segments": fixed,
	}, "success")
	return

}
