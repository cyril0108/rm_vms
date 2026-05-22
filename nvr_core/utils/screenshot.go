package utils

import (
	"errors"
	"fmt"
	"log"
	"nvr_core/db/models"
	"nvr_core/utils/storage"
	"os/exec"
)

const SnapshotProfile = "snapshot"

func GenerateSnapshot(storagePath string, seg *models.Segment) (string, error) {

	sPath := storage.NewStorePath(storagePath)

	snap, err := sPath.ForSnapshot(seg, SnapshotProfile)
	if err != nil {
		return "", err
	}

	result, err := ExtractSnapshot(seg.FilePath, snap)
	if err != nil {
		return "", err
	}

	return result, nil

}

// Extract snapshot for given video to given target path
func ExtractSnapshot(videoPath string, snapshotPath string) (string, error) {

	// snapshotPath := storage.NewStorePath()
	// Swap the .mkv extension for .jpg
	// ext := filepath.Ext(videoPath)
	// snapshotPath := strings.TrimSuffix(videoPath, ext) + ".jpg"

	log.Printf("[Snapshot] Extracting thumbnail for: %s\n", videoPath)

	// -y : Overwrite if exists
	// -i : Input file
	// -vframes 1 : Grab exactly 1 frame
	// -q:v 2 : High quality JPEG (range 1-31, lower is better)
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-vframes", "1", "-q:v", "2", snapshotPath)

	// Run the command. Because it is a separate OS process, a corrupted 
	// video file will just crash this single command, not your Go Manager.
	err := cmd.Run()
	if err != nil {
		errStr := fmt.Sprintf("[Snapshot] ERROR generating thumbnail for %s: %v\n", videoPath, err)
		log.Print(errStr)
		return "", errors.New(errStr)
	}

	log.Printf("[Snapshot] Successfully created: %s\n", snapshotPath)
	return snapshotPath, nil;

}