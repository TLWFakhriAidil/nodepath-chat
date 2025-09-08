package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

/**
 * Debug script to check device analytics issue
 * Investigates why Analytics page shows "No devices available"
 */
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Fallback to MYSQL_URI if DATABASE_URL is not set
		databaseURL = os.Getenv("MYSQL_URI")
	}
	if databaseURL == "" {
		log.Fatal("DATABASE_URL or MYSQL_URI environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✅ Database connection successful")

	// Check total device count
	fmt.Println("\n=== DEVICE ANALYTICS DEBUG ===")
	checkTotalDevices(db)
	checkDevicesByUser(db)
	checkUserTable(db)
	checkDeviceSettingsStructure(db)
}

/**
 * Check total number of devices in device_setting_nodepath table
 */
func checkTotalDevices(db *sql.DB) {
	fmt.Println("\n1. Checking total devices in device_setting_nodepath table:")
	
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath").Scan(&totalCount)
	if err != nil {
		log.Printf("❌ Error counting total devices: %v", err)
		return
	}
	fmt.Printf("   Total devices in table: %d\n", totalCount)

	// Check devices with valid id_device
	var validDeviceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device IS NOT NULL AND id_device != ''").Scan(&validDeviceCount)
	if err != nil {
		log.Printf("❌ Error counting valid devices: %v", err)
		return
	}
	fmt.Printf("   Devices with valid id_device: %d\n", validDeviceCount)

	// Show sample devices
	if validDeviceCount > 0 {
		fmt.Println("   Sample devices:")
		rows, err := db.Query("SELECT id_device, user_id, provider FROM device_setting_nodepath WHERE id_device IS NOT NULL AND id_device != '' LIMIT 5")
		if err != nil {
			log.Printf("❌ Error fetching sample devices: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var idDevice string
			var userID sql.NullInt32
			var provider sql.NullString
			err := rows.Scan(&idDevice, &userID, &provider)
			if err != nil {
				log.Printf("❌ Error scanning device: %v", err)
				continue
			}
			userIDStr := "NULL"
			if userID.Valid {
				userIDStr = fmt.Sprintf("%d", userID.Int32)
			}
			providerStr := "NULL"
			if provider.Valid {
				providerStr = provider.String
			}
			fmt.Printf("     - Device: %s, User ID: %s, Provider: %s\n", idDevice, userIDStr, providerStr)
		}
	}
}

/**
 * Check devices grouped by user_id
 */
func checkDevicesByUser(db *sql.DB) {
	fmt.Println("\n2. Checking devices by user_id:")
	
	rows, err := db.Query(`
		SELECT 
			COALESCE(user_id, -1) as user_id,
			COUNT(*) as device_count,
			GROUP_CONCAT(id_device) as device_list
		FROM device_setting_nodepath 
		WHERE id_device IS NOT NULL AND id_device != ''
		GROUP BY user_id
		ORDER BY device_count DESC
	`)
	if err != nil {
		log.Printf("❌ Error querying devices by user: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var deviceCount int
		var deviceList sql.NullString
		err := rows.Scan(&userID, &deviceCount, &deviceList)
		if err != nil {
			log.Printf("❌ Error scanning user devices: %v", err)
			continue
		}
		userIDStr := fmt.Sprintf("%d", userID)
		if userID == -1 {
			userIDStr = "NULL"
		}
		devices := "NULL"
		if deviceList.Valid {
			devices = deviceList.String
		}
		fmt.Printf("   User ID: %s, Device Count: %d, Devices: %s\n", userIDStr, deviceCount, devices)
	}
}

/**
 * Check users table to understand user IDs
 */
func checkUserTable(db *sql.DB) {
	fmt.Println("\n3. Checking users table:")
	
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users_nodepath").Scan(&userCount)
	if err != nil {
		log.Printf("❌ Error counting users: %v", err)
		return
	}
	fmt.Printf("   Total users: %d\n", userCount)

	// Show sample users
	if userCount > 0 {
		fmt.Println("   Sample users:")
		rows, err := db.Query("SELECT id, email FROM users_nodepath LIMIT 5")
		if err != nil {
			log.Printf("❌ Error fetching sample users: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var email string
			err := rows.Scan(&id, &email)
			if err != nil {
				log.Printf("❌ Error scanning user: %v", err)
				continue
			}
			fmt.Printf("     - User ID: %d, Email: %s\n", id, email)
		}
	}
}

/**
 * Check device_setting_nodepath table structure
 */
func checkDeviceSettingsStructure(db *sql.DB) {
	fmt.Println("\n4. Checking device_setting_nodepath table structure:")
	
	rows, err := db.Query("DESCRIBE device_setting_nodepath")
	if err != nil {
		log.Printf("❌ Error describing table: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("   Table columns:")
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Printf("❌ Error scanning column info: %v", err)
			continue
		}
		fmt.Printf("     - %s: %s (Null: %s, Key: %s)\n", field, fieldType, null, key)
	}
}