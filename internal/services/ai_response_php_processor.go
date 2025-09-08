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
	// Remove markdown code blocks if present
	sanitizedContent := regexp.MustCompile(`^` + "```" + `json|` + "```" + `$`).ReplaceAllString(strings.TrimSpace(replyContent), "")
	
	var data AIResponseData
	var replyParts []AIResponsePart
	
	// Try to decode JSON directly
	if err := json.Unmarshal([]byte(sanitizedContent), &data); err == nil {
		if data.Stage != "" && len(data.Response) > 0 {
			stage = data.Stage
			replyParts = data.Response
			logrus.Info("✅ Parsed standard JSON format")
		}
	}
	
	// If not parsed yet, try Stage: Response: format
	if len(replyParts) == 0 {
		stageResponsePattern := regexp.MustCompile(`Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$`)
		if matches := stageResponsePattern.FindStringSubmatch(replyContent); len(matches) == 3 {
			stage = strings.TrimSpace(matches[1])
			responseJSON := matches[2]
			if err := json.Unmarshal([]byte(responseJSON), &replyParts); err == nil {
				logrus.Info("✅ Parsed Stage: Response: format")
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
					logrus.Info("✅ Parsed clean JSON format")
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
					logrus.Info("✅ Parsed encapsulated JSON format")
				}
			}
		}
	}
	
	// Plain text fallback
	if len(replyParts) == 0 {
		logrus.Warning("Plain text response detected. Using fallback handling.")
		if stage == "" {
			stage = "Problem Identification"
		}
		replyParts = []AIResponsePart{
			{Type: "text", Content: strings.TrimSpace(replyContent)},
		}
	}
	
	// Process reply parts exactly like PHP
	textParts := []string{}
	isOnemessageActive := false
	delay := time.Duration(delayMs) * time.Millisecond
	
	for index, part := range replyParts {
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
				}).Info("✅ Combined onemessage parts")
				
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
	
	return stage, messages, nil
}
