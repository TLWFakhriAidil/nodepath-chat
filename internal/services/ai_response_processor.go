package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ResponseItem represents a single response item (text, image, audio, video)
type ResponseItem struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Jenis   string `json:"Jenis,omitempty"` // For "onemessage" combining
}

// AIResponse represents the structured AI response
type AIResponse struct {
	Stage    string         `json:"Stage"`
	Response []ResponseItem `json:"Response"`
}

// ProcessedMessage represents a message ready to be sent
type ProcessedMessage struct {
	Type    string // "text", "image", "audio", "video"
	Content string
	Delay   time.Duration
}

// AIResponseProcessor handles AI response processing according to PHP logic
type AIResponseProcessor struct {
	defaultDelay time.Duration
}

// NewAIResponseProcessor creates a new AI response processor
func NewAIResponseProcessor(defaultDelay time.Duration) *AIResponseProcessor {
	return &AIResponseProcessor{
		defaultDelay: defaultDelay,
	}
}

// ProcessAIResponse processes AI response according to PHP logic
// Supports 5 formats as per PHP implementation:
// 1. Standard JSON format with Stage and Response array
// 2. JSON with encapsulated content in triple backticks
// 3. Old format with Stage: and Response: text
// 4. Plain text fallback
// 5. Nested JSON within response content
func (p *AIResponseProcessor) ProcessAIResponse(rawResponse string, updateConversation func(stage string, messages []ProcessedMessage)) ([]ProcessedMessage, error) {
	logrus.WithField("raw_response_length", len(rawResponse)).Info("🔍 AI_PROCESSOR: Starting response processing")

	// Clean and prepare response
	sanitizedContent := p.sanitizeResponse(rawResponse)

	// Try to parse the response
	aiResponse, err := p.parseResponse(sanitizedContent, rawResponse)
	if err != nil {
		logrus.WithError(err).Error("❌ AI_PROCESSOR: Failed to parse AI response")
		return nil, err
	}

	// Process response items according to PHP logic
	processedMessages := p.processResponseItems(aiResponse.Response)

	// Update conversation with stage and processed messages
	if updateConversation != nil {
		updateConversation(aiResponse.Stage, processedMessages)
	}

	logrus.WithFields(logrus.Fields{
		"stage":          aiResponse.Stage,
		"message_count":  len(processedMessages),
		"original_items": len(aiResponse.Response),
	}).Info("✅ AI_PROCESSOR: Response processing completed")

	return processedMessages, nil
}

// sanitizeResponse provides adaptive cleaning that preserves user-intended content
// while removing formatting issues that interfere with JSON parsing
func (p *AIResponseProcessor) sanitizeResponse(rawResponse string) string {
	sanitized := strings.TrimSpace(rawResponse)

	// Step 1: Preserve user-intended content by identifying and protecting it
	protectedContent := p.identifyProtectedContent(sanitized)

	// Step 2: Clean markdown code blocks (but preserve content structure)
	sanitized = p.cleanMarkdownBlocks(sanitized)

	// Step 3: Adaptive URL cleaning - preserve intentional formatting
	sanitized = p.adaptiveURLCleaning(sanitized, protectedContent)

	// Step 4: Clean formatting artifacts while preserving content meaning
	sanitized = p.cleanFormattingArtifacts(sanitized)

	// Step 5: Restore protected content if needed
	sanitized = p.restoreProtectedContent(sanitized, protectedContent)

	return strings.TrimSpace(sanitized)
}

// identifyProtectedContent identifies content that should be preserved during sanitization
func (p *AIResponseProcessor) identifyProtectedContent(content string) map[string]string {
	protected := make(map[string]string)

	// Protect quoted strings that might contain intentional formatting
	quotedPattern := regexp.MustCompile(`"([^"]*[!` + "`" + `][^"]*)"`)
	matches := quotedPattern.FindAllStringSubmatch(content, -1)
	for i, match := range matches {
		if len(match) > 1 {
			placeholder := fmt.Sprintf("__PROTECTED_QUOTE_%d__", i)
			protected[placeholder] = match[1]
		}
	}

	// Protect content within JSON strings that might have intentional markdown
	contentPattern := regexp.MustCompile(`"content"\s*:\s*"([^"]*)"`)
	matches = contentPattern.FindAllStringSubmatch(content, -1)
	for i, match := range matches {
		if len(match) > 1 && strings.Contains(match[1], "!") {
			placeholder := fmt.Sprintf("__PROTECTED_CONTENT_%d__", i)
			protected[placeholder] = match[1]
		}
	}

	return protected
}

// cleanMarkdownBlocks removes markdown code block markers while preserving content
func (p *AIResponseProcessor) cleanMarkdownBlocks(content string) string {
	// Remove markdown code block markers but preserve the JSON content
	codeBlockPattern := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)\\s*```")
	if codeBlockPattern.MatchString(content) {
		// Extract content from code blocks
		content = codeBlockPattern.ReplaceAllString(content, "$1")
	}

	// Clean up standalone markdown markers
	content = regexp.MustCompile(`^`+"```"+`json|`+"```"+`$`).ReplaceAllString(content, "")

	return content
}

// adaptiveURLCleaning cleans URLs while preserving intentional formatting
func (p *AIResponseProcessor) adaptiveURLCleaning(content string, protected map[string]string) string {
	// Only clean URLs that are clearly formatting artifacts, not intentional content

	// Clean AI-generated markdown image URLs: ! `URL` -> URL (but only outside protected content)
	// This handles cases where AI generates: Gambar 1: ! `https://example.com/image.jpg`
	aiImagePattern := regexp.MustCompile(`!\s*` + "`" + `(https?://[^` + "`" + `\s]+)` + "`")
	content = aiImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		// Check if this URL is in protected content
		for _, protectedValue := range protected {
			if strings.Contains(protectedValue, match) {
				return match // Don't clean protected content
			}
		}
		// Clean the URL
		return aiImagePattern.ReplaceAllString(match, "$1")
	})

	// Clean backticks around URLs only if they appear to be formatting artifacts
	backtickPattern := regexp.MustCompile("`" + `(https?://[^` + "`" + `\s]+)` + "`")
	content = backtickPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Check if this is in a JSON string context (likely intentional)
		if regexp.MustCompile(`"[^"]*` + regexp.QuoteMeta(match) + `[^"]*"`).MatchString(content) {
			return match // Preserve URLs in JSON strings
		}
		// Clean standalone backticked URLs
		return backtickPattern.ReplaceAllString(match, "$1")
	})

	return content
}

// cleanFormattingArtifacts removes formatting artifacts while preserving content meaning
func (p *AIResponseProcessor) cleanFormattingArtifacts(content string) string {
	// Remove extra whitespace and line breaks that interfere with JSON parsing
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// Clean up common AI response artifacts
	artifacts := []struct {
		pattern string
		replace string
	}{
		{`\s*,\s*,`, ","},      // Double commas
		{`\s*}\s*}`, "}"},      // Double closing braces
		{`\s*]\s*]`, "]"},      // Double closing brackets
		{`"\s*,\s*"`, "\",\""}, // Spaced commas in strings
	}

	for _, artifact := range artifacts {
		content = regexp.MustCompile(artifact.pattern).ReplaceAllString(content, artifact.replace)
	}

	return content
}

// restoreProtectedContent restores content that was protected during sanitization
func (p *AIResponseProcessor) restoreProtectedContent(content string, protected map[string]string) string {
	for placeholder, originalContent := range protected {
		content = strings.ReplaceAll(content, placeholder, originalContent)
	}
	return content
}

// parseResponse attempts to parse AI response using multiple formats
func (p *AIResponseProcessor) parseResponse(sanitizedContent, rawResponse string) (*AIResponse, error) {
	var aiResponse AIResponse
	var replyParts []ResponseItem
	stage := ""

	// Format 1: Standard JSON format
	if err := json.Unmarshal([]byte(sanitizedContent), &aiResponse); err == nil {
		if aiResponse.Stage != "" && len(aiResponse.Response) > 0 {
			logrus.Info("✅ AI_PROCESSOR: Parsed Format 1 - Standard JSON")
			return &aiResponse, nil
		}
	}

	// Format 2: Check if response has encapsulated JSON in first item
	if len(aiResponse.Response) > 0 && aiResponse.Response[0].Type == "text" {
		content := aiResponse.Response[0].Content
		if regexp.MustCompile(`^` + "```" + `json.*` + "```" + `$`).MatchString(content) {
			jsonContent := regexp.MustCompile(`^`+"```"+`json|`+"```"+`$`).ReplaceAllString(strings.TrimSpace(content), "")

			var decodedContent AIResponse
			if err := json.Unmarshal([]byte(jsonContent), &decodedContent); err == nil {
				if decodedContent.Stage != "" && len(decodedContent.Response) > 0 {
					logrus.Info("✅ AI_PROCESSOR: Parsed Format 2 - Encapsulated JSON")
					return &decodedContent, nil
				}
			}
		}
	}

	// Format 3: Old format with Stage: and Response: pattern
	if matches := regexp.MustCompile(`Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$`).FindStringSubmatch(rawResponse); len(matches) == 3 {
		stage = strings.TrimSpace(matches[1])
		responseJSON := matches[2]

		if err := json.Unmarshal([]byte(responseJSON), &replyParts); err == nil {
			logrus.Info("✅ AI_PROCESSOR: Parsed Format 3 - Old Stage/Response format")
			return &AIResponse{
				Stage:    stage,
				Response: replyParts,
			}, nil
		}
	}

	// Format 4: Try parsing raw response as JSON one more time
	if err := json.Unmarshal([]byte(rawResponse), &aiResponse); err == nil {
		if aiResponse.Stage != "" && len(aiResponse.Response) > 0 {
			logrus.Info("✅ AI_PROCESSOR: Parsed Format 4 - Raw JSON")
			return &aiResponse, nil
		}
	}

	// Format 5: Plain text fallback
	logrus.Warning("⚠️ AI_PROCESSOR: Using Format 5 - Plain text fallback")
	return &AIResponse{
		Stage: "Problem Identification", // Default stage
		Response: []ResponseItem{
			{Type: "text", Content: strings.TrimSpace(rawResponse)},
		},
	}, nil
}

// processResponseItems processes response items according to PHP logic
// Implements the exact "onemessage" combining logic from PHP
func (p *AIResponseProcessor) processResponseItems(items []ResponseItem) []ProcessedMessage {
	var processedMessages []ProcessedMessage
	var textParts []string
	isOnemessageActive := false

	logrus.WithField("item_count", len(items)).Info("🔧 AI_PROCESSOR: Processing response items")

	for index, part := range items {
		// Validate item structure
		if part.Type == "" || part.Content == "" {
			logrus.WithField("item", part).Warning("⚠️ AI_PROCESSOR: Invalid response part structure")
			continue
		}

		// Check if this is a text with "Jenis": "onemessage"
		if part.Type == "text" && part.Jenis == "onemessage" {
			// Start or continue collecting text parts
			textParts = append(textParts, part.Content)
			isOnemessageActive = true

			logrus.WithFields(logrus.Fields{
				"index":       index,
				"parts_count": len(textParts),
			}).Debug("📝 AI_PROCESSOR: Collecting onemessage part")

			// Check if next part is also onemessage
			isLastPart := index == len(items)-1
			nextIsNotOnemessage := false

			if !isLastPart {
				nextPart := items[index+1]
				nextIsNotOnemessage = nextPart.Type != "text" || nextPart.Jenis != "onemessage"
			}

			// If this is the last part OR next part is not onemessage, send combined
			if isLastPart || nextIsNotOnemessage {
				combinedMessage := strings.Join(textParts, "\n")
				processedMessages = append(processedMessages, ProcessedMessage{
					Type:    "text",
					Content: combinedMessage,
					Delay:   p.defaultDelay,
				})

				logrus.WithFields(logrus.Fields{
					"combined_parts": len(textParts),
					"message_length": len(combinedMessage),
				}).Info("✅ AI_PROCESSOR: Combined onemessage parts")

				// Reset for next group
				textParts = []string{}
				isOnemessageActive = false
			}
		} else {
			// If we were collecting onemessage parts, flush them first
			if isOnemessageActive && len(textParts) > 0 {
				combinedMessage := strings.Join(textParts, "\n")
				processedMessages = append(processedMessages, ProcessedMessage{
					Type:    "text",
					Content: combinedMessage,
					Delay:   p.defaultDelay,
				})

				logrus.WithFields(logrus.Fields{
					"combined_parts": len(textParts),
					"message_length": len(combinedMessage),
				}).Info("✅ AI_PROCESSOR: Flushed onemessage parts before non-onemessage item")

				textParts = []string{}
				isOnemessageActive = false
			}

			// Process normal item (text without onemessage, image, audio, video)
			processedMessages = append(processedMessages, ProcessedMessage{
				Type:    part.Type,
				Content: p.processContent(part.Type, part.Content),
				Delay:   p.defaultDelay,
			})

			logrus.WithFields(logrus.Fields{
				"type":    part.Type,
				"content": truncateForLog(part.Content, 100),
			}).Debug("📤 AI_PROCESSOR: Added regular message")
		}
	}

	// Handle any remaining onemessage parts (shouldn't happen but just in case)
	if isOnemessageActive && len(textParts) > 0 {
		combinedMessage := strings.Join(textParts, "\n")
		processedMessages = append(processedMessages, ProcessedMessage{
			Type:    "text",
			Content: combinedMessage,
			Delay:   p.defaultDelay,
		})

		logrus.WithField("parts", len(textParts)).Warning("⚠️ AI_PROCESSOR: Flushed remaining onemessage parts at end")
	}

	return processedMessages
}

// processContent processes content based on type
func (p *AIResponseProcessor) processContent(contentType, content string) string {
	switch contentType {
	case "image", "audio", "video":
		// Extract URL from bracket format if present
		if matches := regexp.MustCompile(`\[(IMAGE|AUDIO|VIDEO):\s*(.+?)\]`).FindStringSubmatch(content); len(matches) == 3 {
			url := strings.TrimSpace(matches[2])
			// Remove backticks if present
			url = strings.Trim(url, "`")
			logrus.WithFields(logrus.Fields{
				"original":  content,
				"extracted": url,
			}).Debug("🔗 AI_PROCESSOR: Extracted media URL from bracket format")
			return url
		}
		// Return as-is if already a URL
		return strings.TrimSpace(content)
	default:
		return content
	}
}

// truncateForLog truncates string for logging
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// FormatResponseForLogging formats processed messages for conversation logging
func (p *AIResponseProcessor) FormatResponseForLogging(messages []ProcessedMessage, logType string) []string {
	var logEntries []string

	for _, msg := range messages {
		var entry string

		switch msg.Type {
		case "text":
			// Check if this was a combined message by looking for newlines
			contentJSON, _ := json.Marshal(msg.Content)
			if strings.Contains(msg.Content, "\n") {
				entry = fmt.Sprintf("%s_COMBINED: %s", logType, string(contentJSON))
			} else {
				entry = fmt.Sprintf("%s: %s", logType, string(contentJSON))
			}
		case "image", "audio", "video":
			entry = fmt.Sprintf("%s: %s", logType, msg.Content)
		default:
			contentJSON, _ := json.Marshal(msg.Content)
			entry = fmt.Sprintf("%s: %s", logType, string(contentJSON))
		}

		logEntries = append(logEntries, entry)
	}

	return logEntries
}

// ValidateAIResponse validates that the AI response has required fields
func (p *AIResponseProcessor) ValidateAIResponse(response *AIResponse) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	if response.Stage == "" {
		return fmt.Errorf("stage is empty")
	}

	if len(response.Response) == 0 {
		return fmt.Errorf("response array is empty")
	}

	// Validate each response item
	for i, item := range response.Response {
		if item.Type == "" {
			return fmt.Errorf("response item %d has empty type", i)
		}
		if item.Content == "" {
			return fmt.Errorf("response item %d has empty content", i)
		}
	}

	return nil
}
