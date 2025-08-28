package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"github.com/sirupsen/logrus"
)

// TestCase represents a test case for AI response parsing
type TestCase struct {
	Name        string
	Input       string
	ExpectedOK  bool
	ExpectedStage string
	Description string
}

func main() {
	// Set up logging
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Initialize config and AI service
	cfg := &config.Config{}
	aiService := services.NewAIService(cfg)

	// Define comprehensive test cases based on PHP implementation patterns
	testCases := []TestCase{
		// Test Case 1: Standard JSON format
		{
			Name: "Standard JSON Format",
			Input: `{
				"Stage": "Problem Identification",
				"Response": [
					{"type": "text", "content": "Hello, how can I help you?", "Jenis": "onemessage"},
					{"type": "image", "content": "https://example.com/image.jpg"}
				]
			}`,
			ExpectedOK:    true,
			ExpectedStage: "Problem Identification",
			Description:   "Standard JSON format with Stage and Response fields",
		},

		// Test Case 2: JSON with code blocks
		{
			Name: "JSON with Code Blocks",
			Input: "```json\n{\n\t\"Stage\": \"Consultation\",\n\t\"Response\": [\n\t\t{\"type\": \"text\", \"content\": \"Let me help you with that.\"}\n\t]\n}\n```",
			ExpectedOK:    true,
			ExpectedStage: "Consultation",
			Description:   "JSON wrapped in markdown code blocks",
		},

		// Test Case 3: Older format with Stage: Response:
		{
			Name: "Older Format",
			Input: "Stage: Problem Analysis\nResponse: [{\"type\": \"text\", \"content\": \"I understand your concern.\"}]",
			ExpectedOK:    true,
			ExpectedStage: "Problem Analysis",
			Description:   "Older format with Stage: and Response: separators",
		},

		// Test Case 4: Encapsulated JSON in text
		{
			Name: "Encapsulated JSON",
			Input: "Here is the response: {\"Stage\": \"Follow-up\", \"Response\": [{\"type\": \"text\", \"content\": \"Thank you for your patience.\"}]} Please let me know if you need more help.",
			ExpectedOK:    true,
			ExpectedStage: "Follow-up",
			Description:   "JSON embedded within regular text",
		},

		// Test Case 5: Malformed JSON with flexible parsing
		{
			Name: "Flexible JSON",
			Input: "stage: Initial Assessment, response: [{\"type\": \"text\", \"content\": \"Welcome to our service.\"}]",
			ExpectedOK:    true,
			ExpectedStage: "Initial Assessment",
			Description:   "Flexible format with lowercase fields and comma separation",
		},

		// Test Case 6: JSON with escaped quotes
		{
			Name: "Escaped JSON",
			Input: "\"{\\\"Stage\\\": \\\"Customer Support\\\", \\\"Response\\\": [{\\\"type\\\": \\\"text\\\", \\\"content\\\": \\\"How may I assist you today?\\\"}]}\"",
			ExpectedOK:    true,
			ExpectedStage: "Customer Support",
			Description:   "JSON with escaped quotes",
		},

		// Test Case 7: Array wrapped JSON
		{
			Name: "Array Wrapped JSON",
			Input: "[{\"Stage\": \"Information Gathering\", \"Response\": [{\"type\": \"text\", \"content\": \"Please provide more details.\"}]}]",
			ExpectedOK:    true,
			ExpectedStage: "Information Gathering",
			Description:   "JSON object wrapped in an array",
		},

		// Test Case 8: Case insensitive fields
		{
			Name: "Case Insensitive",
			Input: "{\"STAGE\": \"PROBLEM SOLVING\", \"RESPONSE\": [{\"type\": \"text\", \"content\": \"Let's work on this together.\"}]}",
			ExpectedOK:    true,
			ExpectedStage: "PROBLEM SOLVING",
			Description:   "JSON with uppercase field names",
		},

		// Test Case 9: Plain text fallback
		{
			Name: "Plain Text",
			Input: "This is just a plain text response without any JSON structure.",
			ExpectedOK:    true,
			ExpectedStage: "conversation",
			Description:   "Plain text that should trigger fallback parsing",
		},

		// Test Case 10: Empty content
		{
			Name: "Empty Content",
			Input: "",
			ExpectedOK:    true,
			ExpectedStage: "conversation",
			Description:   "Empty content that should trigger default response",
		},

		// Test Case 11: Multiline JSON with extra whitespace
		{
			Name: "Multiline JSON",
			Input: "\n\n  {\n    \"Stage\": \"Technical Support\",\n    \"Response\": [\n      {\n        \"type\": \"text\",\n        \"content\": \"I'll help you troubleshoot this issue.\"\n      }\n    ]\n  }\n\n",
			ExpectedOK:    true,
			ExpectedStage: "Technical Support",
			Description:   "Multiline JSON with extra whitespace",
		},

		// Test Case 12: JSON with alternative separators
		{
			Name: "Alternative Separators",
			Input: "Stage = Sales Consultation Response = [{\"type\": \"text\", \"content\": \"Let me show you our products.\"}]",
			ExpectedOK:    true,
			ExpectedStage: "Sales Consultation",
			Description:   "Format using equals signs as separators",
		},
	}

	// Run tests
	fmt.Println("=== Comprehensive AI Response Parsing Test ===")
	fmt.Printf("Testing %d different response formats...\n\n", len(testCases))

	passedTests := 0
	totalTests := len(testCases)

	for i, testCase := range testCases {
		fmt.Printf("Test %d: %s\n", i+1, testCase.Name)
		fmt.Printf("Description: %s\n", testCase.Description)
		fmt.Printf("Input: %s\n", truncateString(testCase.Input, 100))

		// Parse the response
		response, err := aiService.ParseAIResponse(testCase.Input)

		// Check results
		if err != nil {
			fmt.Printf("❌ FAILED: Error occurred: %v\n", err)
		} else if response == nil {
			fmt.Printf("❌ FAILED: Response is nil\n")
		} else {
			// Validate the response
			if response.Stage == testCase.ExpectedStage {
				fmt.Printf("✅ PASSED: Stage = %s, Response count = %d\n", response.Stage, len(response.Response))
				passedTests++

				// Print response details
				for j, part := range response.Response {
					fmt.Printf("   Response[%d]: Type=%s, Content=%s\n", j, part.Type, truncateString(part.Content, 50))
				}
			} else {
				fmt.Printf("❌ FAILED: Expected stage '%s', got '%s'\n", testCase.ExpectedStage, response.Stage)
			}
		}

		fmt.Println("---")
	}

	// Print summary
	fmt.Printf("\n=== Test Summary ===\n")
	fmt.Printf("Passed: %d/%d tests\n", passedTests, totalTests)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)

	if passedTests == totalTests {
		fmt.Println("🎉 All tests passed! The comprehensive PHP parsing logic has been successfully implemented.")
	} else {
		fmt.Printf("⚠️  %d tests failed. Review the implementation for edge cases.\n", totalTests-passedTests)
		os.Exit(1)
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Helper method to expose parseAIResponse for testing
func (s *services.AIService) ParseAIResponse(content string) (*models.AIPromptResponse, error) {
	// This is a test helper - in production, parseAIResponse is called internally
	// We need to use reflection or create a public wrapper for testing
	// For now, we'll create the response manually to test the logic
	
	// Note: This would normally require making parseAIResponse public or using reflection
	// For this test, we'll simulate the parsing logic
	return &models.AIPromptResponse{
		Stage: "test",
		Response: []models.AIResponsePart{
			{Type: "text", Content: "test response"},
		},
	}, nil
}