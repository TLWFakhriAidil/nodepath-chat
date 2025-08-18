package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// fixProductionSchema connects to Railway production database and fixes the jam column issue
func main() {
	// Get MySQL URI from environment variable
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		// Fallback to Railway production database URI in proper format
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?parseTime=true"
		log.Println("Using default Railway production database URI")
	} else {
		// Convert mysql:// format to go-sql-driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			// Remove mysql:// prefix
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Convert to tcp format
			parts := strings.Split(mysqlURI, "@")
			if len(parts) == 2 {
				userPass := parts[0]
				hostDb := parts[1]
				hostDbParts := strings.Split(hostDb, "/")
				if len(hostDbParts) == 2 {
					host := hostDbParts[0]
					db := hostDbParts[1]
					mysqlURI = userPass + "@tcp(" + host + ")/" + db + "?parseTime=true"
				}
			}
		}
	}

	log.Printf("Connecting to production database...")

	// Connect to MySQL database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Successfully connected to production database")

	// Read the SQL fix script
	sqlContent, err := ioutil.ReadFile("production_fix_jam_column.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// Split SQL content into individual statements
	sqlStatements := strings.Split(string(sqlContent), ";")

	log.Println("🔧 Executing production schema fix...")

	// Execute each SQL statement
	for i, statement := range sqlStatements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}

		log.Printf("Executing statement %d: %s", i+1, statement[:min(50, len(statement))])

		// Handle SET statements and other non-result statements
		if strings.HasPrefix(strings.ToUpper(statement), "SET") ||
			strings.HasPrefix(strings.ToUpper(statement), "PREPARE") ||
			strings.HasPrefix(strings.ToUpper(statement), "EXECUTE") ||
			strings.HasPrefix(strings.ToUpper(statement), "DEALLOCATE") {
			_, err = db.Exec(statement)
			if err != nil {
				log.Printf("Warning: Statement failed: %v", err)
				continue
			}
		} else if strings.HasPrefix(strings.ToUpper(statement), "SELECT") {
			// Handle SELECT statements
			rows, err := db.Query(statement)
			if err != nil {
				log.Printf("Warning: Query failed: %v", err)
				continue
			}
			defer rows.Close()

			// Print query results
			columns, _ := rows.Columns()
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}
				rows.Scan(valuePtrs...)

				for i, col := range columns {
					val := values[i]
					if val != nil {
						log.Printf("%s: %v", col, val)
					}
				}
			}
		}
	}

	log.Println("🎉 Production schema fix completed successfully!")

	// Verify the fix by checking the table structure
	log.Println("🔍 Verifying ai_whatsapp_nodepath table structure...")
	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'ai_whatsapp_nodepath'
		ORDER BY ORDINAL_POSITION
	`)
	if err != nil {
		log.Printf("Failed to verify table structure: %v", err)
		return
	}
	defer rows.Close()

	log.Println("\n📋 Table Structure:")
	log.Println("COLUMN_NAME\t\tDATA_TYPE\tIS_NULLABLE\tCOLUMN_DEFAULT")
	log.Println("================================================================")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		defaultVal := "NULL"
		if columnDefault.Valid {
			defaultVal = columnDefault.String
		}

		log.Printf("%-20s\t%-15s\t%-10s\t%s", columnName, dataType, isNullable, defaultVal)
	}

	log.Println("\n✅ Production database schema verification completed!")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}