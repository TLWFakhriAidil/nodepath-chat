package antipattern

import (
	"hash/fnv"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// GreetingProcessor handles Malaysian greeting variations for anti-spam
type GreetingProcessor struct {
	templates []string
}

// NewGreetingProcessor creates a new greeting processor with Malaysian templates
func NewGreetingProcessor() *GreetingProcessor {
	return &GreetingProcessor{
		templates: []string{
			"{Hi|Hello|Hai} {name}",
			"{name} {hi|hello}",
			"{Salam|Hi|Hello} {name}",
			"{name}, {apa khabar|hi}",
			"{Selamat pagi|Hi} {name}", // Time-aware
		},
	}
}

// GetAntiSpamGreeting generates a unique greeting for each recipient
func (g *GreetingProcessor) GetAntiSpamGreeting(name string, deviceID string, recipientPhone string) string {
	// 1. Device-specific seed (each device has different pattern)
	deviceSeed := hash(deviceID + time.Now().Format("2006-01-02"))
	
	// 2. Recipient-specific variation
	recipientSeed := hash(recipientPhone)
	
	// 3. Time-based variation (changes every 4 hours)
	timeSeed := time.Now().Hour() / 4
	
	// Combine seeds for unique selection
	rand.Seed(int64(deviceSeed + recipientSeed + uint32(timeSeed)))
	
	// Select template based on time if applicable
	template := g.selectTimeAppropriateTemplate()
	
	// Handle name FIRST - if phone number or empty, use "Cik"
	if isPhoneNumber(name) || name == "" {
		name = "Cik"
	}
	
	// Replace {name} placeholder BEFORE processing spintax
	template = strings.ReplaceAll(template, "{name}", name)
	
	// Process spintax AFTER name replacement
	greeting := g.processSpintax(template)
	
	// Apply micro-variations
	return g.applyMicroVariations(greeting)
}

// selectTimeAppropriateTemplate chooses template based on current time
func (g *GreetingProcessor) selectTimeAppropriateTemplate() string {
	hour := time.Now().Hour()
	
	// Morning templates (5am - 12pm)
	if hour >= 5 && hour < 12 {
		// Prefer templates with "Selamat pagi"
		morningTemplates := []string{
			"{Selamat pagi|Pagi} {name}",
			"{Hi|Hello|Hai} {name}",
			"{name}, {selamat pagi|hi}",
		}
		return morningTemplates[rand.Intn(len(morningTemplates))]
	}
	
	// Evening/Night - avoid "pagi"
	if hour >= 19 {
		eveningTemplates := []string{
			"{Hi|Hello|Hai} {name}",
			"{name} {hi|hello}",
			"{Salam|Hi} {name}",
		}
		return eveningTemplates[rand.Intn(len(eveningTemplates))]
	}
	
	// Afternoon - use any template
	return g.templates[rand.Intn(len(g.templates))]
}

// processSpintax processes spintax notation {option1|option2|option3}
func (g *GreetingProcessor) processSpintax(template string) string {
	spintaxRegex := regexp.MustCompile(`\{([^}]+)\}`)
	
	return spintaxRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Remove braces and split by |
		options := strings.Split(strings.Trim(match, "{}"), "|")
		// Randomly select one option
		return options[rand.Intn(len(options))]
	})
}

// applyMicroVariations adds subtle variations to prevent pattern detection
func (g *GreetingProcessor) applyMicroVariations(text string) string {
	// Punctuation variations (50% chance)
	r := rand.Float32()
	if r < 0.3 {
		text = text + "."
	} else if r < 0.5 {
		text = text + ","
	}
	
	// Case variations (20% chance)
	if rand.Float32() < 0.2 && len(text) > 0 {
		// Lowercase first letter
		text = strings.ToLower(text[:1]) + text[1:]
	}
	
	// Spacing variations (invisible but different)
	if rand.Float32() < 0.4 {
		text = text + " " // Extra space at end
	}
	
	return text
}

// isPhoneNumber checks if the name looks like a phone number
func isPhoneNumber(name string) bool {
	// Check if name contains only numbers, +, -, and spaces
	phoneRegex := regexp.MustCompile(`^[0-9\+\-\s]+$`)
	return phoneRegex.MatchString(name)
}

// hash creates a simple hash from string
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// GetMessageDelay calculates delay to prevent burst sending
func GetMessageDelay(deviceID string, messageCount int) time.Duration {
	baseDelay := 5 * time.Second
	
	// Add random jitter (0-10 seconds)
	jitter := time.Duration(rand.Intn(10)) * time.Second
	
	// Increase delay after every 50 messages
	batchDelay := time.Duration(messageCount/50) * 30 * time.Second
	
	return baseDelay + jitter + batchDelay
}

// PrepareMessageWithGreeting adds greeting to original message
func (g *GreetingProcessor) PrepareMessageWithGreeting(originalMessage string, name string, deviceID string, recipientPhone string) string {
	// Get unique greeting
	greeting := g.GetAntiSpamGreeting(name, deviceID, recipientPhone)
	
	// Process original message to handle any spintax it might contain
	processedMessage := g.processSpintax(originalMessage)
	
	// Combine with proper line breaks for WhatsApp
	// Using \n\n for double line break - WhatsApp will handle the encoding
	return greeting + "\n\n" + processedMessage
}
