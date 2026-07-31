package models

import "time"

type Bookmark struct {
	ID        int64     `json:"id"`
	CameraID  int64     `json:"camera_id"`
	UserID    int64     `json:"user_id"`
	Time      int64     `json:"time"` // Maps to 'time' or 'start_time' column
	Duration  int64     `json:"duration"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}