package repository

import (
	"database/sql"
	"log"
)

func DeleteExpiredRefreshTokens(db *sql.DB) (int64, error) {
	result, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

func DeleteAllRefreshTokensByUser(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func LogRefreshTokenStats(db *sql.DB) {
	var total, expired int
	db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&total)
	db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE expires_at < NOW()`).Scan(&expired)
	log.Printf("[token-cleanup] stats: total=%d expired=%d", total, expired)
}
