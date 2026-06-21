package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SystemLicense represents a database record in the system_licenses table.
type License struct {
	ID         int64     `json:"id" db:"id"`
	RawToken   string    `json:"-" db:"raw_token"` // Hide from JSON responses to the frontend
	Iss        string    `json:"iss" db:"iss"`
	Aud        string    `json:"aud" db:"aud"`
	Kind       string    `json:"kind" db:"kind"`
	MachineID  string    `json:"machine_id" db:"machine_id"`
	MaxDevices int       `json:"max_devices" db:"max_devices"`
	IssuedAt   int64     `json:"iat" db:"issued_at"`
	ExpiresAt  int64     `json:"exp" db:"expires_at"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

func (lic *License) LoadClaims(claims *jwt.MapClaims) {
	if claims == nil {
		return
	}

	// Dereference the pointer to make map access cleaner
	c := *claims

	// Safely assert strings
	if iss, ok := c["iss"].(string); ok {
		lic.Iss = iss
	}
	if aud, ok := c["aud"].(string); ok {
		lic.Aud = aud
	}
	if kind, ok := c["kind"].(string); ok {
		lic.Kind = kind
	}
	if machineID, ok := c["machine_id"].(string); ok {
		lic.MachineID = machineID
	}

	// Safely assert numbers. 
	// JSON numbers are parsed as float64 by default in Go's interface{} maps!
	if maxDevices, ok := c["max_devices"].(float64); ok {
		lic.MaxDevices = int(maxDevices)
	}
	if iat, ok := c["iat"].(float64); ok {
		lic.IssuedAt = int64(iat)
	}
	if exp, ok := c["exp"].(float64); ok {
		lic.ExpiresAt = int64(exp)
	}
}