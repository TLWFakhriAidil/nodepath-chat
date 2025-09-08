package services

import (
	"encoding/json"
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
	
	// Remove markdown code blocks if present
	sanitizedContent := regexp.MustCompile(`^` + "```" + `json|` + "```" + `$`).ReplaceAllString(strings.TrimSpace(replyContent), "")
	
	// Log sanitized content
	logrus.WithFields(logrus.Fields{
		"sanitized_content": sanitizedContent,
	}).Debug("🔍 AI_PROCESSOR: Sanitized content")
	
	var data AIResponseData
	var replyParts []AIResponsePart
	
	// Try to decode JSON directly
	if err := json.Unmarshal([]byte(sanitizedContent), &data); err == nil {
		if data.Stage != "" && len(data.Response) > 0 {
			stage = data.Stage
			replyParts = data.Response
			logrus.WithFields(logrus.Fields{
				"stage": stage,
				"parts_count": len(replyParts),
			}).Info("✅ AI_PROCESSOR: Parsed standard JSON format")
		}
	}
	
	// If not parsed yet, try Stage: Response: format
	if len(replyParts) == 0 {
		stageResponsePattern := regexp.MustCompile(`Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$`)
		if matches := stageResponsePattern.FindStringSubmatch(replyContent); len(matches) == 3 {
			stage = strings.TrimSpace(matches[1])
			responseJSON := matches[2]
			if err := json.Unmarshal([]byte(responseJSON), &replyParts); err == nil {
				logrus.WithFields(logrus.Fields{
					"stage": stage,
					"parts_count": len(replyParts),
				}).Info("✅ AI_PROCESSOR: Parsed Stage: Response: format")
			}
		}
	}
	
	// Try detecting clean JSON format
	if len(replyParts) == 0 {
		jsonPattern := regexp.MustCompile(`^\s*{\s*"Stage":\s*".+?",\s*"Response":\s*\[.*\]\s*}\s*$`)
		if jsonPattern.MatchString(sanitizedContent) {
			if err := json.Unmarshal([]byte(sanitizedContent), &data); err == nil {
				if data.Stage != "" && len(data.Response) > 0 {
					stage = data.Stage
					replyParts = data.Response
					logrus.WithFields(logrus.Fields{
						"stage": stage,
						"parts_count": len(replyParts),
					}).Info("✅ AI_PROCESSOR: Parsed clean JSON format")
				}
			}
		}
	}
	
	// Check for encapsulated JSON in first response item
	if len(replyParts) > 0 && replyParts[0].Type == "text" {
		encapsulatedPattern := regexp.MustCompile(`^` + "```" + `json.*` + "```" + `$`)
		if encapsulatedPattern.MatchString(replyParts[0].Content) {
			jsonContent := regexp.MustCompile(`^` + "```" + `json|` + "```" + `$`).ReplaceAllString(strings.TrimSpace(replyParts[0].Content), "")
			var decodedContent AIResponseData
			if err := json.Unmarshal([]byte(jsonContent), &decodedContent); err == nil {
				if decodedContent.Stage != "" && len(decodedContent.Response) > 0 {
					stage = decodedContent.Stage
					replyParts = decodedContent.Response
					logrus.WithFields(logrus.Fields{
						"stage": stage,
						"parts_count": len(replyParts),
					}).Info("✅ AI_PROCESSOR: Parsed encapsulated JSON format")
				}
			}
		}
	}
	
	// Before fallback, check if content contains image patterns
	if len(replyParts) == 0 {
		// Multiple patterns for detecting images in AI responses
		patterns := []struct {
			name    string
			pattern *regexp.Regexp
		}{
			{"Gambar X: [URL]", regexp.MustCompile(`Gambar\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)},
			{"Image X: [URL]", regexp.MustCompile(`Image\s*\d*\s*:\s*\[(https?://[^\]]+)\]`)},
			{"[Text](URL)", regexp.MustCompile(`\[[^\]]+\]\((https?://[^\)]+\.(?:jpg|jpeg|png|gif|webp)[^\)]*)\)`)},
			{"![Text](URL)", regexp.MustCompile(`!\[[^\]]*\]\((https?://[^\)]+)\)`)},
		}
		
		var allImageURLs []string
		cleanText := replyContent
		
		// Check each pattern
		for _, p := range patterns {
			matches := p.pattern.FindAllStringSubmatch(replyContent, -1)
			if len(matches) > 0 {
				logrus.WithFields(logrus.Fields{
					"pattern": p.name,
					"matches": len(matches),
				}).Info("🖼️ AI_PROCESSOR: Detected image pattern")
				
				for _, match := range matches {
					if len(match) >= 2 {
						allImageURLs = append(allImageURLs, match[1])
						// Remove the matched pattern from clean text
						cleanText = strings.ReplaceAll(cleanText, match[0], "")
					}
				}
			}
		}
		
		// Also check for direct image URLs
		directImagePattern := regexp.MustCompile(`https?://[^\s\[\]\(\)]+\.(?:jpg|jpeg|png|gif|webp)(?:\?[^\s\[\]\(\)]*)?`)
		directMatches := directImagePattern.FindAllString(replyContent, -1)
		for _, url := range directMatches {
			// Check if not already captured
			isDuplicate := false
			for _, existing := range allImageURLs {
				if existing == url {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				allImageURLs = append(allImageURLs, url)
				cleanText = strings.ReplaceAll(cleanText, url, "")
			}
		}
		
		// If we found images, create structured response
		if len(allImageURLs) > 0 {
			logrus.WithField("total_images", len(allImageURLs)).Info("🖼️ AI_PROCESSOR: Creating structured response with images")
			
			// Clean up the text
			cleanText = strings.TrimSpace(cleanText)
			cleanText = regexp.MustCompile(`\s+\n`).ReplaceAllString(cleanText, "\n")
			cleanText = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleanText, "\n\n")
			
			// Add clean text first if exists
			if cleanText != "" {
				replyParts = append(replyParts, AIResponsePart{
					Type:    "text",
					Content: cleanText,
				})
			}
			
			// Add each image
			for _, imageURL := range allImageURLs {
				replyParts = append(replyParts, AIResponsePart{
					Type:    "image",
					Content: imageURL,
				})
			}
			
			if stage == "" {
				stage = "Response with Images"
			}
		}
	}
	
	// Plain text fallback - only if no structured format detected
	if len(replyParts) == 0 {
		logrus.Warning("⚠️ AI_PROCESSOR: Plain text response detected. Using fallback handling.")
		if stage == "" {
			stage = "Problem Identification"
		}
		replyParts = []AIResponsePart{
			{Type: "text", Content: strings.TrimSpace(replyContent)},
		}
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
				// Clean the URL
				mediaURL := strings.TrimSpace(part.Content)
				// Remove URL encoding if present
				if strings.Contains(mediaURL, "%") {
					// Simple URL decode - you might want to use net/url package for proper decoding
					mediaURL = strings.ReplaceAll(mediaURL, "%20", " ")
					mediaURL = strings.ReplaceAll(mediaURL, "%2F", "/")
					mediaURL = strings.ReplaceAll(mediaURL, "%3A", ":")
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
