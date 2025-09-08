package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"nodepath-chat/internal/services"
)

func main() {
	fmt.Println("=== Testing AI Response Processing with Onemessage Combining ===\n")

	// Initialize the AI response processor
	processor := services.NewAIResponseProcessor(5 * time.Second)

	// Test Case 1: Response with onemessage combining
	testCase1 := `{
		"Stage": "Problem Identification",
		"Response": [
			{"type": "text", "Jenis": "onemessage", "content": "Alhamdulillah, ramai ibu yang dah berjaya bantu anak mereka dengan Vitac! 😊"},
			{"type": "image", "content": "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665515"},
			{"type": "image", "content": "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665523"},
			{"type": "image", "content": "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665533"},
			{"type": "text", "Jenis": "onemessage", "content": "Akak pun nak anak lebih sihat dan aktif kan?"}
		]
	}`

	fmt.Println("Test Case 1: Mixed text with onemessage and images")
	fmt.Println("Input JSON:")
	fmt.Println(testCase1)
	fmt.Println()

	// Process the response
	messages1, err := processor.ProcessAIResponse(testCase1, nil)
	if err != nil {
		log.Fatalf("Error processing test case 1: %v", err)
	}

	fmt.Println("Processed Messages:")
	for i, msg := range messages1 {
		fmt.Printf("%d. Type: %s\n", i+1, msg.Type)
		if msg.Type == "text" {
			fmt.Printf("   Content: %s\n", msg.Content)
		} else {
			fmt.Printf("   URL: %s\n", msg.Content)
		}
	}
	fmt.Println()

	// Test Case 2: Multiple onemessage parts that should be combined
	testCase2 := `{
		"Stage": "Solution Presentation",
		"Response": [
			{"type": "text", "Jenis": "onemessage", "content": "Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni."},
			{"type": "text", "Jenis": "onemessage", "content": "Kurang selera makan, sembelit, atau kerap demam?"},
			{"type": "text", "Jenis": "onemessage", "content": "Atau ada masalah lain yang akak nak kongsikan?"}
		]
	}`

	fmt.Println("Test Case 2: Multiple onemessage texts that should be combined")
	fmt.Println("Input JSON:")
	fmt.Println(testCase2)
	fmt.Println()

	messages2, err := processor.ProcessAIResponse(testCase2, nil)
	if err != nil {
		log.Fatalf("Error processing test case 2: %v", err)
	}

	fmt.Println("Processed Messages:")
	for i, msg := range messages2 {
		fmt.Printf("%d. Type: %s\n", i+1, msg.Type)
		fmt.Printf("   Content: %s\n", msg.Content)
		if strings.Contains(msg.Content, "\n") {
			fmt.Println("   ✅ Combined message detected (contains newlines)")
		}
	}
	fmt.Println()

	// Test Case 3: Regular messages without onemessage
	testCase3 := `{
		"Stage": "Closing",
		"Response": [
			{"type": "text", "content": "Terima kasih kerana sudi berbual dengan saya."},
			{"type": "image", "content": "https://example.com/thank-you.jpg"},
			{"type": "text", "content": "Jumpa lagi!"}
		]
	}`

	fmt.Println("Test Case 3: Regular messages without onemessage (should NOT be combined)")
	fmt.Println("Input JSON:")
	fmt.Println(testCase3)
	fmt.Println()

	messages3, err := processor.ProcessAIResponse(testCase3, nil)
	if err != nil {
		log.Fatalf("Error processing test case 3: %v", err)
	}

	fmt.Println("Processed Messages:")
	for i, msg := range messages3 {
		fmt.Printf("%d. Type: %s\n", i+1, msg.Type)
		if msg.Type == "text" {
			fmt.Printf("   Content: %s\n", msg.Content)
		} else {
			fmt.Printf("   URL: %s\n", msg.Content)
		}
	}
	fmt.Println()

	// Test Case 4: Mixed onemessage groups
	testCase4 := `{
		"Stage": "Information Gathering",
		"Response": [
			{"type": "text", "Jenis": "onemessage", "content": "Part 1 of first group"},
			{"type": "text", "Jenis": "onemessage", "content": "Part 2 of first group"},
			{"type": "text", "content": "Regular message in between"},
			{"type": "text", "Jenis": "onemessage", "content": "Part 1 of second group"},
			{"type": "text", "Jenis": "onemessage", "content": "Part 2 of second group"}
		]
	}`

	fmt.Println("Test Case 4: Multiple onemessage groups with regular message in between")
	fmt.Println("Input JSON:")
	fmt.Println(testCase4)
	fmt.Println()

	messages4, err := processor.ProcessAIResponse(testCase4, nil)
	if err != nil {
		log.Fatalf("Error processing test case 4: %v", err)
	}

	fmt.Println("Processed Messages:")
	for i, msg := range messages4 {
		fmt.Printf("%d. Type: %s\n", i+1, msg.Type)
		fmt.Printf("   Content: %s\n", msg.Content)
		if strings.Contains(msg.Content, "\n") {
			fmt.Println("   ✅ Combined message detected")
		}
	}
	fmt.Println()

	// Test logging format
	fmt.Println("=== Test Logging Format ===")
	logEntries := processor.FormatResponseForLogging(messages1, "BOT")
	fmt.Println("Log entries for Test Case 1:")
	for _, entry := range logEntries {
		fmt.Println(entry)
	}

	fmt.Println("\n✅ All tests completed successfully!")
}
