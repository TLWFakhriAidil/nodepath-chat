package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// convertMySQLURLToDSN converts a MySQL URL to a DSN format
func convertMySQLURLToDSN(mysqlURL string) (string, error) {
	u, err := url.Parse(mysqlURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse MySQL URL: %w", err)
	}

	// Extract user info
	username := u.User.Username()
	password, _ := u.User.Password()

	// Extract host and port
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "3306" // Default MySQL port
	}

	// Extract database name
	dbName := strings.TrimPrefix(u.Path, "/")

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, dbName)
	return dsn, nil
}

// migrateChatbotExecutionsTable renames staff_id to id_device in chatbot_executions_nodepath table
func migrateChatbotExecutionsTable(db *sql.DB) error {
	fmt.Println("\n=== Migrating chatbot_executions_nodepath table ===")

	// Check if id_device column already exists
	checkQuery := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'chatbot_executions_nodepath' 
		AND COLUMN_NAME = 'id_device'
	`

	var idDeviceExists int
	err := db.QueryRow(checkQuery).Scan(&idDeviceExists)
	if err != nil {
		return fmt.Errorf("failed to check id_device column: %w", err)
	}

	if idDeviceExists > 0 {
		fmt.Println("✓ id_device column already exists")
		
		// Check if staff_id still exists
		checkStaffQuery := `
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'chatbot_executions_nodepath' 
			AND COLUMN_NAME = 'staff_id'
		`
		
		var staffIdExists int
		err := db.QueryRow(checkStaffQuery).Scan(&staffIdExists)
		if err != nil {
			return fmt.Errorf("failed to check staff_id column: %w", err)
		}
		
		if staffIdExists > 0 {
			fmt.Println("⚠ Both staff_id and id_device exist, copying data and dropping staff_id...")
			
			// Copy data from staff_id to id_device where id_device is NULL
			updateQuery := `UPDATE chatbot_executions_nodepath SET id_device = staff_id WHERE id_device IS NULL AND staff_id IS NOT NULL`
			result, err := db.Exec(updateQuery)
			if err != nil {
				return fmt.Errorf("failed to copy staff_id data to id_device: %w", err)
			}
			
			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("✓ Copied %d rows from staff_id to id_device\n", rowsAffected)
			
			// Drop staff_id column
			dropQuery := `ALTER TABLE chatbot_executions_nodepath DROP COLUMN staff_id`
			_, err = db.Exec(dropQuery)
			if err != nil {
				return fmt.Errorf("failed to drop staff_id column: %w", err)
			}
			
			fmt.Println("✓ Dropped staff_id column")
		} else {
			fmt.Println("✓ Migration already complete - only id_device column exists")
		}
		
		return nil
	}

	// Check if staff_id column exists
	checkStaffQuery := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'chatbot_executions_nodepath' 
		AND COLUMN_NAME = 'staff_id'
	`

	var staffIdExists int
	err = db.QueryRow(checkStaffQuery).Scan(&staffIdExists)
	if err != nil {
		return fmt.Errorf("failed to check staff_id column: %w", err)
	}

	if staffIdExists == 0 {
		fmt.Println("✓ staff_id column does not exist - no migration needed")
		return nil
	}

	fmt.Println("⚠ staff_id column found, renaming to id_device...")

	// Rename staff_id to id_device
	renameQuery := `ALTER TABLE chatbot_executions_nodepath CHANGE COLUMN staff_id id_device VARCHAR(255)`
	_, err = db.Exec(renameQuery)
	if err != nil {
		return fmt.Errorf("failed to rename staff_id to id_device: %w", err)
	}

	fmt.Println("✓ Successfully renamed staff_id to id_device")

	// Update the index name if it exists
	dropIndexQuery := `ALTER TABLE chatbot_executions_nodepath DROP INDEX IF EXISTS idx_staff_id`
	_, err = db.Exec(dropIndexQuery)
	if err != nil {
		// Index might not exist, continue
		fmt.Printf("⚠ Could not drop idx_staff_id index (might not exist): %v\n", err)
	} else {
		fmt.Println("✓ Dropped idx_staff_id index")
	}

	// Create new index for id_device
	createIndexQuery := `ALTER TABLE chatbot_executions_nodepath ADD INDEX idx_id_device (id_device)`
	_, err = db.Exec(createIndexQuery)
	if err != nil {
		// Index might already exist, continue
		fmt.Printf("⚠ Could not create idx_id_device index (might already exist): %v\n", err)
	} else {
		fmt.Println("✓ Created idx_id_device index")
	}

	return nil
}

func main() {
	// Get database URL from environment or use default
	mysqlURL := os.Getenv("MYSQL_URI")
	if mysqlURL == "" {
		mysqlURL = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert MySQL URL to DSN
	dsn, err := convertMySQLURLToDSN(mysqlURL)
	if err != nil {
		log.Fatalf("Failed to convert MySQL URL to DSN: %v", err)
	}

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Migrate chatbot_executions_nodepath table
	err = migrateChatbotExecutionsTable(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("\n=== Migration completed successfully! ===")
}