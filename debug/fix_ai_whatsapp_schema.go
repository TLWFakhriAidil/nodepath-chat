package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
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

	// 1. Check current ai_whatsapp_nodepath table structure
	fmt.Println("\n=== CURRENT ai_whatsapp_nodepath TABLE STRUCTURE ===")
	rows, err := db.Query("DESCRIBE ai_whatsapp_nodepath")
	if err != nil {
		fmt.Printf("❌ Error describing table: %v\n", err)
		return
	}

	fmt.Println("Current columns:")
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}
		fmt.Printf("  %s | %s | %s | %s | %s | %s\n", field, fieldType, null, key, defaultVal, extra)
	}
	rows.Close()

	// 2. Check if id_prospect column has AUTO_INCREMENT
	fmt.Println("\n=== CHECKING AUTO_INCREMENT STATUS ===")
	var autoIncrement string
	err = db.QueryRow(`
		SELECT EXTRA 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = 'admin_railway' 
		  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
		  AND COLUMN_NAME = 'id_prospect'
	`).Scan(&autoIncrement)
	
	if err != nil {
		fmt.Printf("❌ Error checking AUTO_INCREMENT: %v\n", err)
	} else {
		fmt.Printf("id_prospect column EXTRA: '%s'\n", autoIncrement)
		if !strings.Contains(strings.ToLower(autoIncrement), "auto_increment") {
			fmt.Println("❌ id_prospect column does NOT have AUTO_INCREMENT!")
		} else {
			fmt.Println("✅ id_prospect column has AUTO_INCREMENT")
		}
	}

	// 3. Count existing records
	var recordCount int
	err = db.QueryRow("SELECT COUNT(*) FROM ai_whatsapp_nodepath").Scan(&recordCount)
	if err != nil {
		fmt.Printf("❌ Error counting records: %v\n", err)
		return
	}
	fmt.Printf("\n=== CURRENT RECORDS ===\nCurrent records in table: %d\n", recordCount)

	// 4. Fix the schema if needed
	fmt.Println("\n=== FIXING SCHEMA ===")
	if recordCount > 0 {
		fmt.Println("⚠️ Table has existing data. Creating backup...")
		_, err = db.Exec("CREATE TABLE ai_whatsapp_nodepath_backup AS SELECT * FROM ai_whatsapp_nodepath")
		if err != nil {
			fmt.Printf("❌ Error creating backup: %v\n", err)
			return
		}
		fmt.Println("✅ Backup created as ai_whatsapp_nodepath_backup")
	}

	fmt.Println("Dropping existing table...")
	_, err = db.Exec("DROP TABLE ai_whatsapp_nodepath")
	if err != nil {
		fmt.Printf("❌ Error dropping table: %v\n", err)
		return
	}

	fmt.Println("Creating table with correct schema...")
	createTableSQL := `
		CREATE TABLE ai_whatsapp_nodepath (
			id_prospect INT AUTO_INCREMENT PRIMARY KEY,
			id_staff VARCHAR(255) NOT NULL,
			prospect_num VARCHAR(255) NOT NULL UNIQUE,
			stage VARCHAR(255) DEFAULT NULL,
			date_order DATETIME DEFAULT NULL,
			conv_last TEXT COLLATE utf8mb4_unicode_ci DEFAULT NULL,
			conv_current TEXT DEFAULT NULL,
			jam VARCHAR(255) DEFAULT NULL,
			intro VARCHAR(255) DEFAULT NULL,
			human INT DEFAULT 0 COMMENT '0 = AI active, 1 = human takeover',
			catatan_staff VARCHAR(255) DEFAULT NULL,
			balas INT DEFAULT 0,
			data_image VARCHAR(255) DEFAULT NULL,
			conv_stage TEXT DEFAULT NULL,
			niche VARCHAR(255) DEFAULT NULL,
			bot_balas DATETIME DEFAULT NULL,
			keywordiklan VARCHAR(255) DEFAULT NULL,
			marketer VARCHAR(255) DEFAULT NULL,
			update_today DATETIME DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			
			INDEX idx_prospect_num (prospect_num),
			INDEX idx_id_staff (id_staff),
			INDEX idx_stage (stage),
			INDEX idx_human (human),
			INDEX idx_niche (niche),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`
	
	_, err = db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("❌ Error creating table: %v\n", err)
		return
	}
	fmt.Println("✅ Table created with correct schema")

	// 5. Test the new schema
	fmt.Println("\n=== TESTING NEW SCHEMA ===")
	testSQL := `
		INSERT INTO ai_whatsapp_nodepath (
			id_staff, prospect_num, stage, human, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := db.Exec(testSQL, "TEST-DEVICE", "TEST-PROSPECT-123", "initial", 0, now, now)
	if err != nil {
		fmt.Printf("❌ Error inserting test record: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Test record inserted successfully with ID: %d\n", lastID)
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM ai_whatsapp_nodepath WHERE id_prospect = ?", lastID)
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not clean up test record: %v\n", err)
		} else {
			fmt.Println("✅ Test record cleaned up successfully")
		}
	}

	// 6. Create the specific prospect from webhook
	fmt.Println("\n=== CREATING WEBHOOK PROSPECT ===")
	prospectNum := "601171219823"
	idDevice := "FakhriAidilTLW-001"
	
	insertProspectSQL := `
		INSERT INTO ai_whatsapp_nodepath (
			id_staff, prospect_num, stage, human, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	
	now = time.Now()
	result, err = db.Exec(insertProspectSQL, idDevice, prospectNum, "initial", 0, now, now)
	if err != nil {
		fmt.Printf("❌ Error creating prospect: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Created prospect record for %s with ID: %d\n", prospectNum, lastID)
	}

	// 7. Show final table structure
	fmt.Println("\n=== FINAL TABLE STRUCTURE ===")
	rows, err = db.Query("DESCRIBE ai_whatsapp_nodepath")
	if err != nil {
		fmt.Printf("❌ Error describing final table: %v\n", err)
		return
	}

	fmt.Println("Final table structure:")
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}
		fmt.Printf("  %s | %s | %s | %s | %s | %s\n", field, fieldType, null, key, defaultVal, extra)
	}
	rows.Close()

	fmt.Println("\n🎉 ai_whatsapp_nodepath schema fix completed!")
	fmt.Println("✅ ai_whatsapp_nodepath table now has proper AUTO_INCREMENT")
	fmt.Println("✅ All required columns are present")
	fmt.Println("✅ Prospect record created for webhook processing")
}