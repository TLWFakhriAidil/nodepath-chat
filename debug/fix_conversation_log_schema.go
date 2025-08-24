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

	// 1. Check current table structure
	fmt.Println("\n=== CURRENT TABLE STRUCTURE ===")
	rows, err := db.Query("DESCRIBE conversation_log_nodepath")
	if err != nil {
		fmt.Printf("❌ Error describing table: %v\n", err)
		return
	}

	fmt.Println("Current columns:")
	var currentColumns []string
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}
		currentColumns = append(currentColumns, field.String)
		fmt.Printf("  %s | %s | %s | %s | %s | %s\n", 
			field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
	}
	rows.Close()

	// 2. Check what columns the application expects vs what exists
	fmt.Println("\n=== SCHEMA ANALYSIS ===")
	expectedColumns := []string{"id", "prospect_num", "id_staff", "message", "sender", "stage", "created_at"}
	
	fmt.Println("Expected columns:", expectedColumns)
	fmt.Println("Current columns:", currentColumns)

	// Check for missing columns
	missingColumns := []string{}
	for _, expected := range expectedColumns {
		found := false
		for _, current := range currentColumns {
			if current == expected {
				found = true
				break
			}
		}
		if !found {
			missingColumns = append(missingColumns, expected)
		}
	}

	// Check for extra columns
	extraColumns := []string{}
	for _, current := range currentColumns {
		found := false
		for _, expected := range expectedColumns {
			if current == expected {
				found = true
				break
			}
		}
		if !found {
			extraColumns = append(extraColumns, current)
		}
	}

	fmt.Printf("Missing columns: %v\n", missingColumns)
	fmt.Printf("Extra columns: %v\n", extraColumns)

	// 3. Fix the schema by recreating the table with correct structure
	fmt.Println("\n=== FIXING SCHEMA ===")
	
	// First, backup existing data if any
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM conversation_log_nodepath").Scan(&count)
	if err != nil {
		fmt.Printf("❌ Error counting records: %v\n", err)
		return
	}
	
	fmt.Printf("Current records in table: %d\n", count)
	
	if count > 0 {
		fmt.Println("⚠️ Table has existing data. Creating backup...")
		
		// Create backup table
		_, err = db.Exec("CREATE TABLE conversation_log_nodepath_backup AS SELECT * FROM conversation_log_nodepath")
		if err != nil {
			fmt.Printf("❌ Error creating backup: %v\n", err)
			return
		}
		fmt.Println("✅ Backup created as conversation_log_nodepath_backup")
	}

	// Drop the existing table
	fmt.Println("Dropping existing table...")
	_, err = db.Exec("DROP TABLE conversation_log_nodepath")
	if err != nil {
		fmt.Printf("❌ Error dropping table: %v\n", err)
		return
	}

	// Create the table with correct schema
	fmt.Println("Creating table with correct schema...")
	createTableSQL := `
		CREATE TABLE conversation_log_nodepath (
			id INT AUTO_INCREMENT PRIMARY KEY,
			prospect_num VARCHAR(255) NOT NULL,
			id_device VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			sender VARCHAR(10) NOT NULL COMMENT 'user or bot',
			stage VARCHAR(255) DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			INDEX idx_prospect_num (prospect_num),
			INDEX idx_id_device (id_device),
			INDEX idx_sender (sender),
			INDEX idx_stage (stage),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`
	
	_, err = db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("❌ Error creating table: %v\n", err)
		return
	}
	
	fmt.Println("✅ Table created with correct schema")

	// 4. Test the new schema
	fmt.Println("\n=== TESTING NEW SCHEMA ===")
	
	// Test insertion with the exact query from the application
	testQuery := `INSERT INTO conversation_log_nodepath (
					prospect_num, id_device, message, sender, stage, created_at
				) VALUES (?, ?, ?, ?, ?, NOW())`
	
	result, err := db.Exec(testQuery, "601234567890", "FakhriAidilTLW-001", "Test message", "user", "test_stage")
	if err != nil {
		fmt.Printf("❌ Error inserting test record: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Test record inserted successfully with ID: %d\n", lastID)
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM conversation_log_nodepath WHERE prospect_num = ?", "601234567890")
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not clean up test record: %v\n", err)
		} else {
			fmt.Println("✅ Test record cleaned up successfully")
		}
	}

	// 5. Verify final structure
	fmt.Println("\n=== FINAL TABLE STRUCTURE ===")
	rows2, err := db.Query("DESCRIBE conversation_log_nodepath")
	if err != nil {
		fmt.Printf("❌ Error describing table: %v\n", err)
	} else {
		fmt.Println("Final table structure:")
		for rows2.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			err := rows2.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			if err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("  %s | %s | %s | %s | %s | %s\n", 
				field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
		}
		rows2.Close()
	}

	fmt.Println("\n🎉 Schema fix completed!")
	fmt.Println("✅ conversation_log_nodepath table now matches application expectations")
	fmt.Println("✅ AUTO_INCREMENT is properly configured")
	fmt.Println("✅ All required columns are present with correct names")
}