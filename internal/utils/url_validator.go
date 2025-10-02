package utils

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// URLValidator provides URL validation functionality
type URLValidator struct {
	client *http.Client
}

// NewURLValidator creates a new URL validator with timeout configuration
func NewURLValidator() *URLValidator {
	return &URLValidator{
		client: &http.Client{
			Timeout: 10 * time.Second, // 10 second timeout for URL validation
		},
	}
}

// ValidateMediaURL validates if a media URL is accessible and returns appropriate media type
// Returns: isValid, mediaType, error
func (v *URLValidator) ValidateMediaURL(url string) (bool, string, error) {
	// Basic URL format validation
	if url == "" {
		return false, "", fmt.Errorf("empty URL")
	}

	// Check if URL starts with http or https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false, "", fmt.Errorf("invalid URL format: must start with http:// or https://")
	}

	// Log validation attempt
	logrus.WithFields(logrus.Fields{
		"url":        url,
		"url_length": len(url),
	}).Info("🔍 URL_VALIDATOR: Validating media URL accessibility")

	// Make HEAD request to check if URL is accessible
	resp, err := v.client.Head(url)
	if err != nil {
		logrus.WithError(err).WithField("url", url).Warn("❌ URL_VALIDATOR: Failed to access URL")
		return false, "", fmt.Errorf("URL not accessible: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.WithFields(logrus.Fields{
			"url":         url,
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		}).Warn("❌ URL_VALIDATOR: URL returned non-success status code")
		return false, "", fmt.Errorf("URL returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Determine media type from Content-Type header or URL extension
	mediaType := v.determineMediaType(url, resp.Header.Get("Content-Type"))

	logrus.WithFields(logrus.Fields{
		"url":          url,
		"status_code":  resp.StatusCode,
		"content_type": resp.Header.Get("Content-Type"),
		"media_type":   mediaType,
	}).Info("✅ URL_VALIDATOR: URL validation successful")

	return true, mediaType, nil
}

// determineMediaType determines the media type from URL extension or Content-Type header
func (v *URLValidator) determineMediaType(url, contentType string) string {
	// Check Content-Type header first
	if contentType != "" {
		contentType = strings.ToLower(contentType)
		if strings.Contains(contentType, "image") {
			return "image"
		}
		if strings.Contains(contentType, "video") {
			return "video"
		}
		if strings.Contains(contentType, "audio") {
			return "audio"
		}
	}

	// Fallback to URL extension
	url = strings.ToLower(url)

	// Image extensions
	if strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg") ||
		strings.Contains(url, ".png") || strings.Contains(url, ".gif") ||
		strings.Contains(url, ".webp") || strings.Contains(url, ".bmp") {
		return "image"
	}

	// Video extensions
	if strings.Contains(url, ".mp4") || strings.Contains(url, ".avi") ||
		strings.Contains(url, ".mov") || strings.Contains(url, ".wmv") ||
		strings.Contains(url, ".flv") || strings.Contains(url, ".webm") {
		return "video"
	}

	// Audio extensions
	if strings.Contains(url, ".mp3") || strings.Contains(url, ".wav") ||
		strings.Contains(url, ".ogg") || strings.Contains(url, ".m4a") ||
		strings.Contains(url, ".flac") || strings.Contains(url, ".aac") {
		return "audio"
	}

	// Default to image if cannot determine
	return "image"
}

// ValidateURLQuick performs a quick validation without HTTP request
// Useful for basic format checking before making network calls
func (v *URLValidator) ValidateURLQuick(url string) bool {
	if url == "" {
		return false
	}

	// Check if URL starts with http or https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// Check for basic URL structure
	if !strings.Contains(url, ".") {
		return false
	}

	return true
}
