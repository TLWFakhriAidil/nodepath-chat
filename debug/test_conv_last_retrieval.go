package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("=== Testing conv_last Data Retrieval ===")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Connect to database
	db, err := sql.Open("mysql", "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connected successfully")

	// Test prospect number
	prospectNum := "60179645043"
	idDevice := "FakhriAidilTLW-001"

	fmt.Printf("🔍 Looking for AI conversation: %s (device: %s)\n", prospectNum, idDevice)

	// Query conv_last directly
	query := `SELECT id_prospect, prospect_num, conv_last FROM ai_whatsapp_nodepath WHERE prospect_num = ? AND id_device = ?`
	var idProspect int
	var retrievedProspectNum string
	var convLast sql.NullString

	err = db.QueryRow(query, prospectNum, idDevice).Scan(&idProspect, &retrievedProspectNum, &convLast)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No AI conversation found")
			return
		}
		log.Fatalf("Failed to query AI conversation: %v", err)
	}

	fmt.Printf("✅ AI conversation found (ID: %d)\n", idProspect)

	// Display conv_last data
	fmt.Println("\n📋 Conv_last data:")
	if !convLast.Valid || convLast.String == "" || convLast.String == "null" {
		fmt.Println("   (null or empty)")
	} else {
		fmt.Printf("   %s\n", convLast.String)
	}

	// Test getLastAIResponse function logic
	fmt.Println("\n🤖 Testing getLastAIResponse logic:")
	
	lastResponse := getLastAIResponseTest(convLast.String)
	if lastResponse == "" {
		fmt.Println("   (no last AI response found)")
	} else {
		fmt.Printf("   Last AI Response: %s\n", lastResponse)
	}

	fmt.Println("\n=== Test completed ===")
}

// getLastAIResponseTest simulates the getLastAIResponse function logic
func getLastAIResponseTest(convLastStr string) string {
	if convLastStr == "" || convLastStr == "null" {
		return ""
	}

	// Try to parse as JSON first (for backward compatibility)
	var testJSON interface{}
	if err := json.Unmarshal([]byte(convLastStr), &testJSON); err == nil {
		// It's valid JSON, try to extract bot response
		if jsonMap, ok := testJSON.(map[string]interface{}); ok {
			if response, exists := jsonMap["Response"]; exists {
				if responseArray, ok := response.([]interface{}); ok {
					for _, item := range responseArray {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if itemType, exists := itemMap["type"]; exists && itemType == "text" {
								if content, exists := itemMap["content"]; exists {
									if contentStr, ok := content.(string); ok {
										return contentStr
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		// It's plain text format, parse for BOT: entries
		lines := strings.Split(convLastStr, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "BOT:") {
				return strings.TrimPrefix(line, "BOT:")
			}
		}
	}

	return ""
}