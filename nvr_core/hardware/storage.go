package hardware

import (
	"fmt"
	"syscall"
)

// DiskStatus holds the precise byte counts of the target mount.
type DiskStatus struct {
	TotalBytes     uint64 `json:"total_bytes"`
	UsedBytes      uint64 `json:"used_bytes"`
	FreeBytes      uint64 `json:"free_bytes"`
	AvailableBytes uint64 `json:"available_bytes"` // For non-root users
}

// GetDiskUsage queries the Linux VFS (Virtual File System) for mount statistics.
// path should be the absolute path to your recording mount point (e.g., "/storage").
func GetDiskUsage(path string) (DiskStatus, error) {
	var stat syscall.Statfs_t

	// syscall.Statfs directly queries the OS kernel for filesystem statistics.
	// It is extremely fast and operates in O(1) time.
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskStatus{}, fmt.Errorf("failed to read mount point %s: %w", path, err)
	}

	// Calculate sizes. 
	// Note: stat.Bsize represents the optimal transfer block size.
	// We cast to uint64 to prevent integer overflow on massive arrays (e.g., 10TB+ disks).

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return DiskStatus{
		TotalBytes:     total,
		UsedBytes:      used,
		FreeBytes:      free,
		AvailableBytes: available,
	}, nil
}