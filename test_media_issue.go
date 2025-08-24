package main

import (
	"fmt"
	"nodepath-chat/internal/services"
)

// Test function to debug media URL processing issue
func main() {
	fmt.Println("Testing Media URL Processing Issue...")

	// Create media detection service
	mediaService := services.NewMediaDetectionService()

	// Test URLs from the user's screenshot
	testURLs := []string{
		"https://chatbot.growrysb.com/public/images/chatgpt/22811753944188.png",
		"https://chatbot.growrysb.com/public/images/chatgpt/22811753234074.mp3",
		"https://chatbot.growrysb.com/public/images/chatgpt/22811753234054.mp4",
	}

	fmt.Println("\n=== Testing Media Detection ===")
	for i, url := range testURLs {
		fmt.Printf("\nTest %d: %s\n", i+1, url)
		fmt.Printf("HasMedia: %v\n", mediaService.HasMedia(url))
		
		if mediaService.HasMedia(url) {
			mediaInfo := mediaService.ExtractFirstMedia(url)
			if mediaInfo != nil {
				fmt.Printf("MediaType: %s\n", mediaInfo.MediaType)
				fmt.Printf("MediaURL: %s\n", mediaInfo.MediaURL)
			} else {
				fmt.Println("Failed to extract media info")
			}
		}
	}

	// Test AI response parsing simulation
	fmt.Println("\n=== Testing AI Response Parsing ===")
	testAIResponse := `{
		"Stage": "Problem Identification",
		"Response": [
			{"type": "text", "content": "Hai, Assalamualaikum, Saya Fakhri daripada exama hq"},
			{"type": "image", "content": "https://chatbot.growrysb.com/public/images/chatgpt/22811753944188.png"},
			{"type": "audio", "content": "https://chatbot.growrysb.com/public/images/chatgpt/22811753234074.mp3"},
			{"type": "video", "content": "https://chatbot.growrysb.com/public/images/chatgpt/22811753234054.mp4"},
			{"type": "text", "content": "Ni untuk anak atau sendiri"}
		]
	}`

	fmt.Printf("Sample AI Response:\n%s\n", testAIResponse)

	// Test text with embedded URLs
	fmt.Println("\n=== Testing Text with Embedded URLs ===")
	textWithMedia := "Hai, Assalamualaikum, Saya Fakhri daripada exama hq https://chatbot.growrysb.com/public/images/chatgpt/22811753944188.png"
	fmt.Printf("Text: %s\n", textWithMedia)
	fmt.Printf("HasMedia: %v\n", mediaService.HasMedia(textWithMedia))
	
	if mediaService.HasMedia(textWithMedia) {
		mediaInfo := mediaService.ExtractFirstMedia(textWithMedia)
		if mediaInfo != nil {
			fmt.Printf("Extracted MediaType: %s\n", mediaInfo.MediaType)
			fmt.Printf("Extracted MediaURL: %s\n", mediaInfo.MediaURL)
		} else {
			fmt.Println("Failed to extract media from text")
		}
	}

	fmt.Println("\n=== Test Complete ===")
}