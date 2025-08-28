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

func main() {
	// Get database URL from environment or use default
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Parse MySQL URI to Go driver format
	if mysqlURI[:8] == "mysql://" {
		// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
		mysqlURI = mysqlURI[8:] // Remove mysql:// prefix
		// Replace @ with @tcp( and add ) before /
		parts := strings.Split(mysqlURI, "/")
		if len(parts) >= 2 {
			connPart := parts[0]
			dbName := parts[1]
			// Split user:pass@host:port
			atIndex := strings.LastIndex(connPart, "@")
			if atIndex > 0 {
				userPass := connPart[:atIndex]
				hostPort := connPart[atIndex+1:]
				mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, dbName)
			}
		}
	}

	// Connect to database
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

	fmt.Println("Connected to database successfully")

	// Check if device_setting_nodepath table exists, create if not
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS device_setting_nodepath (
			id INT AUTO_INCREMENT PRIMARY KEY,
			id_device VARCHAR(255) UNIQUE NOT NULL,
			api_key TEXT,
			api_key_option VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Printf("Warning: Failed to create device_setting_nodepath table: %v", err)
	} else {
		fmt.Println("Device settings table created/verified successfully")
	}

	// Show existing table structure first
	showTableQuery := "SHOW TABLES LIKE 'chatbot_flows_nodepath'"
	rows, err := db.Query(showTableQuery)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			fmt.Println("Table chatbot_flows_nodepath already exists")
			// Show columns
			columnRows, err := db.Query("DESCRIBE chatbot_flows_nodepath")
			if err == nil {
				defer columnRows.Close()
				fmt.Println("Existing columns:")
				for columnRows.Next() {
					var field, fieldType, null, key, defaultVal, extra string
					columnRows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
					fmt.Printf("  %s %s\n", field, fieldType)
				}
			}
		} else {
			fmt.Println("Creating chatbot_flows_nodepath table...")
		}
	}

	// Create chatbot_flows_nodepath table if not exists
	createFlowTableQuery := `
		CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
			id INT AUTO_INCREMENT PRIMARY KEY,
			id_device VARCHAR(255) NOT NULL,
			stage VARCHAR(255) NOT NULL,
			niche VARCHAR(255),
			prompt TEXT,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_device_stage (id_device, stage),
			INDEX idx_device_default (id_device, is_default)
		)
	`

	_, err = db.Exec(createFlowTableQuery)
	if err != nil {
		log.Printf("Warning: Failed to create chatbot_flows_nodepath table: %v", err)
	} else {
		fmt.Println("Chatbot flows table created/verified successfully")
	}

	// Insert test device SCHQ-S94
	query := `
		INSERT INTO device_setting_nodepath (
			id_device, api_key, api_key_option, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			api_key = VALUES(api_key),
			api_key_option = VALUES(api_key_option),
			updated_at = VALUES(updated_at)
	`

	now := time.Now()
	apiKey := "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"
	apiKeyOption := "gpt-4"

	_, err = db.Exec(query, "SCHQ-S94", apiKey, apiKeyOption, now, now)
	if err != nil {
		log.Fatalf("Failed to insert device: %v", err)
	}

	fmt.Println("Device SCHQ-S94 created/updated successfully")

	// Insert test flow for SCHQ-S94
	flowQuery := `
		INSERT INTO chatbot_flows_nodepath (
			id, id_device, name, niche, nodes, edges, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			name = VALUES(name),
			niche = VALUES(niche),
			nodes = VALUES(nodes),
			edges = VALUES(edges),
			updated_at = VALUES(updated_at)
	`

	// Create basic nodes and edges JSON for welcome flow
	nodesJSON := `[{"id":"welcome","type":"start","data":{"label":"Welcome","prompt":"Welcome to our service! How can I help you today?"},"position":{"x":100,"y":100}}]`
	edgesJSON := `[]`

	_, err = db.Exec(flowQuery, "SCHQ-S94-FLOW", "SCHQ-S94", "Welcome Flow", "general", nodesJSON, edgesJSON, now, now)
	if err != nil {
		log.Printf("Warning: Failed to create test flow: %v", err)
	} else {
		fmt.Println("Test flow created/updated successfully")
	}

	fmt.Println("Setup completed successfully!")
}