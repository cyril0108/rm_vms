package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/security"
	"nvr_core/utils"
)

// HandleAdminInitConfigure
//
func (s *APIServer) HandleReceiveLicense(w http.ResponseWriter, r *http.Request) {

	ll := LOG.Prefix("[HandleReceiveLicense]")

	var licReq dto.LicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&licReq); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	var res dto.LicenseResponse

	if len(licReq.Licenses) <= 0 {
		utils.RespondJSONHTTPStatus(w, res, http.StatusBadRequest)
		return
	}

	ll.Info("got Licenses", "count", len(licReq.Licenses))

	for _, lic := range licReq.Licenses {

		ll.Info("Checking license", "license", lic)

		claims, err := security.GetLicenseInfo(lic)
		if err == nil {
			res.Claims = append(res.Claims, claims)
		} else {
			ll.Error("Error when checking license info", "err", err)
		}

	}

	utils.RespondJSON(w, res, "success")
	return
}
