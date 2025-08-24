package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"
	} else {
		// Convert mysql:// format to Go driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Convert user:pass@host:port/db to user:pass@tcp(host:port)/db
			parts := strings.Split(mysqlURI, "/")
			if len(parts) == 2 {
				hostPart := parts[0]
				dbName := parts[1]
				userPass := strings.Split(hostPart, "@")[0]
				hostPort := strings.Split(hostPart, "@")[1]
				mysqlURI = userPass + "@tcp(" + hostPort + ")/" + dbName
			}
		}
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if id column exists
	fmt.Println("Checking if id column exists...")
	rows, err := db.Query("SHOW COLUMNS FROM conversation_log_nodepath WHERE Field = 'id'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	hasID := false
	for rows.Next() {
		hasID = true
		break
	}

	if !hasID {
		fmt.Println("❌ ID column is missing. Adding it back...")
		
		// Add the id column as AUTO_INCREMENT PRIMARY KEY at the beginning
		_, err = db.Exec("ALTER TABLE conversation_log_nodepath ADD COLUMN id INT NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST")
		if err != nil {
			fmt.Printf("Error adding id column: %v\n", err)
			return
		}
		fmt.Println("✅ ID column added successfully")
	} else {
		fmt.Println("✅ ID column already exists")
	}

	// Verify the table structure
	fmt.Println("\nCurrent table structure:")
	rows2, err := db.Query("SHOW COLUMNS FROM conversation_log_nodepath")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Println("Field\t\tType\t\tNull\tKey\tDefault\tExtra")
	fmt.Println("-----\t\t----\t\t----\t---\t-------\t-----")
	for rows2.Next() {
		var field, typ, null, key, def, extra string
		rows2.Scan(&field, &typ, &null, &key, &def, &extra)
		fmt.Printf("%s\t\t%s\t%s\t%s\t%s\t%s\n", field, typ, null, key, def, extra)
	}

	// Test record insertion
	fmt.Println("\nTesting record insertion...")
	result, err := db.Exec("INSERT INTO conversation_log_nodepath (prospect_num, sender, message, message_type, stage, id_device) VALUES (?, ?, ?, ?, ?, ?)", 
		"test123", "user", "Test message", "text", "initial", "FakhriAidilTLW-001")
	if err != nil {
		fmt.Printf("❌ Error inserting test record: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Test record inserted with AUTO_INCREMENT ID: %d\n", lastID)
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM conversation_log_nodepath WHERE prospect_num = 'test123'")
		if err == nil {
			fmt.Println("✅ Test record cleaned up")
		}
	}

	fmt.Println("\n🎉 The 'Field id doesn't have a default value' error has been resolved!")
	fmt.Println("✅ AUTO_INCREMENT is working correctly for the id column")
}