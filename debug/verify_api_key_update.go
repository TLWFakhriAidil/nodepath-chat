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

// verifyAPIKeyUpdate checks if the API key has been correctly updated in the database
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
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
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

	// Query device settings for FakhriAidilTLW-001
	deviceID := "FakhriAidilTLW-001"
	query := `SELECT id_device, provider, instance, api_key, api_key_option FROM device_settings_nodepath WHERE id_device = ?`

	var idDevice, provider, instance, apiKey, apiKeyOption sql.NullString
	err = db.QueryRow(query, deviceID).Scan(&idDevice, &provider, &instance, &apiKey, &apiKeyOption)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("❌ No device settings found for device: %s\n", deviceID)
			return
		}
		log.Fatalf("Failed to query device settings: %v", err)
	}

	fmt.Printf("📱 Device Settings for %s:\n", deviceID)
	fmt.Printf("   ID Device: %s\n", getStringValue(idDevice))
	fmt.Printf("   Provider: %s\n", getStringValue(provider))
	fmt.Printf("   Instance: %s\n", getStringValue(instance))
	fmt.Printf("   API Key: %s\n", getStringValue(apiKey))
	fmt.Printf("   API Key Option: %s\n", getStringValue(apiKeyOption))
	fmt.Println("===========================================")

	// Analyze API key
	apiKeyValue := getStringValue(apiKey)
	isValidAPIKey := apiKey.Valid && apiKeyValue != "" && !strings.HasPrefix(apiKeyValue, "sk-test")
	isProviderName := apiKeyValue == "whacenter" || apiKeyValue == "wablas"
	isOpenRouterKey := strings.HasPrefix(apiKeyValue, "sk-or-")
	isOpenAIKey := strings.HasPrefix(apiKeyValue, "sk-proj-") || (strings.HasPrefix(apiKeyValue, "sk-") && !strings.HasPrefix(apiKeyValue, "sk-or-") && !strings.HasPrefix(apiKeyValue, "sk-test"))

	fmt.Printf("🔍 API Key Analysis:\n")
	fmt.Printf("   Is Valid: %t\n", apiKey.Valid)
	fmt.Printf("   Is Empty: %t\n", apiKeyValue == "")
	fmt.Printf("   Is Test Key: %t\n", strings.HasPrefix(apiKeyValue, "sk-test"))
	fmt.Printf("   Is Provider Name: %t\n", isProviderName)
	fmt.Printf("   Is OpenRouter Key: %t\n", isOpenRouterKey)
	fmt.Printf("   Is OpenAI Key: %t\n", isOpenAIKey)
	fmt.Printf("   Is Valid API Key: %t\n", isValidAPIKey)
	fmt.Println("===========================================")

	// Provide recommendations
	if isProviderName {
		fmt.Printf("❌ ISSUE FOUND: API key contains provider name '%s' instead of actual API key\n", apiKeyValue)
		fmt.Println("💡 SOLUTION: Update the api_key column with a valid OpenRouter or OpenAI API key")
	} else if isValidAPIKey {
		fmt.Println("✅ API key appears to be valid!")
		if isOpenRouterKey {
			fmt.Println("📡 Detected: OpenRouter API key")
		} else if isOpenAIKey {
			fmt.Println("🤖 Detected: OpenAI API key")
		}
	} else {
		fmt.Println("⚠️  API key validation failed - please check the format")
	}

	fmt.Println("===========================================")
	fmt.Println("🔄 Testing API key retrieval logic...")

	// Simulate the logic from ai_whatsapp_service.go
	if isValidAPIKey {
		fmt.Printf("✅ System would use device-specific API key: %s\n", maskAPIKey(apiKeyValue))
	} else {
		fmt.Println("⚠️  System would fall back to default OpenRouter key")
	}
}

// getStringValue safely extracts string from sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}

// maskAPIKey masks API key for safe display (similar to the one in the service)
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}