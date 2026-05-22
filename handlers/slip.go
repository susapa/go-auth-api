package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	uploadDir     = "uploads/slips"
)

var allowedMIME = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}

func UploadSlip(c *gin.Context) {
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

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create upload directory"})
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitize(file.Filename), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "slip uploaded successfully",
		"filename": filename,
		"path":     dst,
		"size":     file.Size,
	})
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
