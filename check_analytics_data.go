package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	dbURL := os.Getenv("MYSQL_URI")
	if dbURL == "" {
		log.Fatal("MYSQL_URI environment variable not set")
	}

	// Connect to database
	db, err := sql.Open("mysql", dbURL+"?parseTime=true")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println("=" * 50)

	// Check device_setting_nodepath table
	fmt.Println("\n1. Checking device_setting_nodepath table:")
	fmt.Println("-" * 40)
	
	query := `SELECT id_device, provider, user_id, instance FROM device_setting_nodepath LIMIT 10`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query device_setting_nodepath:", err)
	}
	defer rows.Close()

	deviceCount := 0
	for rows.Next() {
		var idDevice, provider sql.NullString
		var userID sql.NullInt32
		var instance sql.NullString
		
		if err := rows.Scan(&idDevice, &provider, &userID, &instance); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		
		deviceCount++
		fmt.Printf("Device %d: id_device=%s, provider=%s, user_id=%v, instance=%s\n",
			deviceCount, 
			nullStringValue(idDevice),
			nullStringValue(provider),
			nullInt32Value(userID),
			nullStringValue(instance))
	}
	
	fmt.Printf("\nTotal devices shown: %d\n", deviceCount)

	// Check ai_whatsapp_nodepath table
	fmt.Println("\n2. Checking ai_whatsapp_nodepath table:")
	fmt.Println("-" * 40)
	
	query2 := `SELECT id_device, prospect_num, human, stage, date_order FROM ai_whatsapp_nodepath LIMIT 10`
	rows2, err := db.Query(query2)
	if err != nil {
		log.Fatal("Failed to query ai_whatsapp_nodepath:", err)
	}
	defer rows2.Close()

	convCount := 0
	for rows2.Next() {
		var idDevice, prospectNum, stage sql.NullString
		var human sql.NullInt32
		var dateOrder sql.NullTime
		
		if err := rows2.Scan(&idDevice, &prospectNum, &human, &stage, &dateOrder); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		
		convCount++
		fmt.Printf("Conv %d: id_device=%s, prospect=%s, human=%v, stage=%s, date=%s\n",
			convCount,
			nullStringValue(idDevice),
			nullStringValue(prospectNum),
			nullInt32Value(human),
			nullStringValue(stage),
			nullTimeValue(dateOrder))
	}
	
	fmt.Printf("\nTotal conversations shown: %d\n", convCount)

	// Check if there are any conversations linked to devices with user_id
	fmt.Println("\n3. Checking for linked data (conversations with user devices):")
	fmt.Println("-" * 40)
	
	query3 := `
		SELECT COUNT(*) as total
		FROM ai_whatsapp_nodepath a
		JOIN device_setting_nodepath d ON a.id_device = d.id_device
		WHERE d.user_id IS NOT NULL
	`
	
	var linkedCount int
	err = db.QueryRow(query3).Scan(&linkedCount)
	if err != nil {
		log.Println("Failed to count linked data:", err)
	} else {
		fmt.Printf("Total conversations linked to user devices: %d\n", linkedCount)
	}

	// Check for test user data
	fmt.Println("\n4. Checking for test user (user_id = 1):")
	fmt.Println("-" * 40)
	
	query4 := `
		SELECT COUNT(*) as device_count
		FROM device_setting_nodepath
		WHERE user_id = 1
	`
	
	var testUserDeviceCount int
	err = db.QueryRow(query4).Scan(&testUserDeviceCount)
	if err != nil {
		log.Println("Failed to count test user devices:", err)
	} else {
		fmt.Printf("Devices for user_id=1: %d\n", testUserDeviceCount)
	}

	// Get analytics data for current month
	fmt.Println("\n5. Testing analytics query (current month):")
	fmt.Println("-" * 40)
	
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	
	analyticsQuery := `
		SELECT 
			COUNT(*) as total_conversations,
			COUNT(CASE WHEN a.human = 0 THEN 1 END) as ai_active,
			COUNT(CASE WHEN a.human = 1 THEN 1 END) as human_takeover
		FROM ai_whatsapp_nodepath a
		JOIN device_setting_nodepath d ON a.id_device = d.id_device
		WHERE d.user_id = 1 AND a.date_order BETWEEN ? AND ?
	`
	
	var totalConv, aiActive, humanTakeover int
	err = db.QueryRow(analyticsQuery, startDate, endDate).Scan(&totalConv, &aiActive, &humanTakeover)
	if err != nil {
		log.Println("Failed to get analytics:", err)
	} else {
		fmt.Printf("Analytics for user_id=1 (current month):\n")
		fmt.Printf("  Total conversations: %d\n", totalConv)
		fmt.Printf("  AI Active: %d\n", aiActive)
		fmt.Printf("  Human Takeover: %d\n", humanTakeover)
	}
}

func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}

func nullInt32Value(ni sql.NullInt32) string {
	if ni.Valid {
		return fmt.Sprintf("%d", ni.Int32)
	}
	return "NULL"
}

func nullTimeValue(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02 15:04:05")
	}
	return "NULL"
}
