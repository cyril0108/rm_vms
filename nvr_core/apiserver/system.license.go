package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/hardware"
	"nvr_core/utils"
)



func (s *APIServer) HandleGetLicenseList(w http.ResponseWriter, r *http.Request) {

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

func (s *APIServer) HandleGetActiveLicenseStatus(w http.ResponseWriter, r *http.Request) {

	lics := s.PM.LicManage.Status
	max := s.PM.LicManage.MaxCamera()
	using := s.PM.ActiveCameraCount()

	utils.RespondJSON(w, dto.LicenseStatusResponse{
		Licenses: lics,
		MaxDevice: max,
		Using: using,
	}, "success")
	return

}


// func (s *APIServer) HandleGetAllLicenses(w http.ResponseWriter, r *http.Request) {

// 	list, err := s.Services.License.GetAllLicenses(r.Context())
// 	if err != nil {
// 		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	utils.RespondJSON(w, list, "success")
// 	return

// }


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

	ctx := r.Context()

	result, licenses := s.Services.License.ProcessLicenses(ctx, licReq.Licenses)

	if len(licenses) > 0 {

		for _, lic := range licenses {

			// license should have ID after successful creation.
			s.PM.LicManage.AddLicense(lic)

		}

	}

	utils.RespondJSON(w, result, "")
	return
}
