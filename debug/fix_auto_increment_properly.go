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

	fmt.Println("=== FIXING AUTO_INCREMENT FOR conversation_log_nodepath ===\n")

	// Check current table structure
	fmt.Println("Current table structure:")
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

	// Check if id column has AUTO_INCREMENT
	var hasAutoIncrement bool
	var columnName string
	err = db.QueryRow("SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'conversation_log_nodepath' AND COLUMN_NAME = 'id' AND EXTRA LIKE '%auto_increment%'").Scan(&columnName)
	if err != nil {
		if err == sql.ErrNoRows {
			hasAutoIncrement = false
		} else {
			log.Fatal(err)
		}
	} else {
		hasAutoIncrement = true
	}

	if hasAutoIncrement {
		fmt.Println("\n✅ ID column already has AUTO_INCREMENT")
	} else {
		fmt.Println("\n❌ ID column does NOT have AUTO_INCREMENT. Fixing...")
		
		// Modify the id column to add AUTO_INCREMENT
		_, err = db.Exec("ALTER TABLE conversation_log_nodepath MODIFY COLUMN id INT AUTO_INCREMENT")
		if err != nil {
			fmt.Printf("❌ Error adding AUTO_INCREMENT: %v\n", err)
			return
		}
		
		fmt.Println("✅ AUTO_INCREMENT added to id column")
	}

	// Verify the fix
	fmt.Println("\nVerifying table structure after fix:")
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

	// Test record insertion without specifying id
	fmt.Println("\nTesting record insertion without specifying id...")
	result, err := db.Exec("INSERT INTO conversation_log_nodepath (prospect_num, sender, message, stage, id_device) VALUES (?, ?, ?, ?, ?)",
		"test123", "user", "Test message", "initial", "FakhriAidilTLW-001")
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

	fmt.Println("\n🎉 AUTO_INCREMENT fix completed!")
	fmt.Println("✅ The 'Field id doesn't have a default value' error should now be resolved")
}