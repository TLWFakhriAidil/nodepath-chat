package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("DATABASE_URL")
	if mysqlURI == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
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

	fmt.Println("=== Checking API Key for TEST-DEVICE ===")

	// Check API key for TEST-DEVICE
	query := `SELECT id_device, api_key, api_key_option, provider, instance FROM device_setting_nodepath WHERE id_device = ?`
	row := db.QueryRow(query, "TEST-DEVICE")

	var idDevice, apiKey, apiKeyOption, provider, instance string
	err = row.Scan(&idDevice, &apiKey, &apiKeyOption, &provider, &instance)
	if err != nil {
		log.Fatalf("Failed to scan device settings: %v", err)
	}

	fmt.Printf("Device: %s\n", idDevice)
	fmt.Printf("API Key: '%s' (length: %d)\n", apiKey, len(apiKey))
	fmt.Printf("API Key Option: %s\n", apiKeyOption)
	fmt.Printf("Provider: %s\n", provider)
	fmt.Printf("Instance: %s\n", instance)

	if apiKey == "" {
		fmt.Println("\n❌ API Key is empty! This is why authentication is failing.")
		fmt.Println("\n=== Updating API Key for TEST-DEVICE ===")
		
		// Update with a test API key
		updateQuery := `UPDATE device_setting_nodepath SET api_key = ? WHERE id_device = ?`
		testAPIKey := "test-api-key-12345"
		_, err = db.Exec(updateQuery, testAPIKey, "TEST-DEVICE")
		if err != nil {
			log.Fatalf("Failed to update API key: %v", err)
		}
		
		fmt.Printf("✅ Updated API key to: %s\n", testAPIKey)
	} else {
		fmt.Println("\n✅ API Key is set.")
	}

	fmt.Println("\nCheck completed!")
}