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

// decodeAPIKey checks the current API key for FakhriAidilTLW-001
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

	// Query the specific device settings
	query := `SELECT id_device, provider, api_key, api_key_option, instance FROM device_setting_nodepath WHERE id_device = ?`
	row := db.QueryRow(query, "FakhriAidilTLW-001")

	var idDevice, provider, apiKey, apiKeyOption, instance string
	err = row.Scan(&idDevice, &provider, &apiKey, &apiKeyOption, &instance)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ Device FakhriAidilTLW-001 not found in device_setting_nodepath")
		} else {
			log.Fatalf("Failed to query device settings: %v", err)
		}
		return
	}

	fmt.Printf("🔍 Device Settings for %s:\n", idDevice)
	fmt.Printf("   Provider: %s\n", provider)
	fmt.Printf("   API Key Option: %s\n", apiKeyOption)
	fmt.Printf("   Instance: %s\n", instance)
	fmt.Printf("   API Key: %s\n", apiKey)

	// Analyze the API key
	fmt.Println("\n📊 API Key Analysis:")
	fmt.Printf("   Length: %d characters\n", len(apiKey))
	fmt.Printf("   First 10 chars: %s\n", getSafeSubstring(apiKey, 0, 10))
	fmt.Printf("   Last 10 chars: %s\n", getSafeSubstring(apiKey, len(apiKey)-10, len(apiKey)))

	// Check if it's a valid API key format
	if strings.HasPrefix(apiKey, "sk-or-") {
		fmt.Println("   ✅ Valid OpenRouter API key format")
	} else if strings.HasPrefix(apiKey, "sk-proj-") {
		fmt.Println("   ✅ Valid OpenAI API key format")
	} else if strings.HasPrefix(apiKey, "sk-") {
		fmt.Println("   ✅ Valid API key format (generic)")
	} else {
		fmt.Printf("   ❌ Invalid API key format - appears to be: %s\n", apiKey)
	}

	// Simulate the system's API key retrieval logic
	fmt.Println("\n🔄 Simulating System API Key Retrieval Logic:")
	if apiKey != "" {
		fmt.Printf("   ✅ Using device-specific API key: %s...%s\n", 
			getSafeSubstring(apiKey, 0, 8), 
			getSafeSubstring(apiKey, len(apiKey)-8, len(apiKey)))
	} else {
		fmt.Println("   ⚠️  Device API key is empty, would fall back to default")
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