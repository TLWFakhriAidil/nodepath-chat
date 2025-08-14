package main

import (
	"fmt"
	"strings"
	"time"
)

func isPhoneNumber(name string) bool {
	// Remove spaces and common phone number characters
	cleaned := strings.ReplaceAll(name, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	
	if len(cleaned) == 0 {
		return false
	}
	
	// If the name is very short (less than 3 chars), it's probably not a phone number
	if len(cleaned) < 3 {
		return false
	}
	
	digitCount := 0
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	
	// If more than 70% digits AND at least 5 digits total, it's likely a phone number
	digitPercentage := float64(digitCount) / float64(len(cleaned))
	isPhone := digitPercentage > 0.7 && digitCount >= 5
	
	return isPhone
}

func main() {
	fmt.Println("=== TESTING NAME DETECTION ===")
	
	testNames := []string{
		"Ahmad",
		"Ali",
		"Siti Nurhaliza",
		"Muhammad Ali",
		"60123456789",
		"+60123456789",
		"012-345 6789",
		"Ali123",
		"User1234",
		"12345",
		"Cik",
		"",
	}
	
	for _, name := range testNames {
		isPhone := isPhoneNumber(name)
		result := "Name"
		if isPhone {
			result = "Phone (will use tuan/puan)"
		}
		if name == "" {
			result = "Empty (will use tuan/puan)"
		}
		fmt.Printf("'%s' -> %s\n", name, result)
	}
	
	fmt.Println("\n=== TESTING TIMEZONE ===")
	
	// Test Malaysia timezone
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return
	}
	
	// Current time
	now := time.Now()
	nowMY := now.In(loc)
	
	fmt.Printf("Server time: %s\n", now.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Malaysia time: %s\n", nowMY.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Malaysia hour: %d\n", nowMY.Hour())
	
	hour := nowMY.Hour()
	var greeting string
	if hour >= 5 && hour < 12 {
		greeting = "Selamat pagi"
	} else if hour >= 12 && hour < 15 {
		greeting = "Selamat tengahari"
	} else if hour >= 15 && hour < 19 {
		greeting = "Selamat petang"
	} else {
		greeting = "Selamat malam"
	}
	
	fmt.Printf("Greeting: %s\n", greeting)
}
