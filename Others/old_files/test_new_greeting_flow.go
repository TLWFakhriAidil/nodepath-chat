package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	// Example content with spintax
	content := "Special {offer|deal|promotion} for you!\n\nGet {50%|half price|huge discount} on all {products|items|goods}.\n\nValid {today|this week|now} only!"
	name := "Ahmad"
	
	fmt.Println("=== NEW GREETING PROCESSOR FLOW ===")
	fmt.Println("Original content:", content)
	fmt.Println("Name:", name)
	fmt.Println()
	
	// STEP 1: Process spintax on CONTENT ONLY
	processedContent := strings.ReplaceAll(content, "{offer|deal|promotion}", "offer")
	processedContent = strings.ReplaceAll(processedContent, "{50%|half price|huge discount}", "50%")
	processedContent = strings.ReplaceAll(processedContent, "{products|items|goods}", "products")
	processedContent = strings.ReplaceAll(processedContent, "{today|this week|now}", "today")
	
	fmt.Println("Step 1 - Content after spintax processing:")
	fmt.Println(processedContent)
	fmt.Println()
	
	// STEP 2: Replace {name} if exists in content
	if strings.Contains(processedContent, "{name}") {
		processedContent = strings.ReplaceAll(processedContent, "{name}", name)
	}
	
	// STEP 3: Fix line breaks (already has \n in this example)
	
	// STEP 4: Get simple greeting based on time (NO SPINTAX)
	hour := time.Now().Hour()
	var greeting string
	
	if hour >= 5 && hour < 12 {
		greeting = "Selamat pagi " + name + ","
	} else if hour >= 12 && hour < 18 {
		greeting = "Selamat petang " + name + ","
	} else {
		greeting = "Selamat malam " + name + ","
	}
	
	fmt.Println("Step 4 - Time-based greeting (no spintax):")
	fmt.Println(greeting)
	fmt.Println()
	
	// STEP 5: Combine with proper line breaks
	finalMessage := greeting + "\n\n" + processedContent
	
	fmt.Println("Final message:")
	fmt.Println("================")
	fmt.Println(finalMessage)
	fmt.Println("================")
	
	// Show another example with {name} in content
	fmt.Println("\n\n=== EXAMPLE WITH {name} IN CONTENT ===")
	content2 := "Hello {name}, check out our {new|latest} products!"
	
	// Process spintax first
	processed2 := strings.ReplaceAll(content2, "{new|latest}", "new")
	// Then replace name
	processed2 = strings.ReplaceAll(processed2, "{name}", name)
	// Add greeting
	final2 := greeting + "\n\n" + processed2
	
	fmt.Println("Final:")
	fmt.Println("================")
	fmt.Println(final2)
	fmt.Println("================")
}
