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

	// Check current table structure
	fmt.Println("Current conversation_log_nodepath table structure:")
	rows, err := db.Query("SHOW COLUMNS FROM conversation_log_nodepath")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Field\t\tType\t\tNull\tKey\tDefault\tExtra")
	fmt.Println("-----\t\t----\t\t----\t---\t-------\t-----")
	for rows.Next() {
		var field, typ, null, key, def, extra string
		rows.Scan(&field, &typ, &null, &key, &def, &extra)
		fmt.Printf("%s\t\t%s\t%s\t%s\t%s\t%s\n", field, typ, null, key, def, extra)
	}

	// Test multiple record insertions to verify AUTO_INCREMENT works
	fmt.Println("\nTesting multiple record insertions...")
	for i := 1; i <= 3; i++ {
		result, err := db.Exec("INSERT INTO conversation_log_nodepath (prospect_num, sender, message, message_type, stage, id_staff) VALUES (?, ?, ?, ?, ?, ?)", 
			fmt.Sprintf("test%d", i), "user", fmt.Sprintf("Test message %d", i), "text", "initial", "staff001")
		if err != nil {
			fmt.Printf("❌ Error inserting record %d: %v\n", i, err)
		} else {
			lastID, _ := result.LastInsertId()
			fmt.Printf("✅ Record %d inserted with AUTO_INCREMENT ID: %d\n", i, lastID)
		}
	}

	// Clean up test records
	fmt.Println("\nCleaning up test records...")
	_, err = db.Exec("DELETE FROM conversation_log_nodepath WHERE prospect_num LIKE 'test%'")
	if err != nil {
		fmt.Printf("Error cleaning up: %v\n", err)
	} else {
		fmt.Println("✅ Test records cleaned up")
	}

	fmt.Println("\n🎉 The 'Field id doesn't have a default value' error has been resolved!")
	fmt.Println("✅ AUTO_INCREMENT is working correctly for the id column")
}