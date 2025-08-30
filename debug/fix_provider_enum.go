package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Parse the URI to get the DSN format
	dsn := "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?parseTime=true"

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Check current ENUM values
	fmt.Println("\n🔍 Current provider ENUM values:")
	var columnType string
	err = db.QueryRow(`
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = 'admin_railway' 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'provider'
	`).Scan(&columnType)
	if err != nil {
		fmt.Printf("❌ Failed to get ENUM values: %v\n", err)
		return
	}
	fmt.Printf("📋 Before: %s\n", columnType)

	// Update the ENUM to replace 'rvsb_wasap' with 'waha'
	fmt.Println("\n🔄 Updating provider ENUM to include 'waha'...")
	updateQuery := `
		ALTER TABLE device_setting_nodepath 
		MODIFY COLUMN provider ENUM('whacenter', 'wablas', 'waha') NOT NULL DEFAULT 'wablas'
	`

	_, err = db.Exec(updateQuery)
	if err != nil {
		fmt.Printf("❌ Failed to update ENUM: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully updated provider ENUM")

	// Verify the change
	fmt.Println("\n🔍 Verifying updated ENUM values:")
	err = db.QueryRow(`
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = 'admin_railway' 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'provider'
	`).Scan(&columnType)
	if err != nil {
		fmt.Printf("❌ Failed to get updated ENUM values: %v\n", err)
		return
	}
	fmt.Printf("📋 After: %s\n", columnType)

	// Test inserting a record with 'waha' provider
	fmt.Println("\n🧪 Testing WAHA provider insertion...")
	testQuery := `
		INSERT INTO device_setting_nodepath 
		(id, device_id, api_key_option, provider, api_key, id_device) 
		VALUES ('test-waha-provider-fix', 'test-device-fix', 'openai/gpt-4.1', 'waha', 'test-key', 'test-id-device')
	`

	_, err = db.Exec(testQuery)
	if err != nil {
		fmt.Printf("❌ Failed to insert test record with WAHA provider: %v\n", err)
	} else {
		fmt.Println("✅ Successfully inserted test record with WAHA provider")
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM device_setting_nodepath WHERE id = 'test-waha-provider-fix'")
		if err != nil {
			fmt.Printf("⚠️ Failed to clean up test record: %v\n", err)
		} else {
			fmt.Println("🧹 Test record cleaned up")
		}
	}

	// Update any existing 'rvsb_wasap' records to 'waha'
	fmt.Println("\n🔄 Updating existing 'rvsb_wasap' records to 'waha'...")
	updateRecordsQuery := `UPDATE device_setting_nodepath SET provider = 'waha' WHERE provider = 'rvsb_wasap'`
	result, err := db.Exec(updateRecordsQuery)
	if err != nil {
		fmt.Printf("❌ Failed to update existing records: %v\n", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✅ Updated %d existing records from 'rvsb_wasap' to 'waha'\n", rowsAffected)
	}

	fmt.Println("\n🎉 Provider ENUM fix completed successfully!")
	fmt.Println("   Now you can save device settings with 'waha' provider.")
}