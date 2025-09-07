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

// convertUUIDToInt converts a UUID string to an integer using FNV-32a hash
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

	fmt.Println("✅ Connected to database successfully")

	// The logged-in user UUID and its converted integer
	userUUID := "eed7f0209432a1af30ec9f69ce097b79"
	correctUserID := convertUUIDToInt(userUUID)

	fmt.Printf("🔄 Linking device to correct user ID: %d\n", correctUserID)

	// Update the device to link to the correct user
	updateQuery := `UPDATE device_setting_nodepath SET user_id = ? WHERE id_device = 'FakhriAidilTLW-001'`
	result, err := db.Exec(updateQuery, correctUserID)
	if err != nil {
		log.Fatalf("Failed to update device linkage: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Warning: Could not get rows affected: %v", err)
	} else {
		fmt.Printf("✅ Updated %d row(s)\n", rowsAffected)
	}

	// Verify the update
	fmt.Println("\n🔍 Verifying the update:")
	verifyQuery := `SELECT user_id FROM device_setting_nodepath WHERE id_device = 'FakhriAidilTLW-001'`
	var updatedUserID sql.NullInt32
	err = db.QueryRow(verifyQuery).Scan(&updatedUserID)
	if err != nil {
		log.Printf("Failed to verify update: %v", err)
	} else {
		if updatedUserID.Valid {
			fmt.Printf("   Device FakhriAidilTLW-001 is now linked to user ID: %d\n", updatedUserID.Int32)
			if int(updatedUserID.Int32) == correctUserID {
				fmt.Println("   ✅ Device is correctly linked to the logged-in user!")
			} else {
				fmt.Printf("   ❌ Mismatch: Expected %d, got %d\n", correctUserID, updatedUserID.Int32)
			}
		} else {
			fmt.Println("   ❌ Device is not linked to any user")
		}
	}

	fmt.Println("\n🎉 Device linkage fix completed! The analytics endpoint should now work.")
}