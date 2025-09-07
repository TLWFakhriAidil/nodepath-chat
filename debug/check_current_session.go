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

	// Check current active sessions
	fmt.Println("\n🔍 Checking active user sessions:")
	sessionQuery := `
		SELECT us.user_id, u.email, us.token, us.expires_at 
		FROM user_sessions us 
		JOIN users u ON us.user_id = u.id 
		WHERE us.expires_at > NOW() 
		ORDER BY us.expires_at DESC
	`
	rows, err := db.Query(sessionQuery)
	if err != nil {
		log.Printf("Failed to query sessions: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID, email, token, expiresAt string
		err := rows.Scan(&userID, &email, &token, &expiresAt)
		if err != nil {
			log.Printf("Error scanning session: %v", err)
			continue
		}

		// Convert UUID to integer like the auth middleware does
		convertedUserID := convertUUIDToInt(userID)

		fmt.Printf("   User: %s (UUID: %s)\n", email, userID)
		fmt.Printf("   Converted to integer: %d\n", convertedUserID)
		fmt.Printf("   Session token: %s...\n", token[:20])
		fmt.Printf("   Expires: %s\n\n", expiresAt)
	}

	// Check what device FakhriAidilTLW-001 is currently linked to
	fmt.Println("🔍 Current device linkage:")
	deviceQuery := `SELECT user_id FROM device_setting_nodepath WHERE id_device = 'FakhriAidilTLW-001'`
	var currentUserID sql.NullInt32
	err = db.QueryRow(deviceQuery).Scan(&currentUserID)
	if err != nil {
		log.Printf("Failed to check device linkage: %v", err)
	} else {
		if currentUserID.Valid {
			fmt.Printf("   Device FakhriAidilTLW-001 is linked to user ID: %d\n", currentUserID.Int32)
		} else {
			fmt.Printf("   Device FakhriAidilTLW-001 is not linked to any user\n")
		}
	}

	fmt.Println("\n💡 To fix the authentication issue, we need to link the device to the converted integer ID of the logged-in user.")
}