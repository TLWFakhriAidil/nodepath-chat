package services

import (
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// MediaDetectionService handles detection of media URLs in various formats
type MediaDetectionService struct {
	// Regex patterns for different media URL formats
	bracketPattern       *regexp.Regexp
	simpleBracketPattern *regexp.Regexp
	labeledPattern       *regexp.Regexp
	markdownLinkPattern  *regexp.Regexp
	directURLPattern     *regexp.Regexp
	markdownPattern      *regexp.Regexp
	aiGeneratedPattern   *regexp.Regexp
	urlPattern           *regexp.Regexp
}

// MediaDetectionResult contains the result of media detection
type MediaDetectionResult struct {
	IsMedia      bool
	MediaType    string // "image", "audio", "video"
	MediaURL     string
	OriginalText string
	CleanText    string // Text with media URLs removed
}

// NewMediaDetectionService creates a new media detection service
func NewMediaDetectionService() *MediaDetectionService {
	// Bracket format: [IMAGE: URL], [AUDIO: URL], [VIDEO: URL] with optional backticks
	bracketPattern := regexp.MustCompile(`\[(IMAGE|AUDIO|VIDEO):\s*` + "`" + `?([^\]` + "`" + `]+)` + "`" + `?\]`)

	// Simple bracket format: [URL] - just URL in square brackets (common in AI responses)
	simpleBracketPattern := regexp.MustCompile(`\[(https?://[^\]]+)\]`)

	// Labeled format: "Gambar X: [URL]" or "Image X: [URL]" - common in multilingual AI responses
	labeledPattern := regexp.MustCompile(`(?:Gambar|Image|Foto|Picture|Video|Audio)\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)

	// Markdown link format: [Text](URL) - e.g., [Gambar: 1 Botol Vitac](https://...)
	markdownLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\)]+)\)`)

	// Direct URL pattern: detect common media file extensions
	directURLPattern := regexp.MustCompile(`https?://[^\s\[\]()]+\.(jpg|jpeg|png|gif|bmp|webp|svg|mp3|wav|flac|aac|ogg|wma|m4a|mp4|avi|mov|wmv|flv|webm|mkv|m4v)(?:\?[^\s\[\]()]*)?`)

	// Markdown format: ![alt](URL) for images
	markdownPattern := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)

	// AI-generated format: ! `URL` (exclamation mark followed by backtick-wrapped URL)
	aiGeneratedPattern := regexp.MustCompile(`!\s*` + "`" + `(https?://[^` + "`" + `\s]+)` + "`" + ``)

	// General URL pattern
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)

	return &MediaDetectionService{
		bracketPattern:       bracketPattern,
		simpleBracketPattern: simpleBracketPattern,
		labeledPattern:       labeledPattern,
		markdownLinkPattern:  markdownLinkPattern,
		directURLPattern:     directURLPattern,
		markdownPattern:      markdownPattern,
		aiGeneratedPattern:   aiGeneratedPattern,
		urlPattern:           urlPattern,
	}
}

// DetectMedia analyzes text for media URLs and returns detection results
func (mds *MediaDetectionService) DetectMedia(text string) []MediaDetectionResult {
	var results []MediaDetectionResult
	cleanText := text

	// 1. Check for bracket format media URLs [IMAGE: URL], [AUDIO: URL], [VIDEO: URL]
	bracketMatches := mds.bracketPattern.FindAllStringSubmatch(text, -1)
	for _, match := range bracketMatches {
		if len(match) >= 3 {
			mediaType := strings.ToLower(match[1])
			mediaURL := strings.TrimSpace(match[2])

			// Remove backticks if present
			mediaURL = strings.Trim(mediaURL, "`")

			results = append(results, MediaDetectionResult{
				IsMedia:      true,
				MediaType:    mediaType,
				MediaURL:     mediaURL,
				OriginalText: match[0],
			})

			// Remove from clean text
			cleanText = strings.ReplaceAll(cleanText, match[0], "")

			logrus.WithFields(logrus.Fields{
				"media_type": mediaType,
				"media_url":  mediaURL,
				"format":     "bracket",
			}).Info("📎 MEDIA DETECTION: Found bracket format media URL")
		}
	}

	// 2. Check for labeled format "Gambar X: [URL]" - common in AI responses
	labeledMatches := mds.labeledPattern.FindAllStringSubmatch(text, -1)
	for _, match := range labeledMatches {
		if len(match) >= 2 {
			mediaURL := strings.TrimSpace(match[1])
			mediaType := mds.getMediaTypeFromURL(mediaURL)
			if mediaType == "" {
				mediaType = "image" // Default to image for labeled formats
			}

			results = append(results, MediaDetectionResult{
				IsMedia:      true,
				MediaType:    mediaType,
				MediaURL:     mediaURL,
				OriginalText: match[0],
			})

			// Remove from clean text
			cleanText = strings.ReplaceAll(cleanText, match[0], "")

			logrus.WithFields(logrus.Fields{
				"media_type": mediaType,
				"media_url":  mediaURL,
				"format":     "labeled",
			}).Info("📎 MEDIA DETECTION: Found labeled format media URL (Gambar/Image)")
		}
	}

	// 3. Check for markdown link format [Text](URL) - e.g., [Gambar: 1 Botol Vitac](https://...)
	markdownLinkMatches := mds.markdownLinkPattern.FindAllStringSubmatch(text, -1)
	for _, match := range markdownLinkMatches {
		if len(match) >= 3 {
			linkText := strings.TrimSpace(match[1])
			mediaURL := strings.TrimSpace(match[2])
			mediaType := mds.getMediaTypeFromURL(mediaURL)

			// Process if it's a media URL or if link text contains media keywords
			if mediaType != "" || strings.Contains(strings.ToLower(linkText), "gambar") ||
				strings.Contains(strings.ToLower(linkText), "image") ||
				strings.Contains(strings.ToLower(linkText), "foto") ||
				strings.Contains(strings.ToLower(linkText), "picture") {

				if mediaType == "" {
					mediaType = "image" // Default to image for links with media keywords
				}

				// Check if this URL was already processed
				alreadyProcessed := false
				for _, result := range results {
					if result.MediaURL == mediaURL {
						alreadyProcessed = true
						break
					}
				}

				if !alreadyProcessed {
					results = append(results, MediaDetectionResult{
						IsMedia:      true,
						MediaType:    mediaType,
						MediaURL:     mediaURL,
						OriginalText: match[0],
					})

					// Remove from clean text
					cleanText = strings.ReplaceAll(cleanText, match[0], "")

					logrus.WithFields(logrus.Fields{
						"media_type": mediaType,
						"media_url":  mediaURL,
						"link_text":  linkText,
						"format":     "markdown_link",
					}).Info("📎 MEDIA DETECTION: Found markdown link format media URL")
				}
			}
		}
	}

	// 4. Check for simple bracket format [URL] - just URL in brackets
	simpleBracketMatches := mds.simpleBracketPattern.FindAllStringSubmatch(text, -1)
	for _, match := range simpleBracketMatches {
		if len(match) >= 2 {
			mediaURL := strings.TrimSpace(match[1])
			mediaType := mds.getMediaTypeFromURL(mediaURL)

			// Only process if it's a media URL (has media extension)
			if mediaType != "" {
				// Check if this URL was already processed by labeled pattern
				alreadyProcessed := false
				for _, result := range results {
					if result.MediaURL == mediaURL {
						alreadyProcessed = true
						break
					}
				}

				if !alreadyProcessed {
					results = append(results, MediaDetectionResult{
						IsMedia:      true,
						MediaType:    mediaType,
						MediaURL:     mediaURL,
						OriginalText: match[0],
					})

					// Remove from clean text
					cleanText = strings.ReplaceAll(cleanText, match[0], "")

					logrus.WithFields(logrus.Fields{
						"media_type": mediaType,
						"media_url":  mediaURL,
						"format":     "simple_bracket",
					}).Info("📎 MEDIA DETECTION: Found simple bracket format media URL")
				}
			}
		}
	}

	// 5. Check for direct media URLs with file extensions
	directMatches := mds.directURLPattern.FindAllString(text, -1)
	for _, url := range directMatches {
		mediaType := mds.getMediaTypeFromURL(url)
		if mediaType != "" {
			// Check if this URL was already processed
			alreadyProcessed := false
			for _, result := range results {
				if result.MediaURL == url {
					alreadyProcessed = true
					break
				}
			}

			if !alreadyProcessed {
				results = append(results, MediaDetectionResult{
					IsMedia:      true,
					MediaType:    mediaType,
					MediaURL:     url,
					OriginalText: url,
				})

				// Remove from clean text
				cleanText = strings.ReplaceAll(cleanText, url, "")

				logrus.WithFields(logrus.Fields{
					"media_type": mediaType,
					"media_url":  url,
					"format":     "direct",
				}).Info("📎 MEDIA DETECTION: Found direct media URL")
			}
		}
	}

	// 6. Check for markdown format media URLs ![alt](URL)
	markdownMatches := mds.markdownPattern.FindAllStringSubmatch(text, -1)
	for _, match := range markdownMatches {
		if len(match) >= 2 {
			url := match[1]
			mediaType := mds.getMediaTypeFromURL(url)
			if mediaType != "" {
				results = append(results, MediaDetectionResult{
					IsMedia:      true,
					MediaType:    mediaType,
					MediaURL:     url,
					OriginalText: match[0],
				})

				// Remove from clean text
				cleanText = strings.ReplaceAll(cleanText, match[0], "")

				logrus.WithFields(logrus.Fields{
					"media_type": mediaType,
					"media_url":  url,
					"format":     "markdown",
				}).Info("📎 MEDIA DETECTION: Found markdown format media URL")
			}
		}
	}

	// 6. Check for AI-generated format media URLs ! `URL`
	aiGeneratedMatches := mds.aiGeneratedPattern.FindAllStringSubmatch(text, -1)
	for _, match := range aiGeneratedMatches {
		if len(match) >= 2 {
			url := match[1]
			mediaType := mds.getMediaTypeFromURL(url)
			if mediaType != "" {
				results = append(results, MediaDetectionResult{
					IsMedia:      true,
					MediaType:    mediaType,
					MediaURL:     url,
					OriginalText: match[0],
				})

				// Remove from clean text
				cleanText = strings.ReplaceAll(cleanText, match[0], "")

				logrus.WithFields(logrus.Fields{
					"media_type": mediaType,
					"media_url":  url,
					"format":     "ai_generated",
				}).Info("📎 MEDIA DETECTION: Found AI-generated format media URL")
			}
		}
	}

	// Update clean text for all results
	for i := range results {
		results[i].CleanText = strings.TrimSpace(cleanText)
	}

	return results
}

// getMediaTypeFromURL determines media type based on file extension
func (mds *MediaDetectionService) getMediaTypeFromURL(url string) string {
	lowerURL := strings.ToLower(url)

	// Image extensions
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff", ".tif"}
	for _, ext := range imageExtensions {
		if strings.Contains(lowerURL, ext) {
			return "image"
		}
	}

	// Audio extensions
	audioExtensions := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".opus", ".aiff", ".au"}
	for _, ext := range audioExtensions {
		if strings.Contains(lowerURL, ext) {
			return "audio"
		}
	}

	// Video extensions
	videoExtensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v", ".3gp", ".ogv", ".ts", ".mts"}
	for _, ext := range videoExtensions {
		if strings.Contains(lowerURL, ext) {
			return "video"
		}
	}

	return ""
}

// HasMedia checks if text contains any media URLs
func (mds *MediaDetectionService) HasMedia(text string) bool {
	results := mds.DetectMedia(text)
	return len(results) > 0
}

// ExtractFirstMedia returns the first media URL found in text
func (mds *MediaDetectionService) ExtractFirstMedia(text string) *MediaDetectionResult {
	results := mds.DetectMedia(text)
	if len(results) > 0 {
		return &results[0]
	}
	return nil
}

// ExtractAllMedia returns all media URLs found in text
func (mds *MediaDetectionService) ExtractAllMedia(text string) []MediaDetectionResult {
	return mds.DetectMedia(text)
}

// RemoveMediaURLs removes all media URLs from text and returns clean text
func (mds *MediaDetectionService) RemoveMediaURLs(text string) string {
	cleanText := text

	// Remove bracket format media URLs
	bracketMatches := mds.bracketPattern.FindAllStringSubmatch(text, -1)
	for _, match := range bracketMatches {
		if len(match) > 0 {
			cleanText = strings.ReplaceAll(cleanText, match[0], "")
		}
	}

	// Remove labeled format media URLs (Gambar X: [URL])
	labeledMatches := mds.labeledPattern.FindAllStringSubmatch(text, -1)
	for _, match := range labeledMatches {
		if len(match) > 0 {
			cleanText = strings.ReplaceAll(cleanText, match[0], "")
		}
	}

	// Remove markdown link format media URLs [Text](URL)
	markdownLinkMatches := mds.markdownLinkPattern.FindAllStringSubmatch(text, -1)
	for _, match := range markdownLinkMatches {
		if len(match) > 0 {
			linkText := ""
			if len(match) > 1 {
				linkText = match[1]
			}
			// Only remove if it's a media link
			if strings.Contains(strings.ToLower(linkText), "gambar") ||
				strings.Contains(strings.ToLower(linkText), "image") ||
				strings.Contains(strings.ToLower(linkText), "foto") ||
				(len(match) > 2 && mds.getMediaTypeFromURL(match[2]) != "") {
				cleanText = strings.ReplaceAll(cleanText, match[0], "")
			}
		}
	}

	// Remove simple bracket format media URLs
	simpleBracketMatches := mds.simpleBracketPattern.FindAllStringSubmatch(text, -1)
	for _, match := range simpleBracketMatches {
		if len(match) > 0 && mds.getMediaTypeFromURL(match[1]) != "" {
			cleanText = strings.ReplaceAll(cleanText, match[0], "")
		}
	}

	// Remove markdown format media URLs
	markdownMatches := mds.markdownPattern.FindAllStringSubmatch(text, -1)
	for _, match := range markdownMatches {
		if len(match) > 0 && mds.getMediaTypeFromURL(match[1]) != "" {
			cleanText = strings.ReplaceAll(cleanText, match[0], "")
		}
	}

	// Remove AI-generated format media URLs
	aiGeneratedMatches := mds.aiGeneratedPattern.FindAllStringSubmatch(text, -1)
	for _, match := range aiGeneratedMatches {
		if len(match) > 0 && mds.getMediaTypeFromURL(match[1]) != "" {
			cleanText = strings.ReplaceAll(cleanText, match[0], "")
		}
	}

	// Remove direct URL media
	urlMatches := mds.urlPattern.FindAllString(text, -1)
	for _, url := range urlMatches {
		if mds.getMediaTypeFromURL(url) != "" {
			cleanText = strings.ReplaceAll(cleanText, url, "")
		}
	}

	// Clean up extra whitespace and newlines
	cleanText = strings.TrimSpace(cleanText)
	// Replace multiple newlines with double newline
	cleanText = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleanText, "\n\n")
	// Remove lines that only contain "Gambar X:" or similar
	cleanText = regexp.MustCompile(`(?m)^(Gambar|Image|Video|Audio)\s*\d*:?\s*$`).ReplaceAllString(cleanText, "")
	// Clean up extra whitespace again
	cleanText = strings.TrimSpace(cleanText)

	return cleanText
}
