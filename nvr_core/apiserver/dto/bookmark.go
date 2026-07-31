package dto

import (
	"nvr_core/db/models"
	"time"
)

type BookmarkRequest struct {
	ID        int64     `json:"id,omitempty"`
	CameraID  int64     `json:"camera_id"`
	Time      int64     `json:"time"` // As seconds
	Duration  int64     `json:"duration"` // As seconds
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type BookmarkResult struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CameraID  int64     `json:"camera_id"`
	Time      int64     `json:"time"` // As seconds
	Duration  int64     `json:"duration"` // As seconds
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ================================
// BookmarkRequest
// ================================

// Casting BookmarkRequest as a new Bookmark model for insert
func (br *BookmarkRequest) AsNewModel(userID int64) *models.Bookmark {

	return &models.Bookmark{
		UserID: userID,
		CameraID: br.CameraID,
		Time: br.Time * 1000,
		Duration: br.Duration * 1000,
		Message: br.Message,
	}

}

// ================================
// BookmarkResult
// ================================

// Load BookmarkResult from db model
func (br *BookmarkResult) LoadFrom(bm *models.Bookmark) {

	br.ID = bm.ID
	br.UserID = bm.UserID
	br.CameraID = bm.CameraID
	br.Time = bm.Time / 1000
	br.Duration = bm.Duration / 1000
	br.Message = bm.Message
	br.CreatedAt = bm.CreatedAt

}