package dto

import "nvr_core/hardware"

/// === Request
type SystemSettingRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
}


/// === Response
type SystemDebugInfo struct {
	Status string `json:"status"`
	TotalSegments int `json:"total_segments"`
	DbSize float64 `json:"db_size_mb"`
	// WalModeActive bool `json:"wal_mode_active"`
	LastSegment SegmentItem `json:"last_segment"`
}

type SystemHealthInfo struct {
	Health     string `json:"health"`
	Configured bool   `json:"configured"`
	Version    string `json:"version"`
}

type SystemMachineInfo struct {
	MachineID     string `json:"machine_id"`
	ServerName    string `json:"server_name"`
	Version       string `json:"version"`
}


type SystemUsageMetrics struct {
	hardware.SystemMetrics
	PrimaryNIC             string `json:"primary_nic"`
}
