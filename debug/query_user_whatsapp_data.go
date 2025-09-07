package main

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// convertMySQLURL converts mysql:// URL to DSN format
func convertMySQLURL(mysqlURL string) string {
	// Remove mysql:// prefix
	if strings.HasPrefix(mysqlURL, "mysql://") {
		mysqlURL = mysqlURL[8:]
	}
	
	// Split into parts: user:pass@host:port/database
	parts := strings.Split(mysqlURL, "/")
	if len(parts) != 2 {
		return mysqlURL // Return as-is if format is unexpected
	}
	
	database := parts[1]
	userHostPart := parts[0]
	
	// Split user:pass@host:port
	atIndex := strings.LastIndex(userHostPart, "@")
	if atIndex == -1 {
		return mysqlURL // Return as-is if no @ found
	}
	
	userPass := userHostPart[:atIndex]
	hostPort := userHostPart[atIndex+1:]
	
	// Format: user:pass@tcp(host:port)/database?parseTime=true&loc=UTC
	return fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&loc=UTC", userPass, hostPort, database)
}

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

	// Get the current logged-in user UUID (from session)
	userUUID := "eed7f0209432a1af30ec9f69ce097b79" // Current logged-in user
	convertedUserID := convertUUIDToInt(userUUID)

	fmt.Printf("\n🔍 Querying WhatsApp data for user: %s (ID: %d)\n", userUUID, convertedUserID)

	// Step 1: Get all id_device values for this user from device_setting_nodepath
	fmt.Println("\n📱 Step 1: Getting user's devices from device_setting_nodepath...")
	deviceQuery := `
		SELECT id_device, provider, api_key_option 
		FROM device_setting_nodepath 
		WHERE user_id = ?
	`
	deviceRows, err := db.Query(deviceQuery, convertedUserID)
	if err != nil {
		log.Fatalf("Failed to query user devices: %v", err)
	}
	defer deviceRows.Close()

	var userDevices []string
	for deviceRows.Next() {
		var idDevice, provider, apiKeyOption string
		err := deviceRows.Scan(&idDevice, &provider, &apiKeyOption)
		if err != nil {
			log.Printf("Error scanning device: %v", err)
			continue
		}
		userDevices = append(userDevices, idDevice)
		fmt.Printf("   Device: %s (Provider: %s, Model: %s)\n", idDevice, provider, apiKeyOption)
	}

	if len(userDevices) == 0 {
		fmt.Println("   ❌ No devices found for this user")
		return
	}

	fmt.Printf("   ✅ Found %d device(s) for user\n", len(userDevices))

	// Step 2: Query ai_whatsapp_nodepath for all conversations from user's devices
	fmt.Println("\n💬 Step 2: Getting WhatsApp conversations for user's devices...")
	
	// Create placeholders for IN clause
	placeholders := strings.Repeat("?,", len(userDevices))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	whatsappQuery := `
		SELECT id, id_device, prospect_num, stage, conv_current, conv_last, created_at
		FROM ai_whatsapp_nodepath 
		WHERE id_device IN (` + placeholders + `)
		ORDER BY created_at DESC
	`

	// Convert userDevices to []interface{} for query
	args := make([]interface{}, len(userDevices))
	for i, device := range userDevices {
		args[i] = device
	}

	whatsappRows, err := db.Query(whatsappQuery, args...)
	if err != nil {
		log.Fatalf("Failed to query WhatsApp conversations: %v", err)
	}
	defer whatsappRows.Close()

	conversationCount := 0
	for whatsappRows.Next() {
		var id int
		var idDevice, prospectNum, stage, convLast, convCurrent, botBalas, createdAt string
		var human, balas int
		
		err := whatsappRows.Scan(&id, &idDevice, &prospectNum, &stage, &convLast, 
								 &convCurrent, &human, &balas, &botBalas, &createdAt)
		if err != nil {
			log.Printf("Error scanning conversation: %v", err)
			continue
		}

		conversationCount++
		fmt.Printf("\n   Conversation #%d:\n", conversationCount)
		fmt.Printf("     ID: %d\n", id)
		fmt.Printf("     Device: %s\n", idDevice)
		fmt.Printf("     Phone: %s\n", prospectNum)
		fmt.Printf("     Stage: %s\n", stage)
		fmt.Printf("     Human Mode: %d (0=AI, 1=Human)\n", human)
		fmt.Printf("     Reply Count: %d\n", balas)
		fmt.Printf("     Created: %s\n", createdAt)
		if convLast != "" {
			fmt.Printf("     Last Conv: %s\n", truncateString(convLast, 50))
		}
		if convCurrent != "" {
			fmt.Printf("     Current Conv: %s\n", truncateString(convCurrent, 50))
		}
	}

	if conversationCount == 0 {
		fmt.Println("   ❌ No WhatsApp conversations found for user's devices")
	} else {
		fmt.Printf("\n   ✅ Found %d conversation(s) for user's devices\n", conversationCount)
	}

	// Step 3: Show summary statistics
	fmt.Println("\n📊 Step 3: Summary statistics...")
	
	// Count total conversations per device
	for _, device := range userDevices {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_device = ?`, device).Scan(&count)
		if err != nil {
			log.Printf("Error counting conversations for device %s: %v", device, err)
			continue
		}
		fmt.Printf("   Device %s: %d total conversations\n", device, count)
	}

	fmt.Println("\n🎉 Query completed! This demonstrates how to get all WhatsApp data for a logged-in user.")
	fmt.Println("\n💡 Query Pattern:")
	fmt.Println("   1. Convert user UUID to integer using convertUUIDToInt()")
	fmt.Println("   2. Get all id_device values for user from device_setting_nodepath")
	fmt.Println("   3. Use those id_device values to query ai_whatsapp_nodepath")
	fmt.Println("   4. This gives you all WhatsApp conversations for all user's devices")
}

// truncateString truncates a string to maxLength characters
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}