package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/go-auth-api/config"
	"github.com/user/go-auth-api/db"
	"github.com/user/go-auth-api/handlers"
	"github.com/user/go-auth-api/jobs"
	"github.com/user/go-auth-api/middleware"
)

func main() {
	config.Load()
	db.Init(config.C.DatabaseURL)

	// context สำหรับ background jobs — cancel ตอน shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// รัน background job ล้าง expired tokens ทุก 24 ชั่วโมง
	jobs.StartTokenCleanup(ctx, db.Get(), 24*time.Hour)

	r := gin.Default()
	r.Use(middleware.CORS())

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/refresh", handlers.RefreshToken)
	}

	protected := r.Group("/auth")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/me", handlers.Me)
	}

	slips := r.Group("/slips")
	slips.Use(middleware.AuthRequired())
	{
		slips.POST("/upload", handlers.UploadSlip)
	}

	srv := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: r,
	}

	// รัน server ใน goroutine แยก เพื่อให้ main goroutine รอ signal
	go func() {
		log.Printf("server starting on :%s", config.C.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// รอ SIGINT (Ctrl+C) หรือ SIGTERM (docker stop, systemd)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	// cancel context → หยุด background jobs ทั้งหมด
	cancel()

	// รอให้ request ที่ค้างอยู่เสร็จ (max 10 วิ)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
