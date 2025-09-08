package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/**
 * Fix Analytics "No devices available" issue
 * Creates a test user that matches the existing device's user_id
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

	fmt.Println("\n=== FIXING ANALYTICS USER ISSUE ===")

	// Check current state
	checkCurrentState(db)

	// Create test user for the existing device
	createTestUser(db)

	// Verify fix
	verifyFix(db)
}

/**
 * Check current state of users and devices
 */
func checkCurrentState(db *sql.DB) {
	fmt.Println("\n1. Current state analysis:")

	// Check users
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Printf("❌ Error counting users: %v", err)
		return
	}
	fmt.Printf("   Users in database: %d\n", userCount)

	// Check devices
	var deviceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device IS NOT NULL AND id_device != ''").Scan(&deviceCount)
	if err != nil {
		log.Printf("❌ Error counting devices: %v", err)
		return
	}
	fmt.Printf("   Devices in database: %d\n", deviceCount)

	// Show device user_ids
	if deviceCount > 0 {
		fmt.Println("   Device user_ids:")
		rows, err := db.Query("SELECT DISTINCT user_id FROM device_setting_nodepath WHERE user_id IS NOT NULL")
		if err != nil {
			log.Printf("❌ Error fetching device user_ids: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var userID int
			err := rows.Scan(&userID)
			if err != nil {
				log.Printf("❌ Error scanning user_id: %v", err)
				continue
			}
			fmt.Printf("     - User ID: %d\n", userID)
		}
	}
}

/**
 * Create a test user that matches the existing device's user_id
 */
func createTestUser(db *sql.DB) {
	fmt.Println("\n2. Creating test user for existing device:")

	// Get the user_id from the existing device
	var existingUserID int
	err := db.QueryRow("SELECT user_id FROM device_setting_nodepath WHERE user_id IS NOT NULL LIMIT 1").Scan(&existingUserID)
	if err != nil {
		log.Printf("❌ Error getting existing user_id: %v", err)
		return
	}
	fmt.Printf("   Found device with user_id: %d\n", existingUserID)

	// Check if user already exists with this device's user_id
	var existingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", fmt.Sprintf("%d", existingUserID)).Scan(&existingCount)
	if err != nil {
		log.Printf("❌ Error checking existing user: %v", err)
		return
	}

	if existingCount > 0 {
		fmt.Printf("   ✅ User with ID %d already exists\n", existingUserID)
		return
	}

	// Generate UUID for user ID (users table uses CHAR(36) UUID)
	userUUID := uuid.New().String()

	// Create the user
	email := fmt.Sprintf("test-user-%d@nodepath.local", existingUserID)
	fullName := fmt.Sprintf("Test User %d", existingUserID)
	password := "test123456" // Test password

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Error hashing password: %v", err)
		return
	}

	// Insert the user (matching actual table structure)
	_, err = db.Exec(`
		INSERT INTO users (id, email, full_name, password_hash, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`, userUUID, email, fullName, string(hashedPassword), time.Now(), time.Now())
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
		return
	}

	// Now update device_setting_nodepath to use the new user UUID
	_, err = db.Exec(`
		UPDATE device_setting_nodepath 
		SET user_id = ? 
		WHERE user_id = ?
	`, userUUID, fmt.Sprintf("%d", existingUserID))
	if err != nil {
		log.Printf("❌ Error updating device user_id: %v", err)
		return
	}

	fmt.Printf("   ✅ Created test user:\n")
	fmt.Printf("      - ID: %s\n", userUUID)
	fmt.Printf("      - Email: %s\n", email)
	fmt.Printf("      - Full Name: %s\n", fullName)
	fmt.Printf("      - Password: %s\n", password)
	fmt.Printf("      - Updated device user_id from %d to %s\n", existingUserID, userUUID)
}

/**
 * Verify that the fix works
 */
func verifyFix(db *sql.DB) {
	fmt.Println("\n3. Verifying fix:")

	// Test the same query that CheckUserDevices uses
	var testUserID int
	err := db.QueryRow("SELECT user_id FROM device_setting_nodepath WHERE user_id IS NOT NULL LIMIT 1").Scan(&testUserID)
	if err != nil {
		log.Printf("❌ Error getting test user_id: %v", err)
		return
	}

	// Run the CheckUserDevices query
	rows, err := db.Query(`
		SELECT id_device FROM device_setting_nodepath 
		WHERE user_id = ? AND id_device IS NOT NULL AND id_device != ''
	`, testUserID)
	if err != nil {
		log.Printf("❌ Error running CheckUserDevices query: %v", err)
		return
	}
	defer rows.Close()

	var deviceIDs []string
	var count int
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			log.Printf("❌ Error scanning device ID: %v", err)
			continue
		}
		deviceIDs = append(deviceIDs, deviceID)
		count++
	}

	fmt.Printf("   User ID %d has %d devices: %v\n", testUserID, count, deviceIDs)

	if count > 0 {
		fmt.Println("   ✅ Fix successful! Analytics should now show devices.")
		fmt.Println("   📝 You can now log in with:")
		fmt.Printf("      - Email: test-user-%d@nodepath.local\n", testUserID)
		fmt.Println("      - Password: test123456")
	} else {
		fmt.Println("   ❌ Fix failed - no devices found for user")
	}
}