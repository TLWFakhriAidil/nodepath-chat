package main

import (
	"database/sql"
	"fmt"
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
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?parseTime=true"
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

	// Execute ID Staff to ID Device migration
	log.Println("\n🔄 Executing ID Staff to ID Device migration...")
	executeIDStaffMigration(db)



	// Execute JSON to TEXT column migration
	log.Println("\n🔄 Executing JSON to TEXT migration...")
	executeJSONToTextMigration(db)

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

// executeIDStaffMigration performs the migration from id_staff to id_device
func executeIDStaffMigration(db *sql.DB) {
	log.Println("Starting ID Staff to ID Device migration...")
	
	// Define tables to migrate
	tables := []string{
		"ai_whatsapp_nodepath",
		"conversation_log_nodepath", 
		// conversation_log_nodepath_backup removed - no longer needed
	}
	
	successCount := 0
	for _, tableName := range tables {
		log.Printf("\n=== Migrating table: %s ===", tableName)
		
		// Check if id_staff column exists
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
							 WHERE table_schema = DATABASE() 
							 AND table_name = ? 
							 AND column_name = 'id_staff'`, tableName).Scan(&count)
		if err != nil {
			log.Printf("❌ Error checking id_staff column in %s: %v", tableName, err)
			continue
		}
		
		if count == 0 {
			log.Printf("ℹ️  Column id_staff does not exist in table %s, skipping", tableName)
			successCount++
			continue
		}
		
		// Check if id_device column already exists
		err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
							 WHERE table_schema = DATABASE() 
							 AND table_name = ? 
							 AND column_name = 'id_device'`, tableName).Scan(&count)
		if err != nil {
			log.Printf("❌ Error checking id_device column in %s: %v", tableName, err)
			continue
		}
		
		if count > 0 {
			log.Printf("⚠️  Column id_device already exists in table %s, dropping id_staff", tableName)
			_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN id_staff", tableName))
			if err != nil {
				log.Printf("❌ Error dropping id_staff column in %s: %v", tableName, err)
				continue
			}
			log.Printf("✅ Dropped id_staff column from %s", tableName)
			successCount++
			continue
		}
		
		// Get column definition for id_staff
		var columnType, isNullable, columnDefault, extra sql.NullString
		err = db.QueryRow(`SELECT COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA 
							 FROM information_schema.columns 
							 WHERE table_schema = DATABASE() 
							 AND table_name = ? 
							 AND column_name = 'id_staff'`, tableName).Scan(&columnType, &isNullable, &columnDefault, &extra)
		if err != nil {
			log.Printf("❌ Error getting column definition for %s: %v", tableName, err)
			continue
		}
		
		// Build column definition
		definition := columnType.String
		if isNullable.String == "NO" {
			definition += " NOT NULL"
		}
		if columnDefault.Valid && columnDefault.String != "" {
			if columnDefault.String == "CURRENT_TIMESTAMP" {
				definition += " DEFAULT CURRENT_TIMESTAMP"
			} else {
				definition += fmt.Sprintf(" DEFAULT '%s'", columnDefault.String)
			}
		}
		if extra.Valid && extra.String != "" {
			definition += " " + extra.String
		}
		
		log.Printf("Current column definition: id_staff %s", definition)
		
		// Rename column from id_staff to id_device
		renameQuery := fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN id_staff id_device %s", tableName, definition)
		log.Printf("Executing: %s", renameQuery)
		
		_, err = db.Exec(renameQuery)
		if err != nil {
			log.Printf("❌ Error renaming column in %s: %v", tableName, err)
			continue
		}
		
		log.Printf("✅ Successfully migrated %s.id_staff to %s.id_device", tableName, tableName)
		successCount++
	}
	
	log.Printf("\n=== Migration Summary ===")
	log.Printf("Total tables: %d", len(tables))
	log.Printf("Successful migrations: %d", successCount)
	log.Printf("Failed migrations: %d", len(tables)-successCount)
	
	if successCount == len(tables) {
		log.Println("🎉 All ID Staff to ID Device migrations completed successfully!")
	} else {
		log.Println("⚠️  Some migrations failed. Please check the logs above.")
	}
}



// executeJSONToTextMigration converts JSON columns to TEXT in ai_whatsapp_nodepath table
func executeJSONToTextMigration(db *sql.DB) {
	log.Println("Starting JSON to TEXT column migration...")
	
	tableName := "ai_whatsapp_nodepath"
	
	// Check current JSON columns
	jsonColumns := checkJSONColumns(db, tableName)
	if len(jsonColumns) == 0 {
		log.Println("✅ No JSON columns found - migration not needed.")
		return
	}

	log.Printf("Found %d JSON columns to migrate: %v\n", len(jsonColumns), jsonColumns)

	// Migrate each JSON column to TEXT
	for _, column := range jsonColumns {
		log.Printf("🔄 Migrating column: %s\n", column)
		
		// Alter column from JSON to TEXT
		alterQuery := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s TEXT COLLATE utf8mb4_unicode_ci", tableName, column)
		log.Printf("Executing: %s\n", alterQuery)
		
		_, err := db.Exec(alterQuery)
		if err != nil {
			log.Printf("❌ Error migrating column %s: %v\n", column, err)
			continue
		}
		
		log.Printf("✅ Successfully migrated %s from JSON to TEXT\n", column)
	}

	// Verify migration
	log.Println("\n🔍 Verifying migration...")
	remainingJSONColumns := checkJSONColumns(db, tableName)
	if len(remainingJSONColumns) == 0 {
		log.Println("✅ All JSON columns successfully migrated to TEXT!")
	} else {
		log.Printf("⚠️  Still have %d JSON columns: %v\n", len(remainingJSONColumns), remainingJSONColumns)
	}

	log.Println("🎉 JSON to TEXT migration completed!")
}

// checkJSONColumns returns list of columns that are JSON type
func checkJSONColumns(db *sql.DB, tableName string) []string {
	query := `SELECT COLUMN_NAME 
			  FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND DATA_TYPE = 'json'`

	rows, err := db.Query(query, tableName)
	if err != nil {
		log.Printf("Error checking JSON columns: %v", err)
		return nil
	}
	defer rows.Close()

	var jsonColumns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Printf("Error scanning column name: %v", err)
			continue
		}
		jsonColumns = append(jsonColumns, columnName)
	}

	return jsonColumns
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}