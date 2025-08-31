package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable not set")
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	// Generate a UUID for the id column
	id := uuid.New().String()
	idDevice := "WAHA-TEST-001"
	provider := "waha"
	phoneNumber := "60123456789"
	instance := "user_WAHA-TEST-001"
	apiKey := "test-api-key"
	apiKeyOption := "anthropic/claude-3.5-sonnet"

	// Insert test device
	query := `INSERT INTO device_setting_nodepath 
			  (id, id_device, provider, phone_number, instance, api_key, api_key_option) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = db.Exec(query, id, idDevice, provider, phoneNumber, instance, apiKey, apiKeyOption)
	if err != nil {
		log.Fatalf("Failed to insert test device: %v", err)
	}

	fmt.Printf("✅ Created test WAHA device:\n")
	fmt.Printf("   ID: %s\n", id)
	fmt.Printf("   ID_DEVICE: %s\n", idDevice)
	fmt.Printf("   PROVIDER: %s\n", provider)
	fmt.Printf("   PHONE_NUMBER: %s\n", phoneNumber)
	fmt.Printf("   INSTANCE: %s\n", instance)
	fmt.Printf("\n💡 You can now test the WAHA status endpoint with:\n")
	fmt.Printf("   http://localhost:8080/api/device-settings/%s/waha-status\n", id)
}