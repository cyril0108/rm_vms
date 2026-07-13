package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"nvr_core/apiserver/dto"
	"nvr_core/hardware"
	// "nvr_core/security"
	"nvr_core/utils"
)

/// Send License Key
func (s *APIServer) HandleReceiveLicenseKey(w http.ResponseWriter, r *http.Request) {

	log := LOG.Prefix("[HandleReceiveLicenseKey]")

	var licReq dto.LicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&licReq); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if licReq.Key == "" {
		utils.RespondErrInvalidPayload(w)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	kl := log.Lin("lic_key", licReq.Key)

	res, err := s.PM.LicManage.API.GetLicense(ctx, licReq.Key)

	if err != nil {
		kl.Info("failed to get license data", "error", err)
		// Return bad gateway for now
		utils.RespondJSONHTTPStatus(w, "Failed to get license data", http.StatusBadGateway)
		return
	}

	if res.Code != 0 || res.Error != "" {
		kl.Info("License API respond with error", "code", res.Code, "error", res.Error)
		// Return bad gateway for now
		utils.RespondJSONHTTPStatus(w, res.Error, http.StatusBadGateway)
		return
	}

	jwt := res.Response.JWT

	// token, err := security.GetLicenseInfo(jwt)
	// if err != nil {

	// 	kl.Error("Failed to decode license data", "error", err)

	// } else {

	// 	kl.Info("license decode success", "claims", token)

	// }
	// utils.RespondJSON(w, token.Claims, "Test successful")


	ctx = r.Context()
	result, licenses := s.Services.License.ProcessLicenses(ctx, []string{ jwt })

	if len(licenses) > 0 {

		for _, lic := range licenses {

			// license should have ID after successful creation.
			s.PM.LicManage.AddLicense(lic)

		}

	}

	utils.RespondJSON(w, result, "")

}

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
		LIC.ID = lic.ID
		LIC.UploadedAt = lic.UploadedAt
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
