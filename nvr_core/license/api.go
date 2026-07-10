package license

import (
	licAPI "nvr_core/reqapi/license"
)

const LicenseAPIURL = "http://license.neo-mobi.com"

func NewLicenseAPI() *licAPI.Service {

	ll := LOG.Prefix("[NewLicenseAPI]")

	api, err := licAPI.NewService(LicenseAPIURL)
	if err != nil {
		ll.Info("Failed to initialize license API", "error", err)
	}

	return api

}