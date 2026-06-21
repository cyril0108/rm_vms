package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/hardware"
	"nvr_core/security"
	"nvr_core/utils"
)



func (s *APIServer) HandleGetLicenseStatus(w http.ResponseWriter, r *http.Request) {

	list, err := s.Services.License.GetAllLicenses(r.Context())
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []*dto.LicenseStatus
	machine := hardware.GetPersistentMachineID()

	for _, lic := range list {
		LIC := &dto.LicenseStatus {}
		LIC.LoadToken(lic.RawToken, machine)
		result = append(result, LIC)
	}

	utils.RespondJSON(w, result, "success")
	return

}


func (s *APIServer) HandleGetAllLicenses(w http.ResponseWriter, r *http.Request) {

	list, err := s.Services.License.GetAllLicenses(r.Context())
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, list, "success")
	return

}


// HandleReceiveLicense
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
