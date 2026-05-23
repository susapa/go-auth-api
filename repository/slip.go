package repository

import (
	"database/sql"

	"github.com/user/go-auth-api/models"
)

func SaveSlipUpload(db *sql.DB, userID int64, filename, path string, size int64) (*models.SlipUpload, error) {
	query := `
		INSERT INTO slip_uploads (user_id, filename, path, size)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, filename, path, size, uploaded_at`

	s := &models.SlipUpload{}
	err := db.QueryRow(query, userID, filename, path, size).
		Scan(&s.ID, &s.UserID, &s.Filename, &s.Path, &s.Size, &s.UploadedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func GetSlipUploadsByUser(db *sql.DB, userID int64) ([]models.SlipUpload, error) {
	query := `
		SELECT id, user_id, filename, path, size, uploaded_at
		FROM slip_uploads
		WHERE user_id = $1
		ORDER BY uploaded_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []models.SlipUpload
	for rows.Next() {
		var s models.SlipUpload
		if err := rows.Scan(&s.ID, &s.UserID, &s.Filename, &s.Path, &s.Size, &s.UploadedAt); err != nil {
			return nil, err
		}
		uploads = append(uploads, s)
	}
	if uploads == nil {
		uploads = []models.SlipUpload{}
	}
	return uploads, rows.Err()
}
