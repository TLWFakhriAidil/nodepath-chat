package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database URL from environment or use default
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Parse MySQL URI to Go driver format
	if mysqlURI[:8] == "mysql://" {
		// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
		mysqlURI = mysqlURI[8:] // Remove mysql:// prefix
		// Replace @ with @tcp( and add ) before /
		parts := strings.Split(mysqlURI, "/")
		if len(parts) >= 2 {
			connPart := parts[0]
			dbName := parts[1]
			// Split user:pass@host:port
			atIndex := strings.LastIndex(connPart, "@")
			if atIndex > 0 {
				userPass := connPart[:atIndex]
				hostPort := connPart[atIndex+1:]
				mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, dbName)
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

	fmt.Println("Connected to database successfully")

	// Create test device with valid enum values
	fmt.Println("\n=== Creating test device ===")
	now := time.Now()
	testQuery := `
		INSERT INTO device_setting_nodepath (
			id, id_device, api_key_option, provider, api_key, instance, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			api_key_option = VALUES(api_key_option),
			provider = VALUES(provider),
			api_key = VALUES(api_key),
			instance = VALUES(instance),
			updated_at = VALUES(updated_at)
	`

	testID := "test-device-001"
	testIDDevice := "TEST-DEVICE"
	testAPIKeyOption := "openai/gpt-4.1" // Valid enum value
	testProvider := "wablas"
	testAPIKey := "test-api-key-12345"
	testInstance := "test-instance"

	_, err = db.Exec(testQuery, testID, testIDDevice, testAPIKeyOption, testProvider, testAPIKey, testInstance, now, now)
	if err != nil {
		log.Printf("Failed to create test device: %v", err)
	} else {
		fmt.Println("Test device created successfully!")
		fmt.Printf("Device ID: %s\n", testIDDevice)
		fmt.Printf("API Key Option: %s\n", testAPIKeyOption)
		fmt.Printf("Provider: %s\n", testProvider)
		fmt.Printf("Instance: %s\n", testInstance)
	}

	// Verify the device was created
	fmt.Println("\n=== Verifying test device ===")
	verifyQuery := "SELECT id, id_device, api_key_option, provider, instance FROM device_setting_nodepath WHERE id_device = ?"
	row := db.QueryRow(verifyQuery, testIDDevice)
	
	var id, idDevice, apiKeyOption, provider, instance string
	err = row.Scan(&id, &idDevice, &apiKeyOption, &provider, &instance)
	if err != nil {
		log.Printf("Failed to verify device: %v", err)
	} else {
		fmt.Printf("✅ Device verified: ID=%s, IDDevice=%s, APIKeyOption=%s, Provider=%s, Instance=%s\n", 
			id, idDevice, apiKeyOption, provider, instance)
	}

	fmt.Println("\nSetup completed!")
}