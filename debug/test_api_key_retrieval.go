package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// testAPIKeyRetrieval simulates the exact logic used in ai_whatsapp_service.go
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		// Use the remote database connection as fallback
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		// Remove mysql:// prefix and convert to proper DSN format
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		// Format: user:password@host:port/database
		// Convert to: user:password@tcp(host:port)/database
		parts := strings.Split(mysqlURI, "/")
		if len(parts) >= 2 {
			userHostPart := parts[0]
			databasePart := parts[1]
			// Split user:password@host:port
			atIndex := strings.LastIndex(userHostPart, "@")
			if atIndex != -1 {
				userPass := userHostPart[:atIndex]
				hostPort := userHostPart[atIndex+1:]
				mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, databasePart)
			}
		}
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully!")
	fmt.Println("===========================================")

	// Test the exact logic from ai_whatsapp_service.go
	idDevice := "FakhriAidilTLW-001"
	fmt.Printf("🧪 Testing API key retrieval logic for device: %s\n\n", idDevice)

	// Step 1: Get device settings (simulating the service call)
	query := `SELECT api_key FROM device_setting_nodepath WHERE id_device = ?`
	var apiKeyFromDB sql.NullString
	err = db.QueryRow(query, idDevice).Scan(&apiKeyFromDB)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("❌ Device %s not found in database\n", idDevice)
			return
		}
		log.Fatalf("Failed to query device settings: %v", err)
	}

	fmt.Printf("📊 Database Query Result:\n")
	fmt.Printf("   API Key Valid: %t\n", apiKeyFromDB.Valid)
	if apiKeyFromDB.Valid {
		fmt.Printf("   API Key Value: %s\n", apiKeyFromDB.String)
		fmt.Printf("   API Key Length: %d\n", len(apiKeyFromDB.String))
	} else {
		fmt.Printf("   API Key Value: NULL\n")
	}

	// Step 2: Apply the exact logic from ai_whatsapp_service.go
	fmt.Printf("\n🔄 Applying AI WhatsApp Service Logic:\n")

	// Check if device has a valid API key (not empty and not a test key)
	isValidAPIKey := apiKeyFromDB.Valid && 
		apiKeyFromDB.String != "" && 
		!strings.HasPrefix(apiKeyFromDB.String, "sk-test")

	fmt.Printf("   Valid API Key Check: %t\n", isValidAPIKey)
	fmt.Printf("   - Is Valid: %t\n", apiKeyFromDB.Valid)
	fmt.Printf("   - Not Empty: %t\n", apiKeyFromDB.String != "")
	fmt.Printf("   - Not Test Key: %t\n", !strings.HasPrefix(apiKeyFromDB.String, "sk-test"))

	var finalAPIKey string
	var apiKeySource string

	if isValidAPIKey {
		finalAPIKey = apiKeyFromDB.String
		apiKeySource = "device_settings"
		fmt.Printf("   ✅ Using device-specific API key\n")
	} else {
		// Use default OpenRouter key for non-special devices
		if idDevice != "SCHQ-S94" && idDevice != "SCHQ-S12" {
			// Simulate getting default key from config
			defaultKey := os.Getenv("OPENROUTER_DEFAULT_KEY")
			if defaultKey == "" {
				defaultKey = "[DEFAULT_OPENROUTER_KEY_FROM_CONFIG]"
			}
			finalAPIKey = defaultKey
			apiKeySource = "default_openrouter"
			fmt.Printf("   ⚠️  Using default OpenRouter API key\n")
		} else {
			// For special devices, use the hardcoded OpenAI key
			finalAPIKey = "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"
			apiKeySource = "hardcoded_openai"
			fmt.Printf("   🔧 Using hardcoded OpenAI API key for special device\n")
		}
	}

	fmt.Printf("\n📋 Final Result:\n")
	fmt.Printf("   API Key Source: %s\n", apiKeySource)
	fmt.Printf("   API Key Preview: %s...%s\n", 
		getSafeSubstring(finalAPIKey, 0, 8), 
		getSafeSubstring(finalAPIKey, len(finalAPIKey)-8, len(finalAPIKey)))
	fmt.Printf("   API Key Length: %d\n", len(finalAPIKey))

	// Validate the final API key format
	fmt.Printf("\n🔍 API Key Validation:\n")
	if strings.HasPrefix(finalAPIKey, "sk-or-") {
		fmt.Printf("   ✅ Valid OpenRouter API key format\n")
	} else if strings.HasPrefix(finalAPIKey, "sk-proj-") {
		fmt.Printf("   ✅ Valid OpenAI API key format\n")
	} else if strings.HasPrefix(finalAPIKey, "sk-") {
		fmt.Printf("   ✅ Valid API key format (generic)\n")
	} else {
		fmt.Printf("   ❌ Invalid API key format: %s\n", finalAPIKey)
	}

	fmt.Printf("\n🎯 Conclusion:\n")
	if isValidAPIKey && (strings.HasPrefix(finalAPIKey, "sk-or-") || strings.HasPrefix(finalAPIKey, "sk-proj-")) {
		fmt.Printf("   ✅ API key retrieval is working correctly!\n")
		fmt.Printf("   ✅ The system should now use the correct API key for AI requests.\n")
	} else {
		fmt.Printf("   ❌ There may still be issues with API key configuration.\n")
	}
}

// getSafeSubstring safely extracts substring without panicking
func getSafeSubstring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return ""
	}
	return s[start:end]
}