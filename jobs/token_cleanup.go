package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/user/go-auth-api/repository"
)

// StartTokenCleanup รัน goroutine ที่ลบ expired refresh tokens ทุก interval
// ใช้ context เพื่อให้หยุดได้สะอาดตอน server shutdown
func StartTokenCleanup(ctx context.Context, db *sql.DB, interval time.Duration) {
	go func() {
		log.Printf("[token-cleanup] started, interval=%s", interval)

		// รันรอบแรกทันทีตอน server boot
		runCleanup(db)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runCleanup(db)
			case <-ctx.Done():
				log.Println("[token-cleanup] stopped")
				return
			}
		}
	}()
}

func runCleanup(db *sql.DB) {
	n, err := repository.DeleteExpiredRefreshTokens(db)
	if err != nil {
		log.Printf("[token-cleanup] error: %v", err)
		return
	}
	if n > 0 {
		log.Printf("[token-cleanup] deleted %d expired tokens", n)
	}
}
