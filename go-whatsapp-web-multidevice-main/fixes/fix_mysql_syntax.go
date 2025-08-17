package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Fix MySQL syntax issues in Go files
func main() {
	fixes := map[string]string{
		// Fix ON CONFLICT syntax to MySQL's ON DUPLICATE KEY UPDATE
		`INSERT INTO whatsapp_chats (device_id, chat_jid, chat_name, last_message_time)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (device_id, chat_jid) 
		DO UPDATE SET 
			chat_name = EXCLUDED.chat_name,
			last_message_time = EXCLUDED.last_message_time`: `INSERT INTO whatsapp_chats (device_id, chat_jid, chat_name, last_message_time)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			chat_name = VALUES(chat_name),
			last_message_time = VALUES(last_message_time)`,
			
		// Fix other ON CONFLICT patterns
		"ON CONFLICT (device_id, chat_jid)": "ON DUPLICATE KEY UPDATE",
		"EXCLUDED.chat_name": "VALUES(chat_name)",
		"EXCLUDED.last_message_time": "VALUES(last_message_time)",
		"EXCLUDED.is_group": "VALUES(is_group)",
		"EXCLUDED.last_message_text": "VALUES(last_message_text)",
		"EXCLUDED.unread_count": "VALUES(unread_count)",
		"EXCLUDED.avatar_url": "VALUES(avatar_url)",
		"EXCLUDED.updated_at": "VALUES(updated_at)",
		
		// Fix UUID casting syntax
		"::uuid": "",
		"::UUID": "",
		"CAST(? AS UUID)": "?",
		"uuid_generate_v4()": "UUID()",
		"gen_random_uuid()": "UUID()",
	}

	// Files to fix
	filesToFix := []string{
		"src/infrastructure/whatsapp/chat_store.go",
		"src/repository/whatsapp_repository.go",
		"src/repository/optimized_whatsapp_repository.go",
		"src/infrastructure/whatsapp/chat_to_leads.go",
		"src/repository/campaign_repository.go",
		"src/repository/sequence_repository.go",
	}

	for _, file := range filesToFix {
		fixFile(file, fixes)
	}
}

func fixFile(filename string, fixes map[string]string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			return
		}
		fmt.Printf("Error reading %s: %v\n", filename, err)
		return
	}

	originalContent := string(content)
	modifiedContent := originalContent

	// Apply all fixes
	for old, new := range fixes {
		modifiedContent = strings.ReplaceAll(modifiedContent, old, new)
	}

	// Only write if changed
	if modifiedContent != originalContent {
		// Backup original
		backupFile := filename + ".bak"
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
