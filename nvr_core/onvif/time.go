package onvif

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"time"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
)

// CheckCameraTimeDrift anonymously fetches the camera's internal clock
// and compares it against the NVR's system time.
func CheckCameraTimeDrift(ip string) (float64, error) {
	address := fmt.Sprintf("%s:80", ip)

	// Initialize the device WITHOUT a username or password.
	// GetSystemDateAndTime does not require authentication.
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr: address,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to init device: %w", err)
	}

	// Request the time
	req := device.GetSystemDateAndTime{}
	resp, err := dev.CallMethod(req)
	if err != nil || resp == nil {
		return 0, fmt.Errorf("failed to get system time: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// 1. Isolate the UTCDateTime block. 
	// (Cameras usually return both Local and UTC, we strictly want UTC)
	utcBlockRe := regexp.MustCompile(`(?s)<(?:\w+:)?UTCDateTime>.*?</(?:\w+:)?UTCDateTime>`)
	utcBlock := utcBlockRe.Find(body)
	if len(utcBlock) == 0 {
		// Fallback just in case the camera doesn't separate them cleanly
		utcBlock = body
	}

	// 2. Extract the time fields
	year := extractInt(utcBlock, "Year")
	month := extractInt(utcBlock, "Month")
	day := extractInt(utcBlock, "Day")
	hour := extractInt(utcBlock, "Hour")
	minute := extractInt(utcBlock, "Minute")
	second := extractInt(utcBlock, "Second")

	if year == 0 {
		return 0, fmt.Errorf("could not parse date from ONVIF response")
	}

	// 3. Construct the Camera Time and get the NVR Time
	camTime := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	nvrTime := time.Now().UTC()

	// 4. Calculate the difference
	drift := nvrTime.Sub(camTime)
	driftSeconds := drift.Seconds()

	// 5. Print the diagnostic report
	fmt.Println("\n=== ONVIF TIME DRIFT DIAGNOSTIC ===")
	fmt.Printf("Camera IP:   %s\n", ip)
	fmt.Printf("Camera Time: %s (UTC)\n", camTime.Format(time.RFC3339))
	fmt.Printf("NVR Time:    %s (UTC)\n", nvrTime.Format(time.RFC3339))
	fmt.Printf("Time Drift:  %.0f seconds\n", driftSeconds)
	
	// The WS-Security standard usually enforces a strict 300-second (5 minute) tolerance
	if math.Abs(driftSeconds) > 300 {
		fmt.Println("STATUS:      CRITICAL DRIFT! (> 5 mins). WS-Security Auth will fail.")
	} else {
		fmt.Println("STATUS:      OK. Time is within acceptable WS-Security limits.")
	}
	fmt.Println("===================================")

	return driftSeconds, nil
}

// extractInt is a helper to parse numeric XML tags
func extractInt(xmlData []byte, tag string) int {
	re := regexp.MustCompile(`<(?:\w+:)?` + tag + `(?:[^>]*)>([^<]+)</(?:\w+:)?` + tag + `>`)
	match := re.FindSubmatch(xmlData)
	if len(match) > 1 {
		val, _ := strconv.Atoi(string(match[1]))
		return val
	}
	return 0
}