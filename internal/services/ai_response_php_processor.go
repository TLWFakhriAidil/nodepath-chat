package services

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AIResponsePart represents a single part of an AI response
type AIResponsePart struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Jenis   string `json:"Jenis,omitempty"`
}

// AIResponseData represents the AI response structure
type AIResponseData struct {
	Stage    string           `json:"Stage"`
	Response []AIResponsePart `json:"Response"`
}

// ProcessedAIMessage represents a message ready to send
type ProcessedAIMessage struct {
	Type    string // "text", "image", "audio", "video"
	Content string
	Delay   time.Duration
}

// ProcessAIResponsePHP processes AI response exactly like PHP code
func ProcessAIResponsePHP(replyContent string, delayMs int) (stage string, messages []ProcessedAIMessage, err error) {
	// Log raw input for debugging
	logrus.WithFields(logrus.Fields{
		"raw_content":    replyContent,
		"content_length": len(replyContent),
	}).Debug("🔍 AI_PROCESSOR: Raw AI response received")

	// Remove markdown code blocks if present (exactly like PHP: '/^```json|```$/')
	sanitizedContent := regexp.MustCompile(`^`+"```"+`json|`+"```"+`$`).ReplaceAllString(strings.TrimSpace(replyContent), "")

	// Log sanitized content
	logrus.WithFields(logrus.Fields{
		"sanitized_content": sanitizedContent,
	}).Debug("🔍 AI_PROCESSOR: Sanitized content")

	var data AIResponseData
	var replyParts []AIResponsePart

	// 1. Try to decode JSON directly (is_array($data) && isset($data['Stage']) && isset($data['Response']))
	err = json.Unmarshal([]byte(sanitizedContent), &data)
	if err == nil && data.Stage != "" && len(data.Response) > 0 {
		stage = data.Stage
		replyParts = data.Response
		logrus.WithFields(logrus.Fields{
			"stage":       stage,
			"parts_count": len(replyParts),
		}).Info("✅ AI_PROCESSOR: Parsed standard JSON format")
	} else if matches := regexp.MustCompile(`Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$`).FindStringSubmatch(replyContent); len(matches) == 3 {
		// 2. Fallback for older format (Stage: Response:)
		stage = strings.TrimSpace(matches[1])
		responseJSON := matches[2]
		if err := json.Unmarshal([]byte(responseJSON), &replyParts); err == nil {
			logrus.WithFields(logrus.Fields{
				"stage":       stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed Stage: Response: format")
		}
	} else if regexp.MustCompile(`^\s*{\s*"Stage":\s*".+?",\s*"Response":\s*\[.*\]\s*}\s*$`).MatchString(sanitizedContent) {
		// 3. Detect clean JSON format
		if err := json.Unmarshal([]byte(sanitizedContent), &data); err == nil && data.Stage != "" && len(data.Response) > 0 {
			stage = data.Stage
			replyParts = data.Response
			logrus.WithFields(logrus.Fields{
				"stage":       stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed clean JSON format")
		} else {
			logrus.WithField("content", sanitizedContent).Error("Failed to parse specified JSON format")
			// In PHP, this returns early, but we'll continue to plain text fallback
		}
	} else if len(replyParts) > 0 && replyParts[0].Type == "text" && regexp.MustCompile(`^`+"```"+`json.*`+"```"+`$`).MatchString(replyParts[0].Content) {
		// 4. Encapsulated JSON within triple backticks (this only runs if replyParts already has content)
		jsonContent := regexp.MustCompile(`^`+"```"+`json|`+"```"+`$`).ReplaceAllString(strings.TrimSpace(replyParts[0].Content), "")
		var decodedContent AIResponseData
		if err := json.Unmarshal([]byte(jsonContent), &decodedContent); err == nil && decodedContent.Stage != "" && len(decodedContent.Response) > 0 {
			stage = decodedContent.Stage
			replyParts = decodedContent.Response
			logrus.WithFields(logrus.Fields{
				"stage":       stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed encapsulated JSON format")
		} else {
			logrus.WithField("content", replyParts[0].Content).Error("Failed to parse encapsulated JSON")
			// In PHP, this returns early, but we'll continue to plain text fallback
		}
	} else {
		// 5. DYNAMIC EXTRACTION - Extract ALL URLs and let the system determine what they are
		extractedParts := ExtractAllMediaDynamically(replyContent)

		if len(extractedParts) > 0 {
			// We found content with media
			replyParts = extractedParts
			if stage == "" {
				stage = "Response with Media"
			}
			logrus.WithFields(logrus.Fields{
				"parts_count": len(replyParts),
				"stage":       stage,
			}).Info("🎯 AI_PROCESSOR: Dynamically extracted content from response")
		} else {
			// 6. True plain text fallback - no URLs found at all
			logrus.Warning("⚠️ AI_PROCESSOR: Plain text response detected. Defaulting to fallback handling.")
			if stage == "" {
				stage = "Problem Identification"
			}
			replyParts = []AIResponsePart{
				{Type: "text", Content: strings.TrimSpace(replyContent)},
			}
		}
	}

	// Validate we have replyParts
	if len(replyParts) == 0 {
		logrus.Error("Failed to decode the response JSON properly.")
		return stage, messages, nil // Return empty like PHP does
	}

	// Log the parts we're about to process
	logrus.WithFields(logrus.Fields{
		"stage":       stage,
		"parts_count": len(replyParts),
	}).Info("📋 AI_PROCESSOR: Processing response parts")

	// Process reply parts exactly like PHP
	textParts := []string{}
	isOnemessageActive := false
	delay := time.Duration(delayMs) * time.Millisecond

	for index, part := range replyParts {
		// Log each part being processed
		logrus.WithFields(logrus.Fields{
			"index":           index,
			"type":            part.Type,
			"jenis":           part.Jenis,
			"content_preview": truncateString(part.Content, 100),
		}).Debug("🔄 AI_PROCESSOR: Processing part")

		// Validate part structure
		if part.Type == "" || part.Content == "" {
			logrus.WithField("part", part).Warning("Invalid response part structure")
			continue
		}

		// Check if type=text and Jenis=onemessage
		if part.Type == "text" && part.Jenis == "onemessage" {
			// Start collecting
			textParts = append(textParts, part.Content)
			isOnemessageActive = true

			// Check if next part is also onemessage
			isLastPart := index == len(replyParts)-1
			nextIsNotOnemessage := false

			if !isLastPart {
				nextPart := replyParts[index+1]
				nextIsNotOnemessage = nextPart.Type != "text" || nextPart.Jenis != "onemessage"
			}

			// If this is last part OR next part is not onemessage, send combined
			if isLastPart || nextIsNotOnemessage {
				combinedMessage := strings.Join(textParts, "\n")
				messages = append(messages, ProcessedAIMessage{
					Type:    "text",
					Content: combinedMessage,
					Delay:   delay,
				})

				logrus.WithFields(logrus.Fields{
					"combined_parts": len(textParts),
					"message_length": len(combinedMessage),
				}).Info("✅ AI_PROCESSOR: Combined onemessage parts")

				// Reset
				textParts = []string{}
				isOnemessageActive = false
			}
		} else {
			// If we were collecting onemessage parts, flush them first
			if isOnemessageActive && len(textParts) > 0 {
				combinedMessage := strings.Join(textParts, "\n")
				messages = append(messages, ProcessedAIMessage{
					Type:    "text",
					Content: combinedMessage,
					Delay:   delay,
				})

				textParts = []string{}
				isOnemessageActive = false
			}

			// Now handle normal text or media
			if part.Type == "text" {
				messages = append(messages, ProcessedAIMessage{
					Type:    "text",
					Content: part.Content,
					Delay:   delay,
				})
			} else if part.Type == "image" || part.Type == "audio" || part.Type == "video" {
				// Clean and decode the URL (exactly like PHP: trim(urldecode($part['content'])))
				mediaURL := strings.TrimSpace(part.Content)

				// URL decode if needed
				if decodedURL, err := url.QueryUnescape(mediaURL); err == nil {
					mediaURL = decodedURL
				}

				messages = append(messages, ProcessedAIMessage{
					Type:    part.Type,
					Content: mediaURL,
					Delay:   delay,
				})
			}
		}
	}

	// Handle any remaining onemessage parts (shouldn't happen but just in case)
	if isOnemessageActive && len(textParts) > 0 {
		combinedMessage := strings.Join(textParts, "\n")
		messages = append(messages, ProcessedAIMessage{
			Type:    "text",
			Content: combinedMessage,
			Delay:   delay,
		})
	}

	// Log final processed messages
	logrus.WithFields(logrus.Fields{
		"stage":          stage,
		"total_messages": len(messages),
	}).Info("✅ AI_PROCESSOR: Response processing complete")

	return stage, messages, nil
}

// ExtractAllMediaDynamically extracts ALL URLs and determines their type dynamically
// This handles ANY format users might create in their prompts
func ExtractAllMediaDynamically(content string) []AIResponsePart {
	var parts []AIResponsePart

	// Find ALL URLs in the content using a very broad pattern
	// This will catch URLs in ANY format: [URL], (URL), <URL>, "URL", 'URL', or just URL
	urlPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\\^` + "`" + `\[\]]+`)
	allMatches := urlPattern.FindAllStringIndex(content, -1)

	if len(allMatches) == 0 {
		return parts // No URLs found, return empty
	}

	// Process content and build parts
	lastEnd := 0

	for _, match := range allMatches {
		urlStart := match[0]
		urlEnd := match[1]
		url := content[urlStart:urlEnd]

		// Clean the URL
		url = strings.TrimRight(url, ".,;:!?)]}")

		// Look for context around the URL to remove (like "Gambar 1: [URL]")
		contextStart := urlStart
		contextEnd := urlEnd

		// Check backwards for media indicators
		beforeURL := ""
		if urlStart > 0 {
			// Look back up to 50 characters or to the previous newline
			lookbackStart := urlStart - 50
			if lookbackStart < 0 {
				lookbackStart = 0
			}
			beforeURL = content[lookbackStart:urlStart]

			// Find media indicator patterns before the URL
			indicatorPatterns := []string{
				`(?i)Gambar\s*\d*\s*:\s*\[?$`,
				`(?i)Image\s*\d*\s*:\s*\[?$`,
				`(?i)Photo\s*\d*\s*:\s*\[?$`,
				`(?i)Foto\s*\d*\s*:\s*\[?$`,
				`(?i)Video\s*\d*\s*:\s*\[?$`,
			}

			for _, pattern := range indicatorPatterns {
				re := regexp.MustCompile(pattern)
				if matches := re.FindStringIndex(beforeURL); matches != nil {
					// Found an indicator, extend context to include it
					contextStart = lookbackStart + matches[0]
					break
				}
			}
		}

		// Check after URL for closing brackets/parentheses
		if urlEnd < len(content) {
			afterURL := ""
			lookAheadEnd := urlEnd + 10
			if lookAheadEnd > len(content) {
				lookAheadEnd = len(content)
			}
			afterURL = content[urlEnd:lookAheadEnd]

			// Count how many closing brackets/parentheses to include
			for i, char := range afterURL {
				if char == ']' || char == ')' || char == '}' || char == '>' {
					contextEnd = urlEnd + i + 1
				} else {
					break
				}
			}
		}

		// Extract text before this media (if any)
		if contextStart > lastEnd {
			textBefore := strings.TrimSpace(content[lastEnd:contextStart])
			if textBefore != "" {
				parts = append(parts, AIResponsePart{
					Type:    "text",
					Content: textBefore,
				})
			}
		}

		// Add the media
		mediaType := determineMediaType(url)
		parts = append(parts, AIResponsePart{
			Type:    mediaType,
			Content: url,
		})

		lastEnd = contextEnd
	}

	// Get any remaining text after the last URL
	if lastEnd < len(content) {
		remainingText := strings.TrimSpace(content[lastEnd:])
		if remainingText != "" {
			parts = append(parts, AIResponsePart{
				Type:    "text",
				Content: remainingText,
			})
		}
	}

	logrus.WithFields(logrus.Fields{
		"total_urls":  len(allMatches),
		"total_parts": len(parts),
	}).Info("🔍 AI_PROCESSOR: Dynamically extracted media from content")

	return parts
}

// determineMediaType checks URL to determine if it's image, video, audio, or generic media
func determineMediaType(url string) string {
	lowerURL := strings.ToLower(url)

	// Check by file extension
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	for _, ext := range imageExts {
		if strings.Contains(lowerURL, ext) {
			return "image"
		}
	}

	videoExts := []string{".mp4", ".avi", ".mov", ".webm", ".mkv", ".m4v"}
	for _, ext := range videoExts {
		if strings.Contains(lowerURL, ext) {
			return "video"
		}
	}

	audioExts := []string{".mp3", ".wav", ".ogg", ".m4a", ".aac"}
	for _, ext := range audioExts {
		if strings.Contains(lowerURL, ext) {
			return "audio"
		}
	}

	// Check for common image hosting patterns
	if strings.Contains(lowerURL, "/images/") ||
		strings.Contains(lowerURL, "/image/") ||
		strings.Contains(lowerURL, "/img/") ||
		strings.Contains(lowerURL, "/photo/") ||
		strings.Contains(lowerURL, "/gambar/") ||
		strings.Contains(lowerURL, "chatgpt") { // Your specific pattern
		return "image"
	}

	// Default to image for unrecognized media
	return "image"
}

// cleanTextFromMediaIndicators removes common media indicators from text
func cleanTextFromMediaIndicators(text string) string {
	// Remove common patterns that indicate media references
	patterns := []string{
		`(?i)Gambar\s*\d*\s*:\s*$`,
		`(?i)Image\s*\d*\s*:\s*$`,
		`(?i)Photo\s*\d*\s*:\s*$`,
		`(?i)Foto\s*\d*\s*:\s*$`,
		`(?i)Video\s*\d*\s*:\s*$`,
		`^\[\s*\]$`,    // Empty brackets
		`^\(\s*\)$`,    // Empty parentheses
		`^\d+[.)\s]*$`, // Just numbers like "1." or "1)"
		`^[*-]\s*$`,    // Just bullets
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	// Clean up multiple blank lines
	text = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// BuildCleanResponse builds a clean response string from processed messages
// This is what should be saved to the database (text only, no media URLs)
func BuildCleanResponse(messages []ProcessedAIMessage) string {
	var textParts []string
	for _, msg := range messages {
		if msg.Type == "text" {
			textParts = append(textParts, msg.Content)
		}
	}
	return strings.Join(textParts, "\n\n")
}

// truncateString truncates a string to max length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
