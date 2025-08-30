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
	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	// Convert mysql:// to proper DSN format
	dsn := mysqlURI
	if len(dsn) > 8 && dsn[:8] == "mysql://" {
		// mysql://user:pass@host:port/db -> user:pass@tcp(host:port)/db
		dsn = dsn[8:] // Remove mysql:// prefix
		// Find the @ symbol to separate user:pass from host:port/db
		if atIndex := strings.Index(dsn, "@"); atIndex != -1 {
			userPass := dsn[:atIndex]
			hostPortDb := dsn[atIndex+1:]
			// Find the / to separate host:port from db
			if slashIndex := strings.Index(hostPortDb, "/"); slashIndex != -1 {
				hostPort := hostPortDb[:slashIndex]
				dbName := hostPortDb[slashIndex+1:]
				dsn = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, dbName)
			}
		}
	}

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Device details
	idDevice := "04bbcc4-026d-4034-b82f-147ff3ca13bd"
	sessionName := fmt.Sprintf("user_%s", idDevice)
	webhookURL := fmt.Sprintf("https://nodepath-chat-production.up.railway.app/api/webhook/%s/%s", idDevice, sessionName)
	phoneNumber := "60123456789" // Default phone number
	apiKey := "sk-proj-example-api-key" // Default OpenRouter API key

	// Check if device already exists
	var existingID string
	err = db.QueryRow("SELECT id FROM device_setting_nodepath WHERE id_device = ?", idDevice).Scan(&existingID)
	if err == nil {
		fmt.Printf("⚠️ Device with ID '%s' already exists (ID: %s)\n", idDevice, existingID)
		return
	} else if err != sql.ErrNoRows {
		log.Fatalf("Error checking existing device: %v", err)
	}

	// Generate unique ID for the device setting
	id := fmt.Sprintf("waha_%d", time.Now().Unix())
	now := time.Now()

	// Insert new WAHA device
	query := `
		INSERT INTO device_setting_nodepath 
		(id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, id_device, id_erp, id_admin, instance, created_at, updated_at)
		VALUES (?, NULL, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, ?, ?)
	`

	_, err = db.Exec(query,
		id,                    // id
		"openai/gpt-4.1",     // api_key_option
		webhookURL,           // webhook_id
		"waha",               // provider
		phoneNumber,          // phone_number
		apiKey,               // api_key
		idDevice,             // id_device
		sessionName,          // instance
		now,                  // created_at
		now,                  // updated_at
	)

	if err != nil {
		log.Fatalf("Failed to insert WAHA device: %v", err)
	}

	fmt.Printf("✅ Successfully created WAHA device:\n")
	fmt.Printf("   ID Device: %s\n", idDevice)
	fmt.Printf("   Session Name: %s\n", sessionName)
	fmt.Printf("   Webhook URL: %s\n", webhookURL)
	fmt.Printf("   Provider: waha\n")
	fmt.Printf("   Phone Number: %s\n", phoneNumber)

	// Verify the device was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device = ? AND provider = 'waha'", idDevice).Scan(&count)
	if err != nil {
		log.Printf("Warning: Could not verify device creation: %v", err)
	} else if count > 0 {
		fmt.Printf("✅ Device verification successful - found %d WAHA device(s) with ID '%s'\n", count, idDevice)
	} else {
		fmt.Printf("❌ Device verification failed - no WAHA device found with ID '%s'\n", idDevice)
	}
}