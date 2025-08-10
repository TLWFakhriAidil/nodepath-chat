package antipattern

import (
	"strings"
	"time"
	"math/rand"
	"unicode"
)

// GreetingProcessor handles greeting for messages
type GreetingProcessor struct {
	// Malaysian greetings - NO SPINTAX, will randomly pick one
	greetings []string
}

// NewGreetingProcessor creates a new greeting processor
func NewGreetingProcessor() *GreetingProcessor {
	return &GreetingProcessor{
		greetings: []string{
			"Salam {name}",
			"Hi {name}",
			"Apa khabar {name}",
			"Maaf ganggu {name}",
			"Pinjam masa {name}",
		},
	}
}

// cleanName removes all numbers from the name and returns cleaned name
func cleanName(name string) string {
	var result strings.Builder
	
	for _, r := range name {
		// Keep only letters and spaces
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	
	// Trim any extra spaces
	cleaned := strings.TrimSpace(result.String())
	
	return cleaned
}

// getSimpleGreeting returns a Malaysian greeting based on time
func (g *GreetingProcessor) getSimpleGreeting(name string) string {
	// Clean the name by removing all numbers
	cleanedName := cleanName(name)
	
	// If cleaned name is empty, use "Cik"
	if cleanedName == "" {
		name = "Cik"
	} else {
		name = cleanedName
	}
	
	// Get current hour and add 5 hours for Malaysia time adjustment
	hour := time.Now().Hour()
	hour = (hour + 5) % 24
	
	var greeting string
	
	// Time-based Malaysian greetings
	if hour >= 5 && hour < 12 {
		// Morning greetings
		morningGreetings := []string{
			"Selamat pagi {name}",
			"Pagi {name}",
			"Assalamualaikum {name}",
		}
		greeting = morningGreetings[rand.Intn(len(morningGreetings))]
	} else if hour >= 12 && hour < 15 {
		// Afternoon
		afternoonGreetings := []string{
			"Selamat tengahari {name}",
			"Salam {name}",
			"Hi {name}",
		}
		greeting = afternoonGreetings[rand.Intn(len(afternoonGreetings))]
	} else if hour >= 15 && hour < 19 {
		// Late afternoon/evening
		eveningGreetings := []string{
			"Selamat petang {name}",
			"Petang {name}",
			"Salam {name}",
		}
		greeting = eveningGreetings[rand.Intn(len(eveningGreetings))]
	} else {
		// Night (7pm - 5am)
		nightGreetings := []string{
			"Selamat malam {name}",
			"Malam {name}",
			"Maaf ganggu {name}",
			"Pinjam masa {name}",
		}
		greeting = nightGreetings[rand.Intn(len(nightGreetings))]
	}
	
	// Replace {name} with actual name
	greeting = strings.ReplaceAll(greeting, "{name}", name)
	
	// Add comma at the end
	return greeting + ","
}

// GetAntiSpamGreeting - keeping for compatibility
func (g *GreetingProcessor) GetAntiSpamGreeting(name string, deviceID string, recipientPhone string) string {
	return g.getSimpleGreeting(name)
}

// PrepareMessageWithGreeting adds greeting to message
func (g *GreetingProcessor) PrepareMessageWithGreeting(originalMessage string, name string, deviceID string, recipientPhone string) string {
	// STEP 1: Get the simple Malaysian greeting (NO SPINTAX!)
	greeting := g.getSimpleGreeting(name)
	
	// STEP 2: Process the message content
	// First replace {name} in the message if exists
	processedMessage := originalMessage
	if strings.Contains(processedMessage, "{name}") {
		// Use cleaned name for message content too
		properName := cleanName(name)
		if properName == "" {
			properName = "Cik"
		}
		processedMessage = strings.ReplaceAll(processedMessage, "{name}", properName)
	}
	
	// STEP 3: Fix line breaks for WhatsApp
	// WhatsApp uses actual newlines, not escaped ones
	// Convert all possible line break formats to proper newlines
	processedMessage = strings.ReplaceAll(processedMessage, "\\n", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "\\r\\n", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "%0A", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "%0a", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "<br>", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "<br/>", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "<br />", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "[br]", "\n")
	processedMessage = strings.ReplaceAll(processedMessage, "{br}", "\n")
	
	// STEP 4: Combine greeting + proper WhatsApp line breaks + message
	// For WhatsApp, we need actual newline characters, not escaped ones
	// Using real newlines that WhatsApp will recognize
	finalMessage := greeting + "\n\n" + processedMessage
	
	// Return the message with real newlines for WhatsApp
	return finalMessage
}

// isPhoneNumber - NOT USED ANYMORE but kept for compatibility
func isPhoneNumber(name string) bool {
	// This function is no longer used
	// We now use cleanName() instead
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// init initializes the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}