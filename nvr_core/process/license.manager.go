package process

import (
	"fmt"
	"sync"
	"time"

	"nvr_core/db/models"
)

var ll = LOG.Prefix("[license]")

// LicenseManager holds the active timers for expiring licenses
type LicenseManager struct {
	mu     sync.Mutex
	timers map[int64]*time.Timer // Map of License ID to its expiration timer
}

func NewLicenseManager() *LicenseManager {
	return &LicenseManager{
		timers: make(map[int64]*time.Timer),
	}
}

func (lm *LicenseManager) InitWithLicenses(list []*models.License) {

	for _, lic := range list {
		lm.ScheduleExpiration(lic)
	}

}

// ScheduleExpiration calculates exactly when the license expires and sets a sleeper timer
func (lm *LicenseManager) ScheduleExpiration(lic *models.License) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// If there's already a timer running for this license, stop it
	if existingTimer, exists := lm.timers[lic.ID]; exists {
		existingTimer.Stop()
	}

	// Perpetual licenses (no expiration)
	if lic.ExpiresAt == 0 {
		return
	}

	// Calculate the exact duration from NOW until the expiration timestamp
	expirationTime := time.Unix(lic.ExpiresAt, 0)
	timeUntilExpiration := time.Until(expirationTime)

	// If it's already in the past, trigger expiration immediately
	if timeUntilExpiration <= 0 {
		lm.triggerExpiration(lic.ID)
		return
	}

	// Schedule the exact moment to wake up and lock the system
	fmt.Printf("Scheduling expiration for license %d in %v\n", lic.ID, timeUntilExpiration)

	lm.timers[lic.ID] = time.AfterFunc(timeUntilExpiration, func() {
		lm.triggerExpiration(lic.ID)
	})
}

// triggerExpiration handles the actual teardown of features
func (lm *LicenseManager) triggerExpiration(licenseID int64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	fmt.Printf("ALERT: License %d has exactly expired!\n", licenseID)
	
	// Remove the timer from the map
	delete(lm.timers, licenseID)

	// TODO: Call your camera engine to recalculate allowed limits 
	// e.g., engine.RecalculateLimits()
}