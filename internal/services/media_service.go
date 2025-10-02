package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
)

// MediaService handles media file operations with CDN integration and optimization
type MediaService struct {
	cdnEnabled    bool
	cdnBaseURL    string
	localBasePath string
	maxFileSize   int64 // Maximum file size in bytes
	allowedTypes  map[string]bool

	// File cache for frequently accessed files
	fileCache map[string]*CachedFile
	cacheMux  sync.RWMutex
	cacheTTL  time.Duration
}

// CachedFile represents a cached file
type CachedFile struct {
	Data      []byte
	MimeType  string
	Timestamp time.Time
}

// MediaUploadResult represents the result of a media upload
type MediaUploadResult struct {
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	URL          string `json:"url"`
	CDNURL       string `json:"cdn_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Compressed   bool   `json:"compressed"`
}

// NewMediaService creates a new media service with performance optimizations
func NewMediaService(cdnEnabled bool, cdnBaseURL, localBasePath string) *MediaService {
	// Create local directory if it doesn't exist
	os.MkdirAll(localBasePath, 0755)

	return &MediaService{
		cdnEnabled:    cdnEnabled,
		cdnBaseURL:    cdnBaseURL,
		localBasePath: localBasePath,
		maxFileSize:   10 * 1024 * 1024, // 10MB default
		allowedTypes: map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"image/gif":       true,
			"image/webp":      true,
			"audio/mpeg":      true,
			"audio/ogg":       true,
			"video/mp4":       true,
			"video/webm":      true,
			"application/pdf": true,
			"text/plain":      true,
		},
		fileCache: make(map[string]*CachedFile),
		cacheTTL:  30 * time.Minute,
	}
}

// UploadFile handles file upload with optimization and CDN integration
func (ms *MediaService) UploadFile(fileHeader *multipart.FileHeader) (*MediaUploadResult, error) {
	// Validate file size
	if fileHeader.Size > ms.maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", ms.maxFileSize)
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer file.Close()

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(fileData)

	// Validate file type
	if !ms.allowedTypes[mimeType] {
		return nil, fmt.Errorf("file type %s is not allowed", mimeType)
	}

	// Generate unique filename
	fileName := ms.generateFileName(fileHeader.Filename, fileData)
	filePath := filepath.Join(ms.localBasePath, fileName)

	// Optimize file if it's an image
	optimizedData := fileData
	compressed := false
	if strings.HasPrefix(mimeType, "image/") {
		if optimized, err := ms.optimizeImage(fileData, mimeType); err == nil {
			optimizedData = optimized
			compressed = true
			logrus.Debug("Image optimized successfully")
		}
	}

	// Save file locally
	err = os.WriteFile(filePath, optimizedData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// Generate URLs
	localURL := fmt.Sprintf("/media/%s", fileName)
	cdnURL := ""
	if ms.cdnEnabled && ms.cdnBaseURL != "" {
		cdnURL = fmt.Sprintf("%s/%s", strings.TrimRight(ms.cdnBaseURL, "/"), fileName)
	}

	// Console log for tracing URL generation
	logrus.WithFields(logrus.Fields{
		"file_name":    fileName,
		"local_url":    localURL,
		"cdn_url":      cdnURL,
		"cdn_enabled":  ms.cdnEnabled,
		"cdn_base_url": ms.cdnBaseURL,
	}).Info("🔍 MEDIA SERVICE: URLs GENERATED FOR TRACING")

	// Generate thumbnail for images
	thumbnailURL := ""
	if strings.HasPrefix(mimeType, "image/") {
		if thumbPath, err := ms.generateThumbnail(optimizedData, fileName, mimeType); err == nil {
			thumbnailURL = fmt.Sprintf("/media/thumbnails/%s", filepath.Base(thumbPath))
			// Console log for tracing thumbnail URL generation
			logrus.WithFields(logrus.Fields{
				"file_name":      fileName,
				"thumbnail_url":  thumbnailURL,
				"thumbnail_path": thumbPath,
			}).Info("🔍 MEDIA SERVICE: THUMBNAIL URL GENERATED FOR TRACING")
		}
	}

	result := &MediaUploadResult{
		FileName:     fileName,
		FileSize:     int64(len(optimizedData)),
		MimeType:     mimeType,
		URL:          localURL,
		CDNURL:       cdnURL,
		ThumbnailURL: thumbnailURL,
		Compressed:   compressed,
	}

	logrus.WithFields(logrus.Fields{
		"file_name":   fileName,
		"file_size":   result.FileSize,
		"mime_type":   mimeType,
		"compressed":  compressed,
		"cdn_enabled": ms.cdnEnabled,
	}).Info("File uploaded successfully")

	return result, nil
}

// ServeFile serves a file with caching for better performance
func (ms *MediaService) ServeFile(fileName string) ([]byte, string, error) {
	// Check cache first
	if cached := ms.getCachedFile(fileName); cached != nil {
		return cached.Data, cached.MimeType, nil
	}

	filePath := filepath.Join(ms.localBasePath, fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("file not found: %s", fileName)
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %v", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(fileData)

	// Cache the file
	ms.setCachedFile(fileName, fileData, mimeType)

	return fileData, mimeType, nil
}

// generateFileName creates a unique filename using MD5 hash
func (ms *MediaService) generateFileName(originalName string, data []byte) string {
	// Generate MD5 hash of file content
	hasher := md5.New()
	hasher.Write(data)
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Get file extension
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin" // Default extension
	}

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()

	return fmt.Sprintf("%s_%d%s", hash, timestamp, ext)
}

// optimizeImage compresses and optimizes images for better performance
func (ms *MediaService) optimizeImage(data []byte, mimeType string) ([]byte, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize if image is too large
	bounds := img.Bounds()
	maxWidth, maxHeight := 1920, 1080

	if bounds.Dx() > maxWidth || bounds.Dy() > maxHeight {
		img = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
		logrus.Debug("Image resized for optimization")
	}

	// Encode with compression
	var buf bytes.Buffer
	switch mimeType {
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	case "image/png":
		err = png.Encode(&buf, img)
	default:
		// For other formats, return original data
		return data, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode optimized image: %v", err)
	}

	return buf.Bytes(), nil
}

// generateThumbnail creates a thumbnail for images
func (ms *MediaService) generateThumbnail(data []byte, fileName, mimeType string) (string, error) {
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("thumbnails only supported for images")
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode image for thumbnail: %v", err)
	}

	// Create thumbnail (200x200 max)
	thumbnail := imaging.Fit(img, 200, 200, imaging.Lanczos)

	// Create thumbnails directory
	thumbnailDir := filepath.Join(ms.localBasePath, "thumbnails")
	os.MkdirAll(thumbnailDir, 0755)

	// Generate thumbnail filename
	thumbFileName := "thumb_" + fileName
	thumbPath := filepath.Join(thumbnailDir, thumbFileName)

	// Save thumbnail
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %v", err)
	}

	err = os.WriteFile(thumbPath, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %v", err)
	}

	return thumbPath, nil
}

// getCachedFile retrieves a file from cache
func (ms *MediaService) getCachedFile(fileName string) *CachedFile {
	ms.cacheMux.RLock()
	defer ms.cacheMux.RUnlock()

	cached, exists := ms.fileCache[fileName]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(cached.Timestamp) > ms.cacheTTL {
		// Cache expired, remove it
		go ms.removeCachedFile(fileName)
		return nil
	}

	return cached
}

// setCachedFile stores a file in cache
func (ms *MediaService) setCachedFile(fileName string, data []byte, mimeType string) {
	ms.cacheMux.Lock()
	defer ms.cacheMux.Unlock()

	ms.fileCache[fileName] = &CachedFile{
		Data:      data,
		MimeType:  mimeType,
		Timestamp: time.Now(),
	}

	// Clean up old cache entries periodically
	go ms.cleanupFileCache()
}

// removeCachedFile removes a specific file from cache
func (ms *MediaService) removeCachedFile(fileName string) {
	ms.cacheMux.Lock()
	defer ms.cacheMux.Unlock()
	delete(ms.fileCache, fileName)
}

// cleanupFileCache removes expired cache entries
func (ms *MediaService) cleanupFileCache() {
	ms.cacheMux.Lock()
	defer ms.cacheMux.Unlock()

	now := time.Now()
	for fileName, cached := range ms.fileCache {
		if now.Sub(cached.Timestamp) > ms.cacheTTL {
			delete(ms.fileCache, fileName)
		}
	}
}

// GetStats returns media service statistics
func (ms *MediaService) GetStats() map[string]interface{} {
	ms.cacheMux.RLock()
	cacheSize := len(ms.fileCache)
	ms.cacheMux.RUnlock()

	return map[string]interface{}{
		"cdn_enabled":     ms.cdnEnabled,
		"cdn_base_url":    ms.cdnBaseURL,
		"max_file_size":   ms.maxFileSize,
		"cached_files":    cacheSize,
		"allowed_types":   len(ms.allowedTypes),
		"local_base_path": ms.localBasePath,
	}
}

// DeleteFile removes a file from local storage and cache
func (ms *MediaService) DeleteFile(fileName string) error {
	// Remove from cache
	ms.removeCachedFile(fileName)

	// Remove local file
	filePath := filepath.Join(ms.localBasePath, fileName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	// Remove thumbnail if exists
	thumbnailPath := filepath.Join(ms.localBasePath, "thumbnails", "thumb_"+fileName)
	os.Remove(thumbnailPath) // Ignore errors for thumbnail deletion

	logrus.WithField("file_name", fileName).Info("File deleted successfully")
	return nil
}
