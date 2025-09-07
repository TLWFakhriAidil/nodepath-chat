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

// convertUUIDToInt converts a UUID string to a consistent integer using hash function
func convertUUIDToInt(uuid string) int {
	h := fnv.New32a()
	h.Write([]byte(uuid))
	return int(h.Sum32())
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
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

	fmt.Println("✅ Connected to database successfully")

	// Create a test user UUID for linking devices
	testUserUUID := "test-user-12345678-1234-1234-1234-123456789012"
	testUserID := convertUUIDToInt(testUserUUID)

	fmt.Printf("🔄 Test User UUID: %s\n", testUserUUID)
	fmt.Printf("🔄 Test User ID (converted): %d\n", testUserID)

	// Check existing devices without user_id
	fmt.Println("\n=== Checking devices without user_id ===")
	query := `SELECT id, id_device, provider, user_id FROM device_setting_nodepath WHERE user_id IS NULL OR user_id = 0`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	var devicesToUpdate []string
	for rows.Next() {
		var id, idDevice, provider string
		var userID sql.NullInt32
		err := rows.Scan(&id, &idDevice, &provider, &userID)
		if err != nil {
			log.Printf("Error scanning device: %v", err)
			continue
		}
		fmt.Printf("Device: %s (ID: %s, Provider: %s, UserID: %v)\n", idDevice, id, provider, userID)
		devicesToUpdate = append(devicesToUpdate, id)
	}

	if len(devicesToUpdate) == 0 {
		fmt.Println("✅ All devices already have user_id assigned")
		return
	}

	// Update devices to link them to the test user
	fmt.Printf("\n🔄 Updating %d devices to link to test user...\n", len(devicesToUpdate))
	updateQuery := `UPDATE device_setting_nodepath SET user_id = ? WHERE id = ?`

	for _, deviceID := range devicesToUpdate {
		_, err := db.Exec(updateQuery, testUserID, deviceID)
		if err != nil {
			log.Printf("Failed to update device %s: %v", deviceID, err)
			continue
		}
		fmt.Printf("✅ Updated device %s\n", deviceID)
	}

	// Verify the updates
	fmt.Println("\n=== Verifying updates ===")
	verifyQuery := `SELECT id_device, provider, user_id FROM device_setting_nodepath WHERE user_id = ?`
	verifyRows, err := db.Query(verifyQuery, testUserID)
	if err != nil {
		log.Printf("Failed to verify updates: %v", err)
		return
	}
	defer verifyRows.Close()

	fmt.Printf("Devices linked to user ID %d:\n", testUserID)
	for verifyRows.Next() {
		var idDevice, provider string
		var userID int
		err := verifyRows.Scan(&idDevice, &provider, &userID)
		if err != nil {
			log.Printf("Error scanning verified device: %v", err)
			continue
		}
		fmt.Printf("  - %s (%s) -> User ID: %d\n", idDevice, provider, userID)
	}

	fmt.Println("\n✅ Device linking completed!")
	fmt.Printf("\n📝 To test analytics, login with a user that has UUID: %s\n", testUserUUID)
	fmt.Printf("   This will convert to user ID: %d and should see the linked devices\n", testUserID)
}