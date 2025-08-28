package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"nodepath-chat/internal/services"
)

// TestOnemessageCombining tests the PHP onemessage combining logic implementation
// This test verifies that consecutive text parts with Jenis="onemessage" are properly combined
// and sent as a single message, matching the PHP implementation behavior
func main() {
	fmt.Println("🧪 Testing Onemessage Combining Logic")
	fmt.Println("=====================================\n")

	// Test Case 1: Multiple consecutive onemessage text parts
	testCase1()

	// Test Case 2: Mixed onemessage and regular text
	testCase2()

	// Test Case 3: Onemessage followed by image
	testCase3()

	// Test Case 4: Single onemessage text
	testCase4()

	// Test Case 5: No onemessage parts (regular response)
	testCase5()

	// Test Case 6: Onemessage at the end of response
	testCase6()

	fmt.Println("\n✅ All test cases completed!")
}

// testCase1 tests multiple consecutive onemessage text parts
func testCase1() {
	fmt.Println("📋 Test Case 1: Multiple consecutive onemessage text parts")
	fmt.Println("Expected: All onemessage parts combined into single message")

	response := &services.AIWhatsappResponse{
		Stage: "Problem Identification",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Kurang selera makan, sembelit, atau kerap demam?",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Tolong bagitahu Layla ya kak.",
			},
		},
	}

	printTestResponse(response)
	expectedCombined := "Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\nKurang selera makan, sembelit, atau kerap demam?\nTolong bagitahu Layla ya kak."
	fmt.Printf("Expected Combined Message: %s\n\n", expectedCombined)
}

// testCase2 tests mixed onemessage and regular text
func testCase2() {
	fmt.Println("📋 Test Case 2: Mixed onemessage and regular text")
	fmt.Println("Expected: Onemessage parts combined, then regular text sent separately")

	response := &services.AIWhatsappResponse{
		Stage: "Information Gathering",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Terima kasih atas maklumatnya.",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Layla faham situasi anak akak sekarang.",
			},
			{
				Type:    "text",
				Content: "Boleh akak ceritakan lebih detail tentang gejala yang dialami?",
			},
		},
	}

	printTestResponse(response)
	expectedCombined := "Terima kasih atas maklumatnya.\nLayla faham situasi anak akak sekarang."
	expectedSeparate := "Boleh akak ceritakan lebih detail tentang gejala yang dialami?"
	fmt.Printf("Expected Combined Message: %s\n", expectedCombined)
	fmt.Printf("Expected Separate Message: %s\n\n", expectedSeparate)
}

// testCase3 tests onemessage followed by image
func testCase3() {
	fmt.Println("📋 Test Case 3: Onemessage followed by image")
	fmt.Println("Expected: Onemessage parts combined, then image sent separately")

	response := &services.AIWhatsappResponse{
		Stage: "Solution Presentation",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Berdasarkan maklumat yang akak berikan,",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Layla cadangkan produk ini untuk anak akak:",
			},
			{
				Type:    "image",
				Content: "https://example.com/product-image.jpg",
			},
		},
	}

	printTestResponse(response)
	expectedCombined := "Berdasarkan maklumat yang akak berikan,\nLayla cadangkan produk ini untuk anak akak:"
	expectedImage := "https://example.com/product-image.jpg"
	fmt.Printf("Expected Combined Message: %s\n", expectedCombined)
	fmt.Printf("Expected Image URL: %s\n\n", expectedImage)
}

// testCase4 tests single onemessage text
func testCase4() {
	fmt.Println("📋 Test Case 4: Single onemessage text")
	fmt.Println("Expected: Single onemessage sent as combined message")

	response := &services.AIWhatsappResponse{
		Stage: "Greeting",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Selamat datang! Saya Layla, pembantu virtual anda.",
			},
		},
	}

	printTestResponse(response)
	expectedCombined := "Selamat datang! Saya Layla, pembantu virtual anda."
	fmt.Printf("Expected Combined Message: %s\n\n", expectedCombined)
}

// testCase5 tests no onemessage parts (regular response)
func testCase5() {
	fmt.Println("📋 Test Case 5: No onemessage parts (regular response)")
	fmt.Println("Expected: All messages sent separately as regular BOT messages")

	response := &services.AIWhatsappResponse{
		Stage: "Regular Chat",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Content: "Bagaimana keadaan anak akak hari ini?",
			},
			{
				Type:    "text",
				Content: "Ada sebarang perubahan yang akak perasan?",
			},
		},
	}

	printTestResponse(response)
	fmt.Printf("Expected: Two separate BOT messages\n\n")
}

// testCase6 tests onemessage at the end of response
func testCase6() {
	fmt.Println("📋 Test Case 6: Onemessage at the end of response")
	fmt.Println("Expected: Regular text first, then onemessage parts combined")

	response := &services.AIWhatsappResponse{
		Stage: "Closing",
		Response: []services.AIWhatsappResponseItem{
			{
				Type:    "text",
				Content: "Terima kasih kerana berkongsi maklumat dengan Layla.",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "Jika ada sebarang soalan lagi,",
			},
			{
				Type:    "text",
				Jenis:   "onemessage",
				Content: "jangan segan untuk hubungi Layla ya!",
			},
		},
	}

	printTestResponse(response)
	expectedSeparate := "Terima kasih kerana berkongsi maklumat dengan Layla."
	expectedCombined := "Jika ada sebarang soalan lagi,\njangan segan untuk hubungi Layla ya!"
	fmt.Printf("Expected Separate Message: %s\n", expectedSeparate)
	fmt.Printf("Expected Combined Message: %s\n\n", expectedCombined)
}

// printTestResponse prints the test response in a readable format
func printTestResponse(response *services.AIWhatsappResponse) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}
	fmt.Printf("Input Response:\n%s\n\n", string(responseJSON))
}

// simulateOnemessageLogic simulates the PHP onemessage combining logic
// This function demonstrates how the Go implementation should behave
func simulateOnemessageLogic(response *services.AIWhatsappResponse) {
	fmt.Println("🔄 Simulating Onemessage Logic:")

	textParts := []string{}
	isOnemessageActive := false

	for index, part := range response.Response {
		// Validate response part structure
		if part.Type == "" || part.Content == "" {
			fmt.Printf("⚠️  Invalid response part at index %d, skipping\n", index)
			continue
		}

		// Handle text type with "Jenis"="onemessage" combining logic
		if part.Type == "text" && part.Jenis == "onemessage" {
			// Start collecting text parts
			textParts = append(textParts, part.Content)
			isOnemessageActive = true
			fmt.Printf("📝 Collecting onemessage part %d: %s\n", index, part.Content)

			// Check if next part isn't also onemessage, then send combined
			nextIsOnemessage := false
			if index+1 < len(response.Response) {
				nextPart := response.Response[index+1]
				if nextPart.Jenis == "onemessage" {
					nextIsOnemessage = true
				}
			}

			if !nextIsOnemessage {
				// Send combined message
				combinedMessage := fmt.Sprintf("\"%s\"", strings.Join(textParts, "\n"))
				fmt.Printf("📤 Sending BOT_COMBINED: %s\n", combinedMessage)

				// Reset temporary variables
				textParts = []string{}
				isOnemessageActive = false
			}
		} else {
			// If we just finished onemessage sequence, send combined first
			if isOnemessageActive {
				combinedMessage := fmt.Sprintf("\"%s\"", strings.Join(textParts, "\n"))
				fmt.Printf("📤 Sending BOT_COMBINED: %s\n", combinedMessage)

				// Reset variables
				textParts = []string{}
				isOnemessageActive = false
			}

			// Now handle normal text or media
			switch part.Type {
			case "text":
				fmt.Printf("📤 Sending BOT: \"%s\"\n", part.Content)
			case "image":
				fmt.Printf("📤 Sending BOT (image): %s\n", part.Content)
			case "audio":
				fmt.Printf("📤 Sending BOT (audio): %s\n", part.Content)
			case "video":
				fmt.Printf("📤 Sending BOT (video): %s\n", part.Content)
			default:
				fmt.Printf("⚠️  Unknown response type: %s\n", part.Type)
			}
		}
	}

	fmt.Println()
}