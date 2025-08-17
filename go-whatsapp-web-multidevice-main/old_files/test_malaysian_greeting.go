package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	// Test the CORRECT Malaysian greeting flow
	fmt.Println("=== MALAYSIAN GREETING + WHATSAPP LINE BREAKS ===")
	
	// Example data
	originalMessage := "Special {offer|deal|promotion} hari ini!\n\nDapatkan diskaun {50%|separuh harga} untuk semua {produk|barangan} kami."
	recipientName := "Ahmad"
	
	fmt.Println("Original message:", originalMessage)
	fmt.Println("Recipient name:", recipientName)
	fmt.Println()
	
	// STEP 1: Malaysian Greeting Based on Time
	hour := time.Now().Hour()
	var greeting string
	
	fmt.Printf("Current hour: %d\n", hour)
	
	if hour >= 5 && hour < 12 {
		greeting = "Selamat pagi " + recipientName + ","
	} else if hour >= 12 && hour < 15 {
		greeting = "Selamat tengahari " + recipientName + ","
	} else if hour >= 15 && hour < 19 {
		greeting = "Selamat petang " + recipientName + ","
	} else {
		greeting = "Selamat malam " + recipientName + ","
	}
	
	fmt.Println("Time-based greeting:", greeting)
	
	// STEP 2: Combine with proper WhatsApp line breaks
	// WhatsApp needs actual newline characters \n, not escaped \\n
	messageWithGreeting := greeting + "\n\n" + originalMessage
	
	fmt.Println("\nMessage with greeting (raw):")
	fmt.Println("---")
	fmt.Println(messageWithGreeting)
	fmt.Println("---")
	
	// STEP 3: Process spintax (in message randomizer)
	processed := strings.ReplaceAll(messageWithGreeting, "{offer|deal|promotion}", "offer")
	processed = strings.ReplaceAll(processed, "{50%|separuh harga}", "50%")
	processed = strings.ReplaceAll(processed, "{produk|barangan}", "produk")
	
	fmt.Println("\nFinal WhatsApp message:")
	fmt.Println("================")
	fmt.Println(processed)
	fmt.Println("================")
	
	// Test with phone number
	fmt.Println("\n\n=== TEST WITH PHONE NUMBER ===")
	phoneAsName := "60123456789"
	fmt.Println("Name field:", phoneAsName)
	fmt.Println("Will use: tuan/puan")
	fmt.Println("\nResult:")
	fmt.Println("Selamat malam tuan/puan,")
	fmt.Println()
	fmt.Println("Special offer hari ini!")
	fmt.Println()
	fmt.Println("Dapatkan diskaun 50% untuk semua produk kami.")
	
	// Show debug view
	fmt.Println("\n=== DEBUG VIEW (to see line breaks) ===")
	debugView := strings.ReplaceAll(processed, "\n", "\\n")
	fmt.Println(debugView)
}
