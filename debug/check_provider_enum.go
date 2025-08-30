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

	// Check the table structure
	fmt.Println("\n📋 Checking device_setting_nodepath table structure...")
	rows, err := db.Query("DESCRIBE device_setting_nodepath")
	if err != nil {
		log.Fatal("Failed to describe table:", err)
	}
	defer rows.Close()

	fmt.Printf("%-15s %-50s %-10s %-10s %-10s %-10s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println("=================================================================================")

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("%-15s %-50s %-10s %-10s %-10s %-10s\n",
			field.String,
			fieldType.String,
			null.String,
			key.String,
			defaultVal.String,
			extra.String)

		// Check if this is the provider field
		if field.String == "provider" {
			fmt.Printf("\n🔍 Provider field details:\n")
			fmt.Printf("   Type: %s\n", fieldType.String)
			fmt.Printf("   Default: %s\n", defaultVal.String)
		}
	}

	// Check if we can insert a test record with 'waha' provider
	fmt.Println("\n🧪 Testing WAHA provider insertion...")
	testQuery := `
		INSERT INTO device_setting_nodepath 
		(id, device_id, api_key_option, provider, api_key, id_device) 
		VALUES ('test-waha-provider', 'test-device', 'openai/gpt-4.1', 'waha', 'test-key', 'test-id-device')
	`

	_, err = db.Exec(testQuery)
	if err != nil {
		fmt.Printf("❌ Failed to insert test record with WAHA provider: %v\n", err)
		fmt.Println("   This confirms the ENUM constraint issue!")
	} else {
		fmt.Println("✅ Successfully inserted test record with WAHA provider")
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM device_setting_nodepath WHERE id = 'test-waha-provider'")
		if err != nil {
			fmt.Printf("⚠️ Failed to clean up test record: %v\n", err)
		} else {
			fmt.Println("🧹 Test record cleaned up")
		}
	}

	// Check current ENUM values
	fmt.Println("\n🔍 Checking current ENUM values for provider column...")
	enumQuery := `
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = 'admin_railway' 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'provider'
	`

	var columnType string
	err = db.QueryRow(enumQuery).Scan(&columnType)
	if err != nil {
		fmt.Printf("❌ Failed to get ENUM values: %v\n", err)
	} else {
		fmt.Printf("📋 Current provider column type: %s\n", columnType)
	}
}