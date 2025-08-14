package antipattern


import (
	"math/rand"
	"strings"
	"time"
	"unicode"
	"regexp"
)

// MessageRandomizer handles message anti-pattern techniques
type MessageRandomizer struct {
	homoglyphs map[rune][]rune
	zeroWidthChars []string
}

// NewMessageRandomizer creates a new message randomizer
func NewMessageRandomizer() *MessageRandomizer {
	return &MessageRandomizer{
		homoglyphs: initHomoglyphs(),
		zeroWidthChars: []string{
			"\u200B", // Zero-width space
			"\u200C", // Zero-width non-joiner
			"\u200D", // Zero-width joiner
			"\uFEFF", // Zero-width no-break space
		},
	}
}

// RandomizeMessage applies all anti-pattern techniques
func (mr *MessageRandomizer) RandomizeMessage(message string) string {
	if message == "" {
		return message
	}
	
	// Apply techniques in order
	// 1. First process any spintax in the message content
	message = mr.processSpintax(message)
	
	// 2. Then apply homoglyphs (5% of characters)
	message = mr.applyHomoglyphs(message, 0.05)
	
	// 3. Insert zero-width spaces
	message = mr.insertZeroWidthSpaces(message, 2)
	
	// 4. Randomize punctuation
	message = mr.randomizePunctuation(message)
	
	return message
}

// processSpintax processes spintax patterns {option1|option2|option3} in the message
func (mr *MessageRandomizer) processSpintax(text string) string {
	spintaxRegex := regexp.MustCompile(`\{([^}]+)\}`)
	
	result := spintaxRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Remove braces
		content := match[1 : len(match)-1]
		
		// Split options
		options := strings.Split(content, "|")
		
		// Randomly select one
		return options[rand.Intn(len(options))]
	})
	
	return result
}

// applyHomoglyphs replaces a percentage of characters with look-alikes
func (mr *MessageRandomizer) applyHomoglyphs(text string, percentage float32) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	
	// Calculate how many characters to replace
	replaceCount := int(float32(len(runes)) * percentage)
	if replaceCount == 0 {
		replaceCount = 1 // At least one
	}
	
	// Get random positions to replace (only letters)
	positions := mr.getLetterPositions(runes)
	if len(positions) == 0 {
		return text
	}
	
	// Shuffle and take only needed positions
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	
	if len(positions) > replaceCount {
		positions = positions[:replaceCount]
	}
	
	// Replace characters at selected positions
	for _, pos := range positions {
		char := unicode.ToLower(runes[pos])
		if replacements, exists := mr.homoglyphs[char]; exists && len(replacements) > 0 {
			// Pick random homoglyph
			replacement := replacements[rand.Intn(len(replacements))]
			
			// Preserve original case
			if unicode.IsUpper(runes[pos]) {
				replacement = unicode.ToUpper(replacement)
			}
			
			runes[pos] = replacement
		}
	}
	
	return string(runes)
}

// insertZeroWidthSpaces adds invisible characters between words
func (mr *MessageRandomizer) insertZeroWidthSpaces(text string, count int) string {
	if count <= 0 || len(text) == 0 {
		return text
	}
	
	words := strings.Fields(text)
	if len(words) <= 1 {
		return text
	}
	
	// Find random positions between words
	possiblePositions := len(words) - 1
	if count > possiblePositions {
		count = possiblePositions
	}
	
	// Insert zero-width characters
	for i := 0; i < count; i++ {
		pos := rand.Intn(len(words)-1) + 1
		zeroWidth := mr.zeroWidthChars[rand.Intn(len(mr.zeroWidthChars))]
		words[pos] = zeroWidth + words[pos]
	}
	
	return strings.Join(words, " ")
}

// randomizePunctuation adds subtle variations to punctuation
func (mr *MessageRandomizer) randomizePunctuation(text string) string {
	// Random variations
	variations := []struct{
		probability float32
		transform   func(string) string
	}{
		{0.1, func(s string) string { return strings.ReplaceAll(s, "!", "! ") }},
		{0.1, func(s string) string { return strings.ReplaceAll(s, "?", "? ") }},
		{0.1, func(s string) string { return strings.ReplaceAll(s, ".", ". ") }},
		{0.05, func(s string) string { return strings.ReplaceAll(s, ",", ", ") }},
	}
	
	for _, v := range variations {
		if rand.Float32() < v.probability {
			text = v.transform(text)
		}
	}
	
	// Clean up multiple spaces
	text = strings.ReplaceAll(text, "  ", " ")
	
	return strings.TrimSpace(text)
}

// getLetterPositions returns positions of letters in the text
func (mr *MessageRandomizer) getLetterPositions(runes []rune) []int {
	positions := []int{}
	for i, r := range runes {
		if unicode.IsLetter(r) {
			positions = append(positions, i)
		}
	}
	return positions
}

// initHomoglyphs initializes the homoglyph map
func initHomoglyphs() map[rune][]rune {
	return map[rune][]rune{
		'a': {'а', 'ɑ', 'α'}, // Cyrillic а, Latin alpha, Greek alpha
		'b': {'Ь', 'ƅ', 'ḃ'}, // Cyrillic soft sign, Latin b with stroke
		'c': {'с', 'ϲ', 'ć'}, // Cyrillic с, Greek lunate sigma
		'd': {'ԁ', 'ɗ', 'ḍ'}, // Cyrillic d, Latin d with hook
		'e': {'е', 'ė', 'ẹ'}, // Cyrillic е, Latin e with dot
		'f': {'ƒ', 'ḟ'},      // Latin f with hook, f with dot
		'g': {'ɡ', 'ġ', 'ǵ'}, // Latin script g, g with dot
		'h': {'һ', 'ḣ', 'ḥ'}, // Cyrillic һ, Latin h with dot
		'i': {'і', 'ı', 'ḭ'}, // Cyrillic і, Latin dotless i
		'j': {'ј', 'ĵ', 'ǰ'}, // Cyrillic ј, Latin j with circumflex
		'k': {'κ', 'ḳ', 'ķ'}, // Greek kappa, Latin k with dot
		'l': {'ⅼ', 'ḷ', 'ļ'}, // Small Roman numeral, Latin l with dot
		'm': {'м', 'ṁ', 'ḿ'}, // Cyrillic м, Latin m with dot
		'n': {'ո', 'ṅ', 'ń'}, // Armenian vo, Latin n with dot
		'o': {'о', 'ο', 'ȯ'}, // Cyrillic о, Greek omicron
		'p': {'р', 'ρ', 'ṗ'}, // Cyrillic р, Greek rho
		'q': {'ԛ', 'ɋ'},      // Cyrillic qa
		'r': {'г', 'ṙ', 'ŕ'}, // Cyrillic г, Latin r with dot
		's': {'ѕ', 'ṡ', 'ś'}, // Cyrillic ѕ, Latin s with dot
		't': {'τ', 'ṫ', 'ť'}, // Greek tau, Latin t with dot
		'u': {'υ', 'ս', 'ů'}, // Greek upsilon, Armenian
		'v': {'ν', 'ѵ', 'ṿ'}, // Greek nu, Cyrillic v
		'w': {'ԝ', 'ẇ', 'ẃ'}, // Cyrillic w, Latin w with dot
		'x': {'х', 'ẋ', 'ẍ'}, // Cyrillic х, Latin x with dot
		'y': {'у', 'ү', 'ẏ'}, // Cyrillic у, Latin y with dot
		'z': {'ᴢ', 'ż', 'ź'}, // Small capital z, Latin z with dot
	}
}

// GetRandomDelay returns a random delay between min and max seconds
func GetRandomDelay(minSeconds, maxSeconds int) time.Duration {
	if minSeconds >= maxSeconds {
		return time.Duration(minSeconds) * time.Second
	}
	
	delay := rand.Intn(maxSeconds-minSeconds) + minSeconds
	// Add random milliseconds for more variation
	milliseconds := rand.Intn(1000)
	
	return time.Duration(delay)*time.Second + time.Duration(milliseconds)*time.Millisecond
}

// AddTypingDelay returns a random typing duration based on message length
func AddTypingDelay(messageLength int) time.Duration {
	// Base: 50ms per character + random variation
	baseTime := messageLength * 50
	variation := rand.Intn(baseTime / 2) // 0-50% variation
	
	totalMs := baseTime + variation
	
	// Cap between 2-8 seconds
	if totalMs < 2000 {
		totalMs = 2000
	} else if totalMs > 8000 {
		totalMs = 8000
	}
	
	return time.Duration(totalMs) * time.Millisecond
}

// init initializes the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}
