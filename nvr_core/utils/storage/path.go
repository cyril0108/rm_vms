package storage

import (
	"fmt"
	"nvr_core/db/models"
	"os"
	"path/filepath"
	"time"
)

// StorePath represents the root directory for NVR storage
type StorePath struct {
	RootPath string
}

// NewStorePath initializes a new StorePath instance
func NewStorePath(root string) *StorePath {
	return &StorePath{RootPath: root}
}


// It replicates the ROOT/camID/profile/YYYY/MM/DD/HH-MM-SS.jpg structure.
func (s *StorePath) ForSnapshot(cam *models.Segment, profile string) (string, error) {




}



// For generates the directory and file path for a recording segment.
// It replicates the ROOT/camID/profile/ structure.
func (s *StorePath) MakeFolder(camID int, profile string, t time.Time) (string, error) {

	// Construct the directory path: ROOT/camID/YYYY/MM/DD
	camFolder := CamFolder(camID)
	dateFolder := DateFolder(t)

	folderPath := filepath.Join(s.RootPath, camFolder, profile, dateFolder)

	// Ensure the directory structure exists (equivalent to `mkdir -p`)
	// 0755 provides read/execute to group/others, full access to owner
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("[StorePath] Critical IO Error: Failed to create directories: %w", err)
	}

	return folderPath, nil

}


func SnapshotName(t time.Time) string {
	return fmt.Sprintf("%d.jpg", t.Unix())
}

func CamFolder(camID int) string {
	return fmt.Sprintf("cam%02d", camID)
}

func DateFolder(t time.Time) string {
	return t.Format("2006/01/02") // Go's reference date for formatting
}