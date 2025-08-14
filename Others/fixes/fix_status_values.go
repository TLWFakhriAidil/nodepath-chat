package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Fix status values across multiple files
	filesToFix := map[string][]string{
		"src/repository/campaign_repository.go": {
			// Fix the status check in GetCampaignBroadcastStats
			`COUNT(CASE WHEN status = 'sent' THEN 1 END)`,
			`COUNT(CASE WHEN status = 'success' THEN 1 END)`,
		},
		"src/repository/sequence_repository.go": {
			// Fix any sequence status checks
			`status = 'sent'`,
			`status = 'success'`,
		},
		"src/ui/rest/app.go": {
			// Fix dashboard status checks
			`b.status = 'sent'`,
			`b.status = 'success'`,
		},
	}
	
	for filename, replacements := range filesToFix {
		fixFile(filename, replacements)
	}
	
	// Also fix the specific campaign repository issue
	fixCampaignRepo()
}

func fixFile(filename string, replacements []string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filename, err)
		return
	}
	
	originalContent := string(content)
	modifiedContent := originalContent
	
	// Apply replacements in pairs (old, new)
	for i := 0; i < len(replacements)-1; i += 2 {
		old := replacements[i]
		new := replacements[i+1]
		modifiedContent = strings.ReplaceAll(modifiedContent, old, new)
	}
	
	if modifiedContent != originalContent {
		// Write fixed content
		err = ioutil.WriteFile(filename, []byte(modifiedContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			return
		}
		fmt.Printf("Fixed status values in: %s\n", filename)
	}
}

func fixCampaignRepo() {
	filename := "src/repository/campaign_repository.go"
	
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filename, err)
		return
	}
	
	originalContent := string(content)
	modifiedContent := originalContent
	
	// Fix the GetCampaignBroadcastStats query
	oldQuery := `SELECT COUNT(CASE WHEN status = 'sent' THEN 1 END) AS done_send,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed_send
		FROM broadcast_messages
		WHERE campaign_id = ?`
		
	newQuery := `SELECT COUNT(CASE WHEN status = 'success' THEN 1 END) AS done_send,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed_send
		FROM broadcast_messages
		WHERE campaign_id = ?`
	
	modifiedContent = strings.Replace(modifiedContent, oldQuery, newQuery, 1)
	
	// Also check if we need to fix the niche matching
	// Change exact match to LIKE for flexibility
	oldNicheMatch := `AND l.niche = ?`
	newNicheMatch := `AND l.niche LIKE CONCAT('%', ?, '%')`
	
	// This replacement is already in place, but let's ensure it
	if !strings.Contains(modifiedContent, newNicheMatch) {
		modifiedContent = strings.Replace(modifiedContent, oldNicheMatch, newNicheMatch, -1)
	}
	
	// Write back if changed
	if modifiedContent != originalContent {
		err = ioutil.WriteFile(filename, []byte(modifiedContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			return
		}
		fmt.Printf("Fixed campaign repository status checks\n")
	} else {
		fmt.Printf("No additional fixes needed for campaign repository\n")
	}
}
