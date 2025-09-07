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

	// Use a simple test user ID that fits in INT range
	// Instead of converting UUID, let's use a simple integer ID
	convertedUserID := 12345
	
	fmt.Printf("🔄 Using Test User ID: %d\n", convertedUserID)

	// Update the device to link to this user
	updateQuery := `UPDATE device_setting_nodepath SET user_id = ? WHERE id_device = ?`
	result, err := db.Exec(updateQuery, convertedUserID, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatalf("Failed to update device user_id: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("✅ Successfully updated device FakhriAidilTLW-001 to user ID %d\n", convertedUserID)
	} else {
		fmt.Println("⚠️ No rows were updated - device might not exist")
	}

	// Verify the update
	fmt.Println("\n=== Verifying update ===")
	var currentUserID sql.NullInt32
	verifyQuery := `SELECT user_id FROM device_setting_nodepath WHERE id_device = ?`
	err = db.QueryRow(verifyQuery, "FakhriAidilTLW-001").Scan(&currentUserID)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}

	if currentUserID.Valid {
		fmt.Printf("✅ Device FakhriAidilTLW-001 is now linked to user ID: %d\n", currentUserID.Int32)
	} else {
		fmt.Println("❌ Device user_id is still NULL")
	}
}