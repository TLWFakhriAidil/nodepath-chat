package rest

import (
	"os"
	"path/filepath"
	"strings"
	
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

// ServeMedia serves media files from the storage directory
func (handler *App) ServeMedia(c *fiber.Ctx) error {
	filename := c.Params("filename")
	
	// Security: Remove any path traversal attempts
	filename = filepath.Base(filename)
	
	// Check if file exists in storage
	mediaPath := filepath.Join(config.PathStorages, filename)
	
	if _, err := os.Stat(mediaPath); os.IsNotExist(err) {
		return c.Status(404).SendString("Media not found")
	}
	
	// Determine content type based on extension
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := "application/octet-stream"
	
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".mp4":
		contentType = "video/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".ogg":
		contentType = "audio/ogg"
	case ".pdf":
		contentType = "application/pdf"
	}
	
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	
	return c.SendFile(mediaPath)
}
