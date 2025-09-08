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
		"raw_content": replyContent,
		"content_length": len(replyContent),
	}).Debug("🔍 AI_PROCESSOR: Raw AI response received")
	
	// Remove markdown code blocks if present (exactly like PHP: '/^```json|```$/')
	sanitizedContent := regexp.MustCompile(`^` + "```" + `json|` + "```" + `$`).ReplaceAllString(strings.TrimSpace(replyContent), "")
	
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
			"stage": stage,
			"parts_count": len(replyParts),
		}).Info("✅ AI_PROCESSOR: Parsed standard JSON format")
	} else if matches := regexp.MustCompile(`Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$`).FindStringSubmatch(replyContent); len(matches) == 3 {
		// 2. Fallback for older format (Stage: Response:)
		stage = strings.TrimSpace(matches[1])
		responseJSON := matches[2]
		if err := json.Unmarshal([]byte(responseJSON), &replyParts); err == nil {
			logrus.WithFields(logrus.Fields{
				"stage": stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed Stage: Response: format")
		}
	} else if regexp.MustCompile(`^\s*{\s*"Stage":\s*".+?",\s*"Response":\s*\[.*\]\s*}\s*$`).MatchString(sanitizedContent) {
		// 3. Detect clean JSON format
		if err := json.Unmarshal([]byte(sanitizedContent), &data); err == nil && data.Stage != "" && len(data.Response) > 0 {
			stage = data.Stage
			replyParts = data.Response
			logrus.WithFields(logrus.Fields{
				"stage": stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed clean JSON format")
		} else {
			logrus.WithField("content", sanitizedContent).Error("Failed to parse specified JSON format")
			// In PHP, this returns early, but we'll continue to plain text fallback
		}
	} else if len(replyParts) > 0 && replyParts[0].Type == "text" && regexp.MustCompile(`^` + "```" + `json.*` + "```" + `$`).MatchString(replyParts[0].Content) {
		// 4. Encapsulated JSON within triple backticks (this only runs if replyParts already has content)
		jsonContent := regexp.MustCompile(`^` + "```" + `json|` + "```" + `$`).ReplaceAllString(strings.TrimSpace(replyParts[0].Content), "")
		var decodedContent AIResponseData
		if err := json.Unmarshal([]byte(jsonContent), &decodedContent); err == nil && decodedContent.Stage != "" && len(decodedContent.Response) > 0 {
			stage = decodedContent.Stage
			replyParts = decodedContent.Response
			logrus.WithFields(logrus.Fields{
				"stage": stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed encapsulated JSON format")
		} else {
			logrus.WithField("content", replyParts[0].Content).Error("Failed to parse encapsulated JSON")
			// In PHP, this returns early, but we'll continue to plain text fallback
		}
	} else {
		// 5. Before plain text fallback, try to extract images from the content
		// This handles cases where AI returns "Gambar 1: [URL]" format in plain text
		extractedParts, _ := ExtractImagesFromPlainText(replyContent)
		
		if len(extractedParts) > 0 {
			// We found structured content in the plain text
			replyParts = extractedParts
			if stage == "" {
				stage = "Response with Media"
			}
			logrus.WithFields(logrus.Fields{
				"parts_count": len(replyParts),
				"stage": stage,
			}).Info("🎯 AI_PROCESSOR: Extracted structured content from plain text response")
		} else {
			// 6. True plain text fallback - no structure found at all
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
		"stage": stage,
		"parts_count": len(replyParts),
	}).Info("📋 AI_PROCESSOR: Processing response parts")
	
	// Process reply parts exactly like PHP
	textParts := []string{}
	isOnemessageActive := false
	delay := time.Duration(delayMs) * time.Millisecond
	
	for index, part := range replyParts {
		// Log each part being processed
		logrus.WithFields(logrus.Fields{
			"index": index,
			"type": part.Type,
			"jenis": part.Jenis,
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
		"stage": stage,
		"total_messages": len(messages),
	}).Info("✅ AI_PROCESSOR: Response processing complete")
	
	return stage, messages, nil
}

// truncateString truncates a string to max length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ExtractImagesFromPlainText extracts images from various plain text formats
// This handles user-defined prompts that result in various response formats
func ExtractImagesFromPlainText(content string) ([]AIResponsePart, string) {
	var parts []AIResponsePart
	cleanText := content
	
	// Define patterns that indicate images in various formats users might prompt
	imagePatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		// Gambar 1: [URL] format (Malaysian/Indonesian)
		{"Gambar [URL]", regexp.MustCompile(`(?i)Gambar\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)},
		// Image 1: [URL] format (English)
		{"Image [URL]", regexp.MustCompile(`(?i)Image\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)},
		// Photo/Foto variations
		{"Photo [URL]", regexp.MustCompile(`(?i)(?:Photo|Foto)\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)},
		// Markdown image syntax ![alt](url)
		{"![alt](url)", regexp.MustCompile(`!\[[^\]]*\]\((https?://[^\)]+)\)`)},
		// Markdown link syntax that might be images [text](url.jpg)
		{"[text](image)", regexp.MustCompile(`\[[^\]]+\]\((https?://[^\)]+\.(?:jpg|jpeg|png|gif|webp|bmp|svg)(?:\?[^\)]*)?)\)`)},
		// Direct image URLs with common extensions
		{"Direct images", regexp.MustCompile(`https?://[^\s\[\]\(\)<>]+\.(?:jpg|jpeg|png|gif|webp|bmp|svg)(?:\?[^\s\[\]\(\)<>]*)?`)},
		// Numbered format: 1. https://... or 1) https://...
		{"Numbered URLs", regexp.MustCompile(`(?m)^\s*\d+[\.\)]\s*(https?://[^\s]+(?:jpg|jpeg|png|gif|webp))`)},
		// Bullet format: * https://... or - https://...
		{"Bullet URLs", regexp.MustCompile(`(?m)^\s*[\*\-]\s*(https?://[^\s]+(?:jpg|jpeg|png|gif|webp))`)},
	}
	
	var allImageURLs []string
	foundPatterns := make(map[string]bool)
	
	// Try each pattern to find images
	for _, p := range imagePatterns {
		matches := p.pattern.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			foundPatterns[p.name] = true
			for _, match := range matches {
				if len(match) >= 2 {
					imageURL := match[1]
					// Check for duplicates
					isDuplicate := false
					for _, existing := range allImageURLs {
						if existing == imageURL {
							isDuplicate = true
							break
						}
					}
					if !isDuplicate {
						allImageURLs = append(allImageURLs, imageURL)
						// Remove the entire match from clean text
						cleanText = strings.ReplaceAll(cleanText, match[0], "")
					}
				}
			}
		}
	}
	
	// Also check for video patterns
	videoPatterns := []string{
		`(?i)Video\s*\d*\s*:\s*\[(https?://[^\]]+)\]`,
		`https?://[^\s\[\]\(\)<>]+\.(?:mp4|avi|mov|wmv|flv|webm|mkv|m4v)(?:\?[^\s\[\]\(\)<>]*)?`,
	}
	
	for _, pattern := range videoPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				videoURL := match[1]
				if videoURL == "" && len(match) >= 1 {
					videoURL = match[0]
				}
				// Check for duplicates
				isDuplicate := false
				for _, existing := range allImageURLs {
					if existing == videoURL {
						isDuplicate = true
						break
					}
				}
				if !isDuplicate {
					// Add as video type
					parts = append(parts, AIResponsePart{
						Type:    "video",
						Content: videoURL,
					})
					cleanText = strings.ReplaceAll(cleanText, match[0], "")
				}
			}
		}
	}
	
	// Clean up the text
	cleanText = strings.TrimSpace(cleanText)
	// Remove multiple blank lines
	cleanText = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(cleanText, "\n\n")
	// Remove orphaned punctuation or numbering
	cleanText = regexp.MustCompile(`(?m)^\s*\d+[\.\)]\s*$`).ReplaceAllString(cleanText, "")
	cleanText = regexp.MustCompile(`(?m)^\s*[\*\-]\s*$`).ReplaceAllString(cleanText, "")
	cleanText = strings.TrimSpace(cleanText)
	
	// Build response parts
	if cleanText != "" {
		parts = append([]AIResponsePart{{Type: "text", Content: cleanText}}, parts...)
	}
	
	// Add all found images
	for _, imageURL := range allImageURLs {
		parts = append(parts, AIResponsePart{
			Type:    "image",
			Content: imageURL,
		})
	}
	
	// Log what we found
	if len(allImageURLs) > 0 || len(parts) > 0 {
		logrus.WithFields(logrus.Fields{
			"patterns_matched": foundPatterns,
			"images_found": len(allImageURLs),
			"total_parts": len(parts),
		}).Info("🔍 AI_PROCESSOR: Extracted media from plain text")
	}
	
	return parts, cleanText
}
