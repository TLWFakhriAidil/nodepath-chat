package main

import (
	"fmt"
	"nodepath-chat/internal/services"
)

// Test media detection with the URLs from user's screenshot
func main() {
	// Initialize media detection service
	mediaService := services.NewMediaDetectionService()
	
	// Test URLs from user's screenshot
	testURLs := []string{
		"https://chatbot.growrvsb.com/public/images/chatgpt/22811753944188.png",
		"https://chatbot.growrvsb.com/public/images/chatgpt/22811753234074.mp3",
		"https://chatbot.growrvsb.com/public/images/chatgpt/22811753234054.mp4",
	}
	
	for i, url := range testURLs {
		fmt.Printf("\n=== Testing URL %d ===\n", i+1)
		fmt.Printf("URL: %s\n", url)
		
		// Test HasMedia
		hasMedia := mediaService.HasMedia(url)
		fmt.Printf("HasMedia: %t\n", hasMedia)
		
		// Test ExtractFirstMedia
		mediaInfo := mediaService.ExtractFirstMedia(url)
		if mediaInfo != nil {
			fmt.Printf("Media Type: %s\n", mediaInfo.MediaType)
			fmt.Printf("Media URL: %s\n", mediaInfo.MediaURL)
			fmt.Printf("Is Media: %t\n", mediaInfo.IsMedia)
		} else {
			fmt.Printf("No media detected\n")
		}
		
		// Test DetectMedia for full results
		results := mediaService.DetectMedia(url)
		fmt.Printf("Detection Results Count: %d\n", len(results))
		for j, result := range results {
			fmt.Printf("  Result %d: Type=%s, URL=%s, IsMedia=%t\n", j+1, result.MediaType, result.MediaURL, result.IsMedia)
		}
	}
	
	// Test combined message like in the screenshot
	fmt.Printf("\n=== Testing Combined Message ===\n")
	combinedMessage := `Hai, Assalamualaikum, Saya Fakhri daripada exama hq
https://chatbot.growrvsb.com/public/images/chatgpt/22811753944188.png
https://chatbot.growrvsb.com/public/images/chatgpt/22811753234074.mp3
https://chatbot.growrvsb.com/public/images/chatgpt/22811753234054.mp4
Ni untuk anak atau sendiri
Ohh, anak rupanya...kejpa nyea`
	
	fmt.Printf("Combined Message:\n%s\n\n", combinedMessage)
	hasMedia := mediaService.HasMedia(combinedMessage)
	fmt.Printf("HasMedia: %t\n", hasMedia)
	
	mediaInfo := mediaService.ExtractFirstMedia(combinedMessage)
	if mediaInfo != nil {
		fmt.Printf("First Media Type: %s\n", mediaInfo.MediaType)
		fmt.Printf("First Media URL: %s\n", mediaInfo.MediaURL)
		fmt.Printf("Clean Text: %s\n", mediaInfo.CleanText)
	} else {
		fmt.Printf("No media detected in combined message\n")
	}
	
	results := mediaService.DetectMedia(combinedMessage)
	fmt.Printf("Total Detection Results: %d\n", len(results))
	for i, result := range results {
		fmt.Printf("  Media %d: Type=%s, URL=%s\n", i+1, result.MediaType, result.MediaURL)
	}
}