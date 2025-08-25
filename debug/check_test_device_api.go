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

	fmt.Println("=== Checking API Key for FakhriAidilTLW-001 ===")

	// Check API key for FakhriAidilTLW-001
	query := `SELECT id_device, api_key, api_key_option, provider, instance FROM device_setting_nodepath WHERE id_device = ?`
	row := db.QueryRow(query, "FakhriAidilTLW-001")

	var idDevice, apiKey, apiKeyOption, provider, instance sql.NullString
	err = row.Scan(&idDevice, &apiKey, &apiKeyOption, &provider, &instance)
	if err != nil {
		log.Fatalf("Failed to scan device settings: %v", err)
	}

	fmt.Printf("Device ID: %s\n", idDevice.String)
	fmt.Printf("API Key: %s\n", apiKey.String)
	fmt.Printf("API Key Option: %s\n", apiKeyOption.String)
	fmt.Printf("Provider: %s\n", provider.String)
	fmt.Printf("Instance: %s\n", instance.String)

	// Check if API key looks like a provider name
	if apiKey.Valid {
		apiKeyValue := apiKey.String
		if apiKeyValue == "whacenter" || apiKeyValue == "wablas" {
			fmt.Printf("\n❌ ISSUE: API key contains provider name '%s' instead of actual API key\n", apiKeyValue)
		} else if apiKeyValue != "" {
			fmt.Printf("\n✅ API key appears to be valid (not a provider name)\n")
			fmt.Printf("Full API Key: %s\n", apiKeyValue)
		} else {
			fmt.Printf("\n⚠️  API key is empty\n")
		}
	} else {
		fmt.Printf("\n⚠️  API key is NULL\n")
	}
}