package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/go-auth-api/db"
	"github.com/user/go-auth-api/models"
	"github.com/user/go-auth-api/repository"
	"github.com/user/go-auth-api/storage"
)

const maxUploadSize = 10 << 20 // 10 MB

var allowedMIME = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}

func UploadSlip(c *gin.Context) {
	claims := c.MustGet("claims").(*models.Claims)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if file.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file exceeds 10 MB limit"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	ext, ok := allowedMIME[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type: " + contentType})
		return
	}

	blobName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitize(file.Filename), ext)

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	blobURL, err := storage.Upload(c.Request.Context(), blobName, src, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	record, err := repository.SaveSlipUpload(db.Get(), claims.UserID, blobName, blobURL, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log upload"})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func GetSlipReport(c *gin.Context) {
	claims := c.MustGet("claims").(*models.Claims)

	uploads, err := repository.GetSlipUploadsByUser(db.Get(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch report"})
		return
	}

	c.JSON(http.StatusOK, uploads)
}

func sanitize(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	var b strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
