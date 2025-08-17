package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the current flow
	originalMessage := "Special offer {today|now}! Get {50%|half price} discount on {our|the} products."
	name := "John"
	
	fmt.Println("=== CURRENT FLOW TEST ===")
	fmt.Println("Original message:", originalMessage)
	fmt.Println("Name:", name)
	
	// Step 1: Replace {name} in message
	processedMessage := strings.ReplaceAll(originalMessage, "{name}", name)
	fmt.Println("\nAfter name replacement:", processedMessage)
	
	// Step 2: Process spintax (simulate)
	// In real code this would randomly select options
	processedMessage = strings.ReplaceAll(processedMessage, "{today|now}", "today")
	processedMessage = strings.ReplaceAll(processedMessage, "{50%|half price}", "50%")
	processedMessage = strings.ReplaceAll(processedMessage, "{our|the}", "our")
	fmt.Println("\nAfter spintax processing:", processedMessage)
	
	// Step 3: Generate greeting (already has name)
	greeting := "Hi John,"
	fmt.Println("\nGreeting:", greeting)
	
	// Step 4: Combine
	finalMessage := greeting + "\n\n" + processedMessage
	fmt.Println("\nFinal message:")
	fmt.Println(finalMessage)
	
	fmt.Println("\n=== WHAT YOU WANT ===")
	// What you want: Process spintax FIRST on content, THEN add greeting
	originalMessage2 := "Special offer {today|now}! Get {50%|half price} discount on {our|the} products."
	
	// Step 1: Process spintax on content FIRST
	processedContent := strings.ReplaceAll(originalMessage2, "{today|now}", "today")
	processedContent = strings.ReplaceAll(processedContent, "{50%|half price}", "50%") 
	processedContent = strings.ReplaceAll(processedContent, "{our|the}", "our")
	fmt.Println("Content after spintax:", processedContent)
	
	// Step 2: Add greeting to the TOP
	greeting2 := "Hi John,"
	finalMessage2 := greeting2 + "\n\n" + processedContent
	fmt.Println("\nFinal message:")
	fmt.Println(finalMessage2)
}
