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
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
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

	// Check device_setting_nodepath table structure
	fmt.Println("\n=== Checking device_setting_nodepath table structure ===")
	rows, err := db.Query("DESCRIBE device_setting_nodepath")
	if err != nil {
		log.Printf("Error describing table: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-8s %-8s %-15s %s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Printf("%-20s %-20s %-8s %-8s %-15s %s\n", "-----", "----", "----", "---", "-------", "-----")

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("%-20s %-20s %-8s %-8s %-15s %s\n", field, fieldType, null, key, defaultVal, extra)
	}

	// Check existing devices
	fmt.Println("\n=== Existing devices ===")
	deviceRows, err := db.Query("SELECT id_device, api_key_option, provider FROM device_setting_nodepath LIMIT 5")
	if err != nil {
		log.Printf("Error querying devices: %v", err)
	} else {
		defer deviceRows.Close()
		for deviceRows.Next() {
			var idDevice, apiKeyOption, provider string
			err := deviceRows.Scan(&idDevice, &apiKeyOption, &provider)
			if err != nil {
				log.Printf("Error scanning device: %v", err)
				continue
			}
			fmt.Printf("Device: %s, API Key Option: %s, Provider: %s\n", idDevice, apiKeyOption, provider)
		}
	}

	// Try to create a simple test device
	fmt.Println("\n=== Creating test device ===")
	now := time.Now()
	testQuery := `
		INSERT INTO device_setting_nodepath (
			id, id_device, api_key_option, provider, api_key, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			api_key_option = VALUES(api_key_option),
			provider = VALUES(provider),
			api_key = VALUES(api_key),
			updated_at = VALUES(updated_at)
	`

	testID := "test-device-001"
	testIDDevice := "TEST-DEVICE"
	testAPIKeyOption := "gpt-4"
	testProvider := "wablas"
	testAPIKey := "test-api-key"

	_, err = db.Exec(testQuery, testID, testIDDevice, testAPIKeyOption, testProvider, testAPIKey, now, now)
	if err != nil {
		log.Printf("Failed to create test device: %v", err)
	} else {
		fmt.Println("Test device created successfully!")
	}

	fmt.Println("\nCheck completed!")
}