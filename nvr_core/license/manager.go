package license

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
	licAPI "nvr_core/reqapi/license"
)

var ll = LOG.Prefix("[manager]")

// LicenseManager holds the active timers for expiring licenses
type LicenseManager struct {
	mu        sync.Mutex
	timers    map[int64]*time.Timer // Map of License ID to its expiration timer
	machineID string
	maxCamera int
	licStats  map[int64]*dto.LicenseStatus
	Status    []*dto.LicenseStatus

	API       *licAPI.Service
}

func NewLicenseManager() *LicenseManager {

	return &LicenseManager{
		timers: make(map[int64]*time.Timer),
		licStats: make(map[int64]*dto.LicenseStatus),
		API: NewLicenseAPI(),
	}

}

// ============================
// Initialize given licenses with machine id
// The licenses will be checked with validity.
func (lm *LicenseManager) InitWithLicenses(list []*models.License, mID string) {

	lll := ll.Lin("fn", "[InitWithLicenses]", "license_number", len(list))

	lm.mu.Lock()
	lm.machineID = mID
	lm.maxCamera = 0
	lm.mu.Unlock()

	for _, lic := range list {

		lm.AddLicense(lic)

	}

	lll.Info("done", "max", lm.maxCamera)

}

// Make sure
func (lm *LicenseManager) AddLicense(lic *models.License) {

	stats := &dto.LicenseStatus{}
	stats.LoadToken(lic.RawToken, lm.machineID)
	stats.ID = lic.ID
	stats.UploadedAt = lic.UploadedAt

	lm.mu.Lock()
	defer lm.mu.Unlock()

	if stats.IsValid {

		lm.maxCamera += lic.MaxDevices
		lm.licStats[lic.ID] = stats

		/// scheduleExpirationLocked should not lock,
		/// otherwise we will have deadlock situation
		lm.scheduleExpirationLocked(lic)

		ll.Info("[AddLicense] Add valid license", "number", lic.MaxDevices)

	} else {

		ll.Info("[AddLicense] invalid license", "license", lic.ID)

	}

	lm.Status = append(lm.Status, stats)

}

func (lm *LicenseManager) RemoveLicenseStatus(lic *dto.LicenseStatus) {

	lm.maxCamera -= lic.MaxDevices
	delete(lm.licStats, lic.ID)
	lm.Status = slices.DeleteFunc(lm.Status, func(l *dto.LicenseStatus) bool {
		return l.RawToken == lic.RawToken
	})

}

func (lm *LicenseManager) ScheduleExpiration(lic *models.License) {

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.scheduleExpirationLocked(lic)

}


// ScheduleExpiration calculates exactly when the license expires and sets a sleeper timer
func (lm *LicenseManager) scheduleExpirationLocked(lic *models.License) {

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

	fmt.Printf("ALERT: License %d has expired!\n", licenseID)

	lic, exists := lm.licStats[licenseID]
	if exists {
		lm.RemoveLicenseStatus(lic)
	}

	// Remove the timer from the map
	delete(lm.timers, licenseID)

}


//---------------------------------------------
// Getters
//---------------------------------------------
func (lm *LicenseManager) MaxCamera() int {
	return lm.maxCamera
}

func (lm *LicenseManager) MachineID() string {
	return lm.machineID
}
