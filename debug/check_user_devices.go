package main

import (
	"database/sql"
	"fmt"
	"hash/fnv"
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

	// Get database connection string from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🔍 Checking user devices association...")

	// Check all devices and their user_id associations
	query := `SELECT id, id_device, user_id, provider FROM device_setting_nodepath ORDER BY id_device`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-40s %-25s %-10s %-12s\n", "ID", "ID_DEVICE", "USER_ID", "PROVIDER")
	fmt.Println("====================================================================================")

	totalCount := 0
	devicesWithUser := 0
	devicesWithoutUser := 0

	for rows.Next() {
		var id string
		var idDevice, provider sql.NullString
		var userID sql.NullInt32
		
		err := rows.Scan(&id, &idDevice, &userID, &provider)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		totalCount++
		
		userIDStr := "NULL"
		if userID.Valid {
			userIDStr = fmt.Sprintf("%d", userID.Int32)
			devicesWithUser++
		} else {
			devicesWithoutUser++
		}

		fmt.Printf("%-40s %-25s %-10s %-12s\n", 
			id,
			getStringValue(idDevice), 
			userIDStr,
			getStringValue(provider))
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total devices: %d\n", totalCount)
	fmt.Printf("   Devices with user_id: %d\n", devicesWithUser)
	fmt.Printf("   Devices without user_id: %d\n", devicesWithoutUser)

	// Check users table to see what users exist
	fmt.Printf("\n👥 Checking users table:\n")
	userQuery := `SELECT id, name, email FROM users ORDER BY id`
	userRows, err := db.Query(userQuery)
	if err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		defer userRows.Close()
		
		fmt.Printf("%-5s %-20s %-30s\n", "ID", "NAME", "EMAIL")
		fmt.Println("========================================================")
		
		for userRows.Next() {
			var userID int
			var name, email sql.NullString
			
			err := userRows.Scan(&userID, &name, &email)
			if err != nil {
				log.Printf("Error scanning user row: %v", err)
				continue
			}
			
			fmt.Printf("%-5d %-20s %-30s\n", 
				userID,
				getStringValue(name),
				getStringValue(email))
		}
	}

	// Test the CheckUserDevices logic for the actual user ID found
	fmt.Printf("\n🧪 Testing CheckUserDevices logic for user ID 1727068808:\n")
	testUserID := 1727068808
	testQuery := `SELECT id_device FROM device_setting_nodepath WHERE user_id = ? AND id_device IS NOT NULL AND id_device != ''`
	testRows, err := db.Query(testQuery, testUserID)
	if err != nil {
		log.Printf("Failed to test query: %v", err)
	} else {
		defer testRows.Close()
		
		var deviceIDs []string
		for testRows.Next() {
			var deviceID string
			if err := testRows.Scan(&deviceID); err != nil {
				log.Printf("Error scanning device ID: %v", err)
				continue
			}
			deviceIDs = append(deviceIDs, deviceID)
		}
		
		fmt.Printf("   User ID %d has %d devices: %v\n", testUserID, len(deviceIDs), deviceIDs)
	}

	// Also test the UUID to int conversion function
	fmt.Printf("\n🔄 Testing UUID to int conversion:\n")
	// Simulate the convertUUIDToInt function
	convertUUIDToInt := func(uuid string) int {
		h := fnv.New32a()
		h.Write([]byte(uuid))
		return int(h.Sum32())
	}
	
	// Test with some sample UUIDs to see what integer they convert to
	sampleUUIDs := []string{
		"e9057de2-ddae-452a-8b01-b68b9e6aec49", // The device ID we found
		"test-uuid-1",
		"test-uuid-2",
	}
	
	for _, uuid := range sampleUUIDs {
		convertedInt := convertUUIDToInt(uuid)
		fmt.Printf("   UUID '%s' converts to int: %d\n", uuid, convertedInt)
	}

	fmt.Println("\nCheck completed!")
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}