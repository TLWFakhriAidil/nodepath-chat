package main

import (
	"fmt"
	"strings"
	"unicode"
)

func cleanName(name string) string {
	var result strings.Builder
	
	for _, r := range name {
		// Keep all non-digit characters
		if !unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	
	// Trim any extra spaces
	cleaned := strings.TrimSpace(result.String())
	
	return cleaned
}

func main() {
	fmt.Println("=== TESTING SIMPLE NAME CLEANING ===")
	
	testNames := []string{
		"Ahmad",
		"Ahmad123",
		"Ali 2",
		"Siti Nurhaliza",
		"Muhammad Ali",
		"60123456789",
		"+60123456789",
		"012-345 6789",
		"User1234",
		"12345",
		"Cik",
		"",
		"123Ahmad456",
		"A1h2m3a4d",
	}
	
	fmt.Println("Original Name -> Cleaned Name -> Final Greeting Name")
	fmt.Println("----------------------------------------------------")
	
	for _, name := range testNames {
		cleaned := cleanName(name)
		finalName := cleaned
		if cleaned == "" {
			finalName = "Cik"
		}
		
		fmt.Printf("'%s' -> '%s' -> '%s'\n", name, cleaned, finalName)
	}
	
	fmt.Println("\n=== EXAMPLE GREETINGS ===")
	
	// Example greetings
	examples := []string{"Ahmad123", "60123456789", "Ali 2", ""}
	
	for _, name := range examples {
		cleaned := cleanName(name)
		if cleaned == "" {
			cleaned = "Cik"
		}
		fmt.Printf("\nOriginal: '%s'\n", name)
		fmt.Printf("Greeting: 'Selamat malam %s,'\n", cleaned)
	}
}
