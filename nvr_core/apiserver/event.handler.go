package apiserver

import (
	"net/http"
	"time"

	"nvr_core/db/models"
	"nvr_core/utils"
)

// HandleGetEvents retrieves all available user roles.
//
//	@Summary		Get Events from the given time
//	@Description	Retrieves a list of events.
//	@Tags			Events
//	@Produce		json
//	@Security		BearerAuth
//	@Param        time      query     int     true  "Unix timestamp (seconds)"
//	@Param        mstime    query     int     true  "Unix timestamp (milliseconds)"
//	@Success		200		{array}		models.Role	"List of roles"
//	@Failure		401		{string}	string		"Unauthorized"
//	@Failure		500		{string}	string		"Internal server error"
//	@Router			/api/events [get]
func (api *APIServer) HandleGetEvents(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	// Parse timestamps
	start, err := GetMSFromTime(r, "msstart", "start")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "need start arguments", http.StatusBadRequest)
		return
	}

	end, _ := GetMSFromTime(r, "msend", "end")

	var events []*models.Event

	if end == 0 {

		t := time.UnixMilli(start)
		events, err = api.Services.Event.GetEventsFrom(ctx, t)
		if err != nil {
			utils.RespondJSONHTTPStatus(w, "Failed to get events: "+err.Error(), http.StatusInternalServerError)
			return
		}

	} else {



	}


	if events == nil {
		events = []*models.Event{}
	}

	utils.RespondJSON(w, events, "")

}
