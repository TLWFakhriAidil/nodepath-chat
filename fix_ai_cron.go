package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Read the file
	content, err := ioutil.ReadFile("internal/services/ai_cron_service.go")
	if err != nil {
		panic(err)
	}
	
	fileStr := string(content)
	
	// Find where the orphaned code starts (after line 1147)
	// We need to keep everything up to and including the first closing brace of sendWahaMultimediaMessage
	// Then remove everything after that until we find the next valid function
	
	// Split by lines
	lines := strings.Split(fileStr, "\n")
	
	// Find line 1147 and remove orphaned code
	newLines := []string{}
	inOrphanedSection := false
	braceCount := 0
	foundFirstEnd := false
	
	for i, line := range lines {
		// Keep lines up to 1147
		if i < 1147 {
			newLines = append(newLines, line)
			continue
		}
		
		// After line 1147, we're looking for the end of the function
		if !foundFirstEnd {
			newLines = append(newLines, line)
			if strings.Contains(line, "{") {
				braceCount++
			}
			if strings.Contains(line, "}") {
				braceCount--
				if braceCount <= 0 && strings.TrimSpace(line) == "}" {
					foundFirstEnd = true
					inOrphanedSection = true
				}
			}
		} else if inOrphanedSection {
			// Skip orphaned code until we find a proper function declaration
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "// ") && strings.Contains(line, "func") {
				inOrphanedSection = false
				newLines = append(newLines, line)
			}
			// Skip this line if we're still in orphaned section
		} else {
			// Keep everything else
			newLines = append(newLines, line)
		}
	}
	
	// Write back
	newContent := strings.Join(newLines, "\n")
	err = ioutil.WriteFile("internal/services/ai_cron_service.go", []byte(newContent), 0644)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("File fixed!")
}
