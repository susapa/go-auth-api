package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/user/go-auth-api/models"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("email already exists")

func CreateUser(db *sql.DB, name, email, hashedPassword string) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at`

	u := &models.User{}
	err := db.QueryRow(query, name, email, hashedPassword).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return u, nil
}

func FindByEmail(db *sql.DB, email string) (*models.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`
	u := &models.User{}
	err := db.QueryRow(query, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func FindByID(db *sql.DB, id int64) (*models.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	u := &models.User{}
	err := db.QueryRow(query, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func SaveRefreshToken(db *sql.DB, userID int64, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, userID, token, expiresAt)
	return err
}

func FindRefreshToken(db *sql.DB, token string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1`
	err := db.QueryRow(query, token).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token expired")
	}
	return userID, nil
}

func DeleteRefreshToken(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && contains(err.Error(), "unique constraint") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
