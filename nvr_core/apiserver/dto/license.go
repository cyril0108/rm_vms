package dto

import 	"github.com/golang-jwt/jwt/v5"

type LicenseRequest struct {
	Licenses   []string `json:"licenses"`
}

type LicenseResponse struct {
	Claims	[]jwt.MapClaims
}