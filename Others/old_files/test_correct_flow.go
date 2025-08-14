package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the CORRECT flow
	fmt.Println("=== CORRECT GREETING + SPINTAX FLOW ===")
	
	// Example data
	originalMessage := "Special {offer|deal|promotion} for you! Get {50%|half price} discount on {our|the} products."
	recipientName := "Ahmad"  // Real name from lead
	
	fmt.Println("Original message:", originalMessage)
	fmt.Println("Recipient name:", recipientName)
	fmt.Println()
	
	// STEP 1: Greeting Processor - Just add simple greeting
	fmt.Println("STEP 1: Greeting Processor")
	greeting := "Hi " + recipientName + ","
	messageWithGreeting := greeting + "\n\n" + originalMessage
	fmt.Println("After greeting added:")
	fmt.Println("---")
	fmt.Println(messageWithGreeting)
	fmt.Println("---")
	fmt.Println()
	
	// STEP 2: Message Randomizer - Process spintax + homoglyphs
	fmt.Println("STEP 2: Message Randomizer")
	// 2a. Process spintax
	processed := strings.ReplaceAll(messageWithGreeting, "{offer|deal|promotion}", "offer")
	processed = strings.ReplaceAll(processed, "{50%|half price}", "50%")
	processed = strings.ReplaceAll(processed, "{our|the}", "our")
	fmt.Println("After spintax:")
	fmt.Println("---")
	fmt.Println(processed)
	fmt.Println("---")
	
	// 2b. Apply homoglyphs (10% - just example)
	// Would replace some characters like 'a' -> 'а' (Cyrillic)
	fmt.Println("\nAfter homoglyphs (10% chars changed):")
	fmt.Println("Example: 'Special' might become 'Speciаl' (with Cyrillic 'а')")
	
	fmt.Println("\n=== FINAL MESSAGE SENT ===")
	fmt.Println("Hi Ahmad,")
	fmt.Println()
	fmt.Println("Special offer for you! Get 50% discount on our products.")
	fmt.Println("(with some letters replaced by homoglyphs)")
	
	// Test with phone number as name
	fmt.Println("\n\n=== TEST WITH PHONE NUMBER AS NAME ===")
	phoneAsName := "60123456789"
	fmt.Println("Name field:", phoneAsName)
	fmt.Println("Detected as phone number: YES")
	fmt.Println("Will use greeting: Hi there,")
	fmt.Println("\nFinal:")
	fmt.Println("Hi there,")
	fmt.Println()
	fmt.Println("Special offer for you! Get 50% discount on our products.")
}
