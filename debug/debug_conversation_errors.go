package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Function to convert mysql:// URL to Go MySQL driver format
func convertMySQLURL(mysqlURL string) string {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL
	}

	// Remove mysql:// prefix
	url := strings.TrimPrefix(mysqlURL, "mysql://")
	
	// Split by @ to separate credentials from host/db
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return mysqlURL
	}

	credentials := parts[0]
	hostAndDB := parts[1]

	// Split host and database
	hostParts := strings.Split(hostAndDB, "/")
	if len(hostParts) != 2 {
		return mysqlURL
	}

	host := hostParts[0]
	database := hostParts[1]

	// Format: user:pass@tcp(host:port)/database
	return fmt.Sprintf("%s@tcp(%s)/%s", credentials, host, database)
}

func main() {
	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert to Go MySQL driver format
	dsn := convertMySQLURL(mysqlURI)
	fmt.Printf("Connecting to database with DSN: %s\n", dsn)

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Database connection successful")

	// 1. Check conversation_log_nodepath table structure
	fmt.Println("\n=== CHECKING conversation_log_nodepath TABLE STRUCTURE ===")
	rows, err := db.Query("DESCRIBE conversation_log_nodepath")
	if err != nil {
		fmt.Printf("❌ Error describing table: %v\n", err)
	} else {
		fmt.Println("Table structure:")
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			if err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("  %s | %s | %s | %s | %s | %s\n", 
				field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
		}
		rows.Close()
	}

	// 2. Check if AUTO_INCREMENT is working
	fmt.Println("\n=== TESTING AUTO_INCREMENT FUNCTIONALITY ===")
	testQuery := `INSERT INTO conversation_log_nodepath 
				  (prospect_num, id_device, message, sender, stage, created_at) 
				  VALUES (?, ?, ?, ?, ?, NOW())`
	
	result, err := db.Exec(testQuery, "601234567890", "TEST-DEVICE", "Test message for AUTO_INCREMENT", "user", "test_stage")
	if err != nil {
		fmt.Printf("❌ Error inserting test record: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Test record inserted successfully with ID: %d\n", lastID)
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM conversation_log_nodepath WHERE id_device = ? AND prospect_num = ?", "TEST-DEVICE", "601234567890")
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not clean up test record: %v\n", err)
		} else {
			fmt.Println("✅ Test record cleaned up successfully")
		}
	}

	// 3. Check prospects_nodepath table for the specific prospect
	fmt.Println("\n=== CHECKING PROSPECTS TABLE FOR SPECIFIC PROSPECT ===")
	prospectNum := "601171219823"
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM prospects_nodepath WHERE prospect_num = ?", prospectNum).Scan(&count)
	if err != nil {
		fmt.Printf("❌ Error checking prospects table: %v\n", err)
	} else {
		fmt.Printf("Prospect %s found in prospects_nodepath: %d records\n", prospectNum, count)
	}

	// 4. Check conversation_nodepath table for the specific prospect
	fmt.Println("\n=== CHECKING CONVERSATION TABLE FOR SPECIFIC PROSPECT ===")
	err = db.QueryRow("SELECT COUNT(*) FROM conversation_nodepath WHERE prospect_num = ?", prospectNum).Scan(&count)
	if err != nil {
		fmt.Printf("❌ Error checking conversation table: %v\n", err)
	} else {
		fmt.Printf("Prospect %s found in conversation_nodepath: %d records\n", prospectNum, count)
	}

	// 5. Show recent conversation logs to understand the pattern
	fmt.Println("\n=== RECENT CONVERSATION LOGS ===")
	rows, err = db.Query(`SELECT id, id_device, prospect_num, stage, message_content, sender_type, timestamp 
						 FROM conversation_log_nodepath 
						 ORDER BY timestamp DESC LIMIT 5`)
	if err != nil {
		fmt.Printf("❌ Error fetching recent logs: %v\n", err)
	} else {
		fmt.Println("Recent conversation logs:")
		for rows.Next() {
			var id sql.NullInt64
			var idDevice, prospectNum, stage, messageContent, senderType, timestamp sql.NullString
			err := rows.Scan(&id, &idDevice, &prospectNum, &stage, &messageContent, &senderType, &timestamp)
			if err != nil {
				fmt.Printf("Error scanning log row: %v\n", err)
				continue
			}
			fmt.Printf("  ID: %v | Device: %s | Prospect: %s | Stage: %s | Type: %s | Time: %s\n", 
				id.Int64, idDevice.String, prospectNum.String, stage.String, senderType.String, timestamp.String)
		}
		rows.Close()
	}

	fmt.Println("\n=== DIAGNOSIS COMPLETE ===")
}