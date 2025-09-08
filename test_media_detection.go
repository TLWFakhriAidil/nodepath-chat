package main

import (
	"fmt"
	"nodepath-chat/internal/services"
)

func main() {
	// Create media detection service
	mediaService := services.NewMediaDetectionService()
	
	// Test text that simulates AI response
	testText := `Alhamdulillah, Layla kongsikan beberapa testimoni ibu-ibu lain yang anaknya dah dapat manfaat dari Vitac!
Masalah: Kerap Demam
Gambar 1: [https://chatbot.growrvsb.com/public/images/chatgpt/23141741665515]
Gambar 2: [https://chatbot.growrvsb.com/public/images/chatgpt/23141741665523]
Gambar 3: [https://chatbot.growrvsb.com/public/images/chatgpt/23141741665533]
Alhamdulillah, akak ni dah settle masalah anak dia yang kerap demam dengan Vitac. Akak pun nak anak lebih sihat dan aktif kan?
Layla ada maklumkan juga, Vitac diluluskan oleh KKM di bawah Perkelasan Pemakanan (MESTI), selamat dan berkesan. Akak nak tahu lebih lanjut pasal promosi ke?`

	fmt.Println("Testing media detection for AI response format...")
	fmt.Println("================================================================================")
	fmt.Println("Original text:")
	fmt.Println(testText)
	fmt.Println("================================================================================")
	
	// Check if media is detected
	hasMedia := mediaService.HasMedia(testText)
	fmt.Printf("\nHas media: %v\n", hasMedia)
	
	// Extract all media
	allMedia := mediaService.ExtractAllMedia(testText)
	fmt.Printf("\nFound %d media items:\n", len(allMedia))
	for i, media := range allMedia {
		fmt.Printf("  %d. Type: %s, URL: %s\n", i+1, media.MediaType, media.MediaURL)
		fmt.Printf("     Original: %s\n", media.OriginalText)
	}
	
	// Get clean text
	cleanText := mediaService.RemoveMediaURLs(testText)
	fmt.Println("\nClean text (without media URLs):")
	fmt.Println(cleanText)
	
	// Test other formats
	fmt.Println("\n================================================================================")
	fmt.Println("Testing other formats:")
	
	testFormats := []string{
		"Check this image: [IMAGE: https://example.com/test.jpg]",
		"Simple bracket: [https://example.com/test.png]",
		"Image 1: [https://example.com/test.jpeg]",
		"Foto 2: [https://example.com/test.gif]",
		"Picture 3: [https://example.com/test.webp]",
	}
	
	for _, test := range testFormats {
		fmt.Printf("\nTest: %s\n", test)
		if mediaService.HasMedia(test) {
			media := mediaService.ExtractFirstMedia(test)
			fmt.Printf("  ✓ Detected as %s: %s\n", media.MediaType, media.MediaURL)
		} else {
			fmt.Println("  ✗ Not detected as media")
		}
	}
}
