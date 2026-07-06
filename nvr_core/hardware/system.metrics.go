package hardware

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// SystemMetrics represents the JSON payload sent to the Vue.js dashboard
type SystemMetrics struct {
	CPUUsagePercent   float64 `json:"cpu_usage_percent"`
	MemUsedMB         uint64  `json:"mem_used_mb"`
	MemTotalMB        uint64  `json:"mem_total_mb"`
	MemUsagePercent   float64 `json:"mem_usage_percent"`
	DiskFreeGB        float64 `json:"disk_free_gb"`
	DiskUsagePercent  float64 `json:"disk_usage_percent"`
	NetworkRxMbps     float64 `json:"network_rx_mbps"` // Megabits per second
	NetworkTxMbps     float64 `json:"network_tx_mbps"`
}

type TelemetryWatchdog struct {
	mu             sync.RWMutex
	currentMetrics SystemMetrics
	recordPath     string // e.g., "/mnt/nvr_storage"
	targetNIC      string // e.g., "eth0"

	// State for network delta calculation
	lastNetStat *net.IOCountersStat
	lastNetTime time.Time
}

func NewTelemetryWatchdog(recordPath, targetNIC string) *TelemetryWatchdog {
	return &TelemetryWatchdog{
		recordPath: recordPath,
		targetNIC:  targetNIC,
	}
}

// Start begins the background polling. Run this once on NVR boot.
func (tw *TelemetryWatchdog) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			tw.pollSystem()
		}
	}()
}

// GetMetrics is what your HTTP/WebSocket handler calls. It returns instantly.
func (tw *TelemetryWatchdog) GetMetrics() SystemMetrics {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	return tw.currentMetrics
}

func (tw *TelemetryWatchdog) pollSystem() {
	metrics := tw.PollSystemMetrics()

	// Safely update the cached struct
	tw.mu.Lock()
	tw.currentMetrics = metrics
	tw.mu.Unlock()
}

func (tw *TelemetryWatchdog) PollSystemMetrics() SystemMetrics {
	metrics := SystemMetrics{}

	// CPU (Non-blocking because we pass 0, it uses the delta since the last call)
	// Setting `false` means overall CPU. Setting `true` gives per-core stats.
	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		metrics.CPUUsagePercent = cpuPercents[0]
	}

	// Memory
	vMem, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemTotalMB = vMem.Total / 1024 / 1024
		metrics.MemUsedMB = vMem.Used / 1024 / 1024
		metrics.MemUsagePercent = vMem.UsedPercent
	}

	// Disk (Point this strictly to your video recording drive, not necessarily root /)
	dStat, err := disk.Usage(tw.recordPath)
	if err == nil {
		metrics.DiskFreeGB = float64(dStat.Free) / 1024 / 1024 / 1024
		metrics.DiskUsagePercent = dStat.UsedPercent
	}

	// Network (We must calculate the delta bytes over the interval time)
	netStats, err := net.IOCounters(true)
	if err == nil {
		for _, stat := range netStats {
			if stat.Name == tw.targetNIC { // Only monitor the physical NIC, ignore loopback/docker
				now := time.Now()
				if tw.lastNetStat != nil {
					durationSec := now.Sub(tw.lastNetTime).Seconds()

					// Calculate Mbps (Megabits per second)
					rxBytes := stat.BytesRecv - tw.lastNetStat.BytesRecv
					txBytes := stat.BytesSent - tw.lastNetStat.BytesSent

					metrics.NetworkRxMbps = (float64(rxBytes) * 8) / 1000000 / durationSec
					metrics.NetworkTxMbps = (float64(txBytes) * 8) / 1000000 / durationSec
				}
				tw.lastNetStat = &stat
				tw.lastNetTime = now
				break
			}
		}
	}

	return metrics
}