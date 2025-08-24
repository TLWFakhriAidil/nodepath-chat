package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// TableMigration represents a table migration operation
type TableMigration struct {
	TableName string
	OldColumn string
	NewColumn string
	ColumnType string
}

// executeSQL executes a SQL statement and handles errors
func executeSQL(db *sql.DB, query string, description string) error {
	fmt.Printf("Executing: %s\n", description)
	fmt.Printf("SQL: %s\n", query)
	
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error %s: %v", description, err)
	}
	
	fmt.Printf("✅ %s completed successfully\n\n", description)
	return nil
}

// checkColumnExists checks if a column exists in a table
func checkColumnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ?`
	
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// getColumnDefinition gets the full column definition
func getColumnDefinition(db *sql.DB, tableName, columnName string) (string, error) {
	query := `SELECT COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA 
			  FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ?`
	
	var columnType, isNullable, columnDefault, extra sql.NullString
	err := db.QueryRow(query, tableName, columnName).Scan(&columnType, &isNullable, &columnDefault, &extra)
	if err != nil {
		return "", err
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
	
	return definition, nil
}

// migrateTable performs the migration for a single table
func migrateTable(db *sql.DB, migration TableMigration) error {
	fmt.Printf("\n=== Migrating table: %s ===\n", migration.TableName)
	
	// Check if old column exists
	hasOldColumn, err := checkColumnExists(db, migration.TableName, migration.OldColumn)
	if err != nil {
		return fmt.Errorf("error checking old column: %v", err)
	}
	
	if !hasOldColumn {
		fmt.Printf("ℹ️  Column %s does not exist in table %s, skipping migration\n", migration.OldColumn, migration.TableName)
		return nil
	}
	
	// Check if new column already exists
	hasNewColumn, err := checkColumnExists(db, migration.TableName, migration.NewColumn)
	if err != nil {
		return fmt.Errorf("error checking new column: %v", err)
	}
	
	if hasNewColumn {
		fmt.Printf("⚠️  Column %s already exists in table %s\n", migration.NewColumn, migration.TableName)
		
		// Check if we should drop the old column
		fmt.Printf("Dropping old column %s...\n", migration.OldColumn)
		dropQuery := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", migration.TableName, migration.OldColumn)
		err = executeSQL(db, dropQuery, fmt.Sprintf("dropping old column %s from %s", migration.OldColumn, migration.TableName))
		if err != nil {
			return err
		}
		return nil
	}
	
	// Get the current column definition
	columnDef, err := getColumnDefinition(db, migration.TableName, migration.OldColumn)
	if err != nil {
		return fmt.Errorf("error getting column definition: %v", err)
	}
	
	fmt.Printf("Current column definition: %s %s\n", migration.OldColumn, columnDef)
	
	// Rename the column
	renameQuery := fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN %s %s %s", 
		migration.TableName, migration.OldColumn, migration.NewColumn, columnDef)
	
	err = executeSQL(db, renameQuery, fmt.Sprintf("renaming %s to %s in %s", migration.OldColumn, migration.NewColumn, migration.TableName))
	if err != nil {
		return err
	}
	
	fmt.Printf("✅ Successfully migrated %s.%s to %s.%s\n", migration.TableName, migration.OldColumn, migration.TableName, migration.NewColumn)
	return nil
}

// convertMySQLURL converts mysql:// URL to DSN format
func convertMySQLURL(mysqlURL string) string {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL
	}
	
	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURL, "mysql://")
	
	// Replace the first @ with @ and add tcp() wrapper for host:port
	parts := strings.Split(dsn, "@")
	if len(parts) == 2 {
		userPass := parts[0]
		hostDB := parts[1]
		
		// Split host and database
		hostParts := strings.Split(hostDB, "/")
		if len(hostParts) == 2 {
			host := hostParts[0]
			db := hostParts[1]
			return fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
		}
	}
	
	return dsn
}

func main() {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		// Fallback to default connection for local testing
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
		fmt.Println("Using default MySQL URI for testing")
	}
	
	fmt.Printf("Starting ID Staff to ID Device Migration...\n")
	fmt.Printf("Connecting to database...\n")
	
	// Convert MySQL URL to DSN format
	dsn := convertMySQLURL(mysqlURI)
	
	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	
	fmt.Println("✅ Database connection successful")
	
	// Define migrations
	migrations := []TableMigration{
		{
			TableName: "ai_whatsapp_nodepath",
			OldColumn: "id_staff",
			NewColumn: "id_device",
		},
		{
			TableName: "conversation_log_nodepath",
			OldColumn: "id_staff",
			NewColumn: "id_device",
		},
		{
			TableName: "conversation_log_nodepath_backup",
			OldColumn: "id_staff",
			NewColumn: "id_device",
		},
	}
	
	// Execute migrations
	successCount := 0
	for _, migration := range migrations {
		err := migrateTable(db, migration)
		if err != nil {
			fmt.Printf("❌ Migration failed for table %s: %v\n", migration.TableName, err)
		} else {
			successCount++
		}
	}
	
	fmt.Printf("\n=== Migration Summary ===\n")
	fmt.Printf("Total tables: %d\n", len(migrations))
	fmt.Printf("Successful migrations: %d\n", successCount)
	fmt.Printf("Failed migrations: %d\n", len(migrations)-successCount)
	
	if successCount == len(migrations) {
		fmt.Println("\n🎉 All migrations completed successfully!")
	} else {
		fmt.Println("\n⚠️  Some migrations failed. Please check the logs above.")
	}
}