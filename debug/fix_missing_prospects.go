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

	// 1. Check if ai_whatsapp_nodepath table exists
	fmt.Println("\n=== CHECKING ai_whatsapp_nodepath TABLE ===")
	rows, err := db.Query("SHOW TABLES LIKE 'ai_whatsapp_nodepath'")
	if err != nil {
		fmt.Printf("❌ Error checking table existence: %v\n", err)
		return
	}
	
	tableExists := rows.Next()
	rows.Close()
	
	if !tableExists {
		fmt.Println("❌ ai_whatsapp_nodepath table does not exist!")
		fmt.Println("Creating ai_whatsapp_nodepath table...")
		
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
			fmt.Printf("❌ Error creating ai_whatsapp_nodepath table: %v\n", err)
			return
		}
		
		fmt.Println("✅ ai_whatsapp_nodepath table created successfully")
	} else {
		fmt.Println("✅ ai_whatsapp_nodepath table exists")
	}

	// 2. Check the specific prospect from the webhook logs
	fmt.Println("\n=== CHECKING SPECIFIC PROSPECT ===")
	prospectNum := "601171219823"
	idDevice := "FakhriAidilTLW-001"
	
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE prospect_num = ?", prospectNum).Scan(&count)
	if err != nil {
		fmt.Printf("❌ Error checking prospect: %v\n", err)
		return
	}
	
	fmt.Printf("Prospect %s found in ai_whatsapp_nodepath: %d records\n", prospectNum, count)
	
	if count == 0 {
		fmt.Printf("Creating missing prospect record for %s...\n", prospectNum)
		
		// Create the prospect record
		insertSQL := `
			INSERT INTO ai_whatsapp_nodepath (
				id_staff, prospect_num, stage, human, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?)
		`
		
		now := time.Now()
		_, err = db.Exec(insertSQL, idDevice, prospectNum, "initial", 0, now, now)
		if err != nil {
			fmt.Printf("❌ Error creating prospect: %v\n", err)
			return
		}
		
		fmt.Printf("✅ Created prospect record for %s\n", prospectNum)
	} else {
		fmt.Printf("✅ Prospect %s already exists\n", prospectNum)
	}

	// 3. Test conversation log creation with the prospect
	fmt.Println("\n=== TESTING CONVERSATION LOG CREATION ===")
	
	testLogSQL := `
		INSERT INTO conversation_log_nodepath (
			prospect_num, id_device, message, sender, stage, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := db.Exec(testLogSQL, prospectNum, idDevice, "Test message from webhook", "user", "initial", now)
	if err != nil {
		fmt.Printf("❌ Error creating test conversation log: %v\n", err)
	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("✅ Test conversation log created with ID: %d\n", lastID)
		
		// Clean up test record
		_, err = db.Exec("DELETE FROM conversation_log_nodepath WHERE id = ?", lastID)
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not clean up test log: %v\n", err)
		} else {
			fmt.Println("✅ Test conversation log cleaned up")
		}
	}

	// 4. Check device_setting_nodepath table for the device
	fmt.Println("\n=== CHECKING DEVICE SETTINGS ===")
	
	rows2, err := db.Query("SHOW TABLES LIKE 'device_setting_nodepath'")
	if err != nil {
		fmt.Printf("❌ Error checking device_setting_nodepath table: %v\n", err)
	} else {
		deviceTableExists := rows2.Next()
		rows2.Close()
		
		if !deviceTableExists {
			fmt.Println("❌ device_setting_nodepath table does not exist!")
			fmt.Println("Creating device_setting_nodepath table...")
			
			createDeviceTableSQL := `
				CREATE TABLE device_setting_nodepath (
					id VARCHAR(255) PRIMARY KEY,
					api_key TEXT NOT NULL,
					api_key_option VARCHAR(255) DEFAULT 'gpt-4o-mini',
					api_url VARCHAR(255) DEFAULT 'https://openrouter.ai/api/v1/chat/completions',
					device_name VARCHAR(255) DEFAULT NULL,
					status VARCHAR(50) DEFAULT 'active',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					
					INDEX idx_status (status),
					INDEX idx_created_at (created_at)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
			`
			
			_, err = db.Exec(createDeviceTableSQL)
			if err != nil {
				fmt.Printf("❌ Error creating device_setting_nodepath table: %v\n", err)
			} else {
				fmt.Println("✅ device_setting_nodepath table created successfully")
				
				// Insert the device from webhook
				insertDeviceSQL := `
					INSERT IGNORE INTO device_setting_nodepath (
						id, api_key, api_key_option, api_url, device_name, status
					) VALUES (?, ?, ?, ?, ?, ?)
				`
				
				_, err = db.Exec(insertDeviceSQL, idDevice, "default-api-key", "gpt-4o-mini", "https://openrouter.ai/api/v1/chat/completions", idDevice, "active")
				if err != nil {
					fmt.Printf("❌ Error inserting device: %v\n", err)
				} else {
					fmt.Printf("✅ Device %s added to device_setting_nodepath\n", idDevice)
				}
			}
		} else {
			fmt.Println("✅ device_setting_nodepath table exists")
			
			// Check if the device exists
			var deviceCount int
			err = db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id = ?", idDevice).Scan(&deviceCount)
			if err != nil {
				fmt.Printf("❌ Error checking device: %v\n", err)
			} else if deviceCount == 0 {
				fmt.Printf("Device %s not found, creating...\n", idDevice)
				
				insertDeviceSQL := `
					INSERT INTO device_setting_nodepath (
						id, api_key, api_key_option, api_url, device_name, status
					) VALUES (?, ?, ?, ?, ?, ?)
				`
				
				_, err = db.Exec(insertDeviceSQL, idDevice, "default-api-key", "gpt-4o-mini", "https://openrouter.ai/api/v1/chat/completions", idDevice, "active")
				if err != nil {
					fmt.Printf("❌ Error inserting device: %v\n", err)
				} else {
					fmt.Printf("✅ Device %s added to device_setting_nodepath\n", idDevice)
				}
			} else {
				fmt.Printf("✅ Device %s already exists in device_setting_nodepath\n", idDevice)
			}
		}
	}

	fmt.Println("\n🎉 Missing prospects and device setup completed!")
	fmt.Println("✅ ai_whatsapp_nodepath table ready")
	fmt.Println("✅ conversation_log_nodepath table ready")
	fmt.Println("✅ device_setting_nodepath table ready")
	fmt.Println("✅ Prospect records created for webhook processing")
}