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

// checkDeviceAPIKey checks the API key and model configuration for FakhriAidilTLW-001
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
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

	fmt.Println("🔍 Checking device API key configuration for FakhriAidilTLW-001...")

	// Query device settings
	query := `
		SELECT id, device_id, api_key_option, provider, api_key, id_device, instance, webhook_id
		FROM device_setting_nodepath 
		WHERE id_device = ?
	`

	var id, deviceID, apiKeyOption, provider, apiKey, idDevice, instance, webhookID sql.NullString

	err = db.QueryRow(query, "FakhriAidilTLW-001").Scan(
		&id, &deviceID, &apiKeyOption, &provider, &apiKey, &idDevice, &instance, &webhookID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ Device FakhriAidilTLW-001 not found in device_setting_nodepath table")
			fmt.Println("\n🔧 Creating default device settings...")
			
			// Create default device settings
			createQuery := `
				INSERT INTO device_setting_nodepath 
				(id, id_device, api_key_option, provider, api_key, instance, created_at, updated_at)
				VALUES (UUID(), ?, ?, ?, ?, ?, NOW(), NOW())
			`
			
			// Use a valid OpenRouter API key for testing
			testAPIKey := "sk-or-v1-test-key-placeholder"
			
			_, err = db.Exec(createQuery, "FakhriAidilTLW-001", "openai/gpt-4o-mini", "whacenter", testAPIKey, "test-instance-001")
			if err != nil {
				log.Printf("Failed to create device settings: %v", err)
				return
			}
			
			fmt.Println("✅ Default device settings created")
			fmt.Println("⚠️  Please update the API key with a valid OpenRouter key")
			return
		}
		log.Printf("Error querying device: %v", err)
		return
	}

	fmt.Println("\n📋 Device Configuration:")
	fmt.Printf("ID: %s\n", getStringValue(id))
	fmt.Printf("Device ID: %s\n", getStringValue(deviceID))
	fmt.Printf("ID Device: %s\n", getStringValue(idDevice))
	fmt.Printf("Provider: %s\n", getStringValue(provider))
	fmt.Printf("Instance: %s\n", getStringValue(instance))
	fmt.Printf("Webhook ID: %s\n", getStringValue(webhookID))
	fmt.Printf("API Key Option (Model): %s\n", getStringValue(apiKeyOption))
	
	// Show API key preview (first 10 and last 4 characters)
	apiKeyStr := getStringValue(apiKey)
	if apiKeyStr != "" {
		if len(apiKeyStr) > 14 {
			apiKeyPreview := apiKeyStr[:10] + "***" + apiKeyStr[len(apiKeyStr)-4:]
			fmt.Printf("API Key: %s\n", apiKeyPreview)
		} else {
			fmt.Printf("API Key: %s\n", strings.Repeat("*", len(apiKeyStr)))
		}
	} else {
		fmt.Println("API Key: ❌ NOT SET")
	}

	// Check if this is a special device
	if getStringValue(idDevice) == "SCHQ-S94" || getStringValue(idDevice) == "SCHQ-S12" {
		fmt.Println("\n🔧 Special Device Configuration:")
		fmt.Println("API URL: https://api.openai.com/v1/chat/completions")
		fmt.Println("Model: gpt-4")
	} else {
		fmt.Println("\n🔧 Standard Device Configuration:")
		fmt.Println("API URL: https://openrouter.ai/api/v1/chat/completions")
		fmt.Printf("Model: %s\n", getStringValue(apiKeyOption))
	}

	// Validate API key format
	apiKeyValue := getStringValue(apiKey)
	if apiKeyValue == "" {
		fmt.Println("\n❌ ISSUE: API key is empty")
	} else if strings.HasPrefix(apiKeyValue, "sk-or-") {
		fmt.Println("\n✅ API key format looks like OpenRouter")
	} else if strings.HasPrefix(apiKeyValue, "sk-proj-") {
		fmt.Println("\n✅ API key format looks like OpenAI")
	} else {
		fmt.Println("\n⚠️  API key format is unrecognized")
	}

	fmt.Println("\n🔍 Check completed!")
}

// getStringValue safely extracts string value from sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "(null)"
}