package services

import (
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/sirupsen/logrus"
)

// SmartMediaDetector detects media by checking actual URLs rather than formats
type SmartMediaDetector struct {
	client *http.Client
}

// NewSmartMediaDetector creates a new smart media detector
func NewSmartMediaDetector() *SmartMediaDetector {
	return &SmartMediaDetector{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ExtractAndValidateMedia extracts ALL URLs from text and validates them
func (s *SmartMediaDetector) ExtractAndValidateMedia(text string) ([]MediaDetectionResult, string) {
	var validMedia []MediaDetectionResult
	cleanText := text
	
	// Extract ALL URLs from the text using a simple pattern
	urlPattern := regexp.MustCompile(`https?://[^\s\)\]\}]+`)
	allURLs := urlPattern.FindAllString(text, -1)
	
	// Track which URLs we've already processed to avoid duplicates
	processedURLs := make(map[string]bool)
	
	for _, url := range allURLs {
		// Clean the URL (remove trailing punctuation)
		url = strings.TrimRight(url, ".,;:!?")
		
		// Skip if already processed
		if processedURLs[url] {
			continue
		}
		processedURLs[url] = true
		
		// Check if this URL is actually a media file
		if mediaType := s.checkIfMedia(url); mediaType != "" {
			validMedia = append(validMedia, MediaDetectionResult{
				IsMedia:   true,
				MediaType: mediaType,
				MediaURL:  url,
			})
			
			// Remove this URL from the text (including any surrounding brackets/formatting)
			// Remove patterns like [text](url), [url], etc.
			patterns := []string{
				`\[[^\]]*\]\(` + regexp.QuoteMeta(url) + `\)`, // [text](url)
				`\[` + regexp.QuoteMeta(url) + `\]`,           // [url]
				regexp.QuoteMeta(url),                          // just url
			}
			
			for _, pattern := range patterns {
				re := regexp.MustCompile(pattern)
				cleanText = re.ReplaceAllString(cleanText, "")
			}
			
			logrus.WithFields(logrus.Fields{
				"url": url,
				"type": mediaType,
			}).Info("✅ SMART_MEDIA: Validated media URL")
		}
	}
	
	// Clean up extra whitespace and empty lines
	cleanText = regexp.MustCompile(`\s+\n`).ReplaceAllString(cleanText, "\n")
	cleanText = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleanText, "\n\n")
	cleanText = strings.TrimSpace(cleanText)
	
	return validMedia, cleanText
}

// checkIfMedia validates if a URL is actually a media file
func (s *SmartMediaDetector) checkIfMedia(url string) string {
	// First check by extension
	lowerURL := strings.ToLower(url)
	
	// Image extensions
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	for _, ext := range imageExts {
		if strings.Contains(lowerURL, ext) {
			return "image"
		}
	}
	
	// Video extensions
	videoExts := []string{".mp4", ".avi", ".mov", ".webm", ".mkv"}
	for _, ext := range videoExts {
		if strings.Contains(lowerURL, ext) {
			return "video"
		}
	}
	
	// Audio extensions
	audioExts := []string{".mp3", ".wav", ".ogg", ".m4a"}
	for _, ext := range audioExts {
		if strings.Contains(lowerURL, ext) {
			return "audio"
		}
	}
	
	// If no extension found, try HEAD request to check Content-Type
	resp, err := s.client.Head(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	
	// Check Content-Type header
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	}
	
	return ""
}
