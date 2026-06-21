package dto

import (
	"nvr_core/db/models"
	"nvr_core/security"

	"github.com/golang-jwt/jwt/v5"
)

type LicenseRequest struct {
	Licenses   []string `json:"licenses"`
}

type LicenseResponse struct {
	Token   string          `json:"token"`
	Claims  []jwt.MapClaims `json:"claims"`
	Error   string          `json:"error"`
}


type LicenseResult struct {
	Token   string          `json:"token"`
	Claims  *jwt.MapClaims `json:"claims"`
	Error   string          `json:"error"`
}

type LicenseProcessResult struct {
	Accepted []*LicenseResult `json:"accepted"`
	Rejected []*LicenseResult `json:"rejected"`
}

type LicenseStatus struct {
	Token      string    `json:"token"`
	models.License
	IsValid    bool      `json:"is_valid"`
	Errors     []string  `json:"errs"`
}

func (ls *LicenseStatus) LoadToken(token string, machine_id string) {

	isValid := true

	ls.Token = token
	claims, err := security.GetLicenseInfo(token)
	if err != nil {
		ls.Errors = append(ls.Errors, err.Error())
		isValid = false
	}

	if claims != nil {
		ls.LoadClaims(&claims)
	} else {
		ls.Errors = append(ls.Errors, "empty claims")
	}

	if ls.MachineID != machine_id {
		isValid = false
		ls.Errors = append(ls.Errors, "Machine id not match")
	}

	ls.IsValid = isValid

}