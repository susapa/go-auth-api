package models

import "time"

type SlipUpload struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}
