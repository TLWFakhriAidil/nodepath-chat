package main

import (
	"fmt"
	"regexp"
	"strings"
	"math/rand"
)

// Simulate the actual processSpintax function
func processSpintax(text string) string {
	spintaxRegex := regexp.MustCompile(`\{([^}]+)\}`)
	
	result := spintaxRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Remove braces
		content := match[1 : len(match)-1]
		
		// Split options
		options := strings.Split(content, "|")
		
		// For testing, just take first option
		return options[0]
	})
	
	return result
}

func main() {
	// Test actual flow from greeting_processor.go
	originalMessage := "Special {offer|deal} for {name}! Get {50%|half price} off."
	name := "Ahmad"
	
	fmt.Println("=== ACTUAL FLOW IN greeting_processor.go ===")
	fmt.Println("Original message:", originalMessage)
	fmt.Println("Name:", name)
	
	// Step 1: Get greeting (with its own spintax)
	greetingTemplate := "{Hi|Hello|Hai} {name},"
	greeting := strings.ReplaceAll(greetingTemplate, "{name}", name)
	greeting = processSpintax(greeting) // This becomes "Hi Ahmad,"
	fmt.Println("\nGreeting after processing:", greeting)
	
	// Step 2: Process message
	// First replace {name}
	processedMessage := strings.ReplaceAll(originalMessage, "{name}", name)
	fmt.Println("\nMessage after name replacement:", processedMessage)
	
	// Then process spintax on message
	processedMessage = processSpintax(processedMessage)
	fmt.Println("Message after spintax:", processedMessage)
	
	// Step 3: Fix line breaks
	processedMessage = strings.ReplaceAll(processedMessage, "\\n", "\n")
	
	// Step 4: Combine
	finalMessage := greeting + "\n\n" + processedMessage
	fmt.Println("\nFinal combined message:")
	fmt.Println("---")
	fmt.Println(finalMessage)
	fmt.Println("---")
	
	// Show what happens with line breaks in content
	fmt.Println("\n=== WITH LINE BREAKS IN CONTENT ===")
	messageWithBreaks := "Special {offer|deal}!\\n\\nGet {50%|half price} off {today|now}."
	
	// Process it
	processedWithBreaks := processSpintax(messageWithBreaks)
	processedWithBreaks = strings.ReplaceAll(processedWithBreaks, "\\n", "\n")
	
	finalWithBreaks := greeting + "\n\n" + processedWithBreaks
	fmt.Println("Final message with breaks:")
	fmt.Println("---")
	fmt.Println(finalWithBreaks)
	fmt.Println("---")
}
