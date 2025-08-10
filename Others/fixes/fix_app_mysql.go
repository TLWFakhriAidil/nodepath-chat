package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Fix the UUID cast issue in app.go
	filename := "src/ui/rest/app.go"
	
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filename, err)
		return
	}
	
	originalContent := string(content)
	modifiedContent := originalContent
	
	// Fix 1: Replace the problematic query with MySQL-compatible version
	oldQuery1 := `err = db.QueryRow(` + "`" + `
				SELECT COUNT(*) 
				FROM sequence_steps ss
				WHERE EXISTS (
					SELECT 1 FROM sequences s 
					WHERE s.id = ss.sequence_id 
					AND s.user_id = ?
				)
			` + "`" + `, session.UserID).Scan(&flowCount)`
			
	newQuery1 := `err = db.QueryRow(` + "`" + `
				SELECT COUNT(*) 
				FROM sequence_steps ss
				INNER JOIN sequences s ON s.id = ss.sequence_id
				WHERE s.user_id = ?
			` + "`" + `, session.UserID).Scan(&flowCount)`
	
	modifiedContent = strings.Replace(modifiedContent, oldQuery1, newQuery1, 1)
	
	// Fix 2: Remove the UUID cast fallback query
	oldFallback := `// Try with UUID casting
				err = db.QueryRow(` + "`" + `
					SELECT COUNT(*) 
					FROM sequence_steps ss
					WHERE sequence_id::text IN (
						SELECT id::text FROM sequences WHERE user_id = ?
					)
				` + "`" + `, session.UserID).Scan(&flowCount)
				
				if err != nil {
					fmt.Printf("Error with UUID cast query: %v\n", err)
				} else {
					totalFlows = flowCount
				}`
				
	newFallback := `// UUID casting not needed for MySQL
				totalFlows = 0 // Default to 0 if query fails`
	
	modifiedContent = strings.Replace(modifiedContent, oldFallback, newFallback, 1)
	
	// Fix 3: Fix campaign summary query if present
	// Look for campaign summary issues
	oldCampaignQuery := `COUNT(DISTINCT CASE WHEN b.status = 'sent' THEN b.recipient_phone END)`
	newCampaignQuery := `COUNT(DISTINCT CASE WHEN b.status = 'success' THEN b.recipient_phone END)`
	
	modifiedContent = strings.Replace(modifiedContent, oldCampaignQuery, newCampaignQuery, -1)
	
	// Write the fixed content
	if modifiedContent != originalContent {
		// Backup original
		backupFile := filename + ".mysql_fix.bak"
		ioutil.WriteFile(backupFile, content, 0644)
		
		// Write fixed content
		err = ioutil.WriteFile(filename, []byte(modifiedContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			return
		}
		fmt.Printf("Fixed: %s (backup: %s)\n", filename, backupFile)
	} else {
		fmt.Printf("No changes needed: %s\n", filename)
	}
}
