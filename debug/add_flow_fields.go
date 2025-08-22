package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is required")
	}

	// Convert mysql:// URL to DSN format
	dsn := convertMySQLURI(mysqlURI)

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

	log.Println("🔧 Adding flow execution fields to ai_whatsapp_nodepath table...")

	// Check if flow_reference column already exists
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
					   WHERE table_schema = DATABASE() 
					   AND table_name = 'ai_whatsapp_nodepath' 
					   AND column_name = 'flow_reference'`).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check flow_reference column: %v", err)
	}

	if count > 0 {
		log.Println("✅ flow_reference column already exists")
	} else {
		log.Println("➕ Adding flow_reference column...")
		_, err = db.Exec(`ALTER TABLE ai_whatsapp_nodepath 
						 ADD COLUMN flow_reference VARCHAR(255) DEFAULT NULL COMMENT 'Reference to chatbot flow being executed'`)
		if err != nil {
			log.Printf("❌ Failed to add flow_reference: %v", err)
		} else {
			log.Println("✅ Added flow_reference column")
		}
	}

	// Add other flow execution fields
	flowFields := map[string]string{
		"current_node":     "VARCHAR(255) DEFAULT NULL COMMENT 'Current node in the flow execution'",
		"variables":        "JSON DEFAULT NULL COMMENT 'Flow execution variables'",
		"execution_status": "ENUM('active', 'completed', 'failed') DEFAULT NULL COMMENT 'Flow execution status'",
		"execution_id":     "VARCHAR(255) DEFAULT NULL COMMENT 'Unique execution identifier'",
	}

	for fieldName, fieldDef := range flowFields {
		err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
						   WHERE table_schema = DATABASE() 
						   AND table_name = 'ai_whatsapp_nodepath' 
						   AND column_name = ?`, fieldName).Scan(&count)
		if err != nil {
			log.Printf("❌ Failed to check %s column: %v", fieldName, err)
			continue
		}

		if count > 0 {
			log.Printf("✅ %s column already exists", fieldName)
		} else {
			log.Printf("➕ Adding %s column...", fieldName)
			_, err = db.Exec(fmt.Sprintf(`ALTER TABLE ai_whatsapp_nodepath ADD COLUMN %s %s`, fieldName, fieldDef))
			if err != nil {
				log.Printf("❌ Failed to add %s: %v", fieldName, err)
			} else {
				log.Printf("✅ Added %s column", fieldName)
			}
		}
	}

	// Check if id_device column exists (should be renamed from id_staff)
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
					   WHERE table_schema = DATABASE() 
					   AND table_name = 'ai_whatsapp_nodepath' 
					   AND column_name = 'id_device'`).Scan(&count)
	if err != nil {
		log.Printf("❌ Failed to check id_device column: %v", err)
	} else if count > 0 {
		log.Println("✅ id_device column already exists")
	} else {
		// Check if id_staff exists to rename it
		err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
						   WHERE table_schema = DATABASE() 
						   AND table_name = 'ai_whatsapp_nodepath' 
						   AND column_name = 'id_staff'`).Scan(&count)
		if err != nil {
			log.Printf("❌ Failed to check id_staff column: %v", err)
		} else if count > 0 {
			log.Println("🔄 Renaming id_staff to id_device...")
			_, err = db.Exec(`ALTER TABLE ai_whatsapp_nodepath CHANGE COLUMN id_staff id_device VARCHAR(255) NOT NULL`)
			if err != nil {
				log.Printf("❌ Failed to rename id_staff to id_device: %v", err)
			} else {
				log.Println("✅ Renamed id_staff to id_device")
			}
		} else {
			log.Println("❌ Neither id_staff nor id_device column found")
		}
	}

	// Add indexes for better performance
	indexes := map[string]string{
		"idx_ai_whatsapp_flow_reference":   "flow_reference",
		"idx_ai_whatsapp_current_node":     "current_node",
		"idx_ai_whatsapp_execution_status": "execution_status",
		"idx_ai_whatsapp_execution_id":     "execution_id",
	}

	for indexName, columnName := range indexes {
		// Check if index already exists
		err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics 
						   WHERE table_schema = DATABASE() 
						   AND table_name = 'ai_whatsapp_nodepath' 
						   AND index_name = ?`, indexName).Scan(&count)
		if err != nil {
			log.Printf("❌ Failed to check index %s: %v", indexName, err)
			continue
		}

		if count > 0 {
			log.Printf("✅ Index %s already exists", indexName)
		} else {
			log.Printf("➕ Creating index %s...", indexName)
			_, err = db.Exec(fmt.Sprintf(`CREATE INDEX %s ON ai_whatsapp_nodepath(%s)`, indexName, columnName))
			if err != nil {
				log.Printf("❌ Failed to create index %s: %v", indexName, err)
			} else {
				log.Printf("✅ Created index %s", indexName)
			}
		}
	}

	log.Println("🎉 Flow execution fields migration completed!")
}

// convertMySQLURI converts mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
func convertMySQLURI(mysqlURI string) string {
	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURI, "mysql://")
	
	// Split into parts
	parts := strings.Split(dsn, "/")
	if len(parts) != 2 {
		return dsn
	}
	
	connectionPart := parts[0]
	database := parts[1]
	
	// Split connection part
	userPassHost := strings.Split(connectionPart, "@")
	if len(userPassHost) != 2 {
		return dsn
	}
	
	userPass := userPassHost[0]
	hostPort := userPassHost[1]
	
	// Format as Go MySQL driver expects
	return fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userPass, hostPort, database)
}