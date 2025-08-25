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

// updateDeviceAPIKey updates the API key for FakhriAidilTLW-001 with a valid OpenRouter key
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

	// Update API key for FakhriAidilTLW-001
	deviceID := "FakhriAidilTLW-001"
	newAPIKey := "sk-or-v1-b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3" // Valid OpenRouter API key format

	fmt.Printf("🔧 Updating API key for device: %s\n", deviceID)
	fmt.Printf("🔑 New API key: %s...\n", newAPIKey[:20]) // Show only first 20 chars for security

	// Update the API key
	updateQuery := `
		UPDATE device_setting_nodepath 
		SET api_key = ?, updated_at = NOW()
		WHERE id_device = ?
	`

	result, err := db.Exec(updateQuery, newAPIKey, deviceID)
	if err != nil {
		log.Fatalf("Failed to update API key: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		fmt.Println("❌ No device found with id_device = FakhriAidilTLW-001")
		return
	}

	fmt.Printf("✅ Successfully updated API key for device FakhriAidilTLW-001\n")
	fmt.Printf("   Rows affected: %d\n", rowsAffected)

	// Verify the update
	verifyQuery := `
		SELECT api_key, api_key_option, provider 
		FROM device_setting_nodepath 
		WHERE id_device = ?
	`

	var apiKey, apiKeyOption, provider sql.NullString
	err = db.QueryRow(verifyQuery, "FakhriAidilTLW-001").Scan(&apiKey, &apiKeyOption, &provider)
	if err != nil {
		log.Printf("Failed to verify update: %v", err)
		return
	}

	fmt.Println("\n📋 Updated Device Configuration:")
	fmt.Printf("Provider: %s\n", getStringValue(provider))
	fmt.Printf("Model: %s\n", getStringValue(apiKeyOption))
	fmt.Printf("API Key: %s\n", getStringValue(apiKey))

	fmt.Println("\n🔍 Next Steps:")
	fmt.Println("1. Replace the placeholder API key with a valid OpenRouter API key")
	fmt.Println("2. Test the AI prompt functionality again")
	fmt.Println("3. Check server logs for authentication success")
}

// getStringValue safely extracts string value from sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "(null)"
}