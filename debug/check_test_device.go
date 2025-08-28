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
	// Use environment variable or fallback to remote database
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway"
	} else {
		// Convert mysql:// format to go-sql-driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Convert user:pass@host:port/db to user:pass@tcp(host:port)/db
			parts := strings.Split(mysqlURI, "/")
			if len(parts) == 2 {
				mysqlURI = parts[0] + "/" + parts[1]
				if strings.Contains(parts[0], "@") {
					hostPart := strings.Split(parts[0], "@")[1]
					userPart := strings.Split(parts[0], "@")[0]
					mysqlURI = userPart + "@tcp(" + hostPart + ")/" + parts[1]
				}
			}
		}
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Checking test device FakhriAidilTLW-001 ===")

	// Check if the test device exists
	query := `SELECT id_device, instance, provider, api_key_option FROM device_setting_nodepath WHERE id_device = ?`
	row := db.QueryRow(query, "FakhriAidilTLW-001")

	var idDevice, instance, provider, apiKeyOption string
	err = row.Scan(&idDevice, &instance, &provider, &apiKeyOption)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ Test device FakhriAidilTLW-001 not found in database")
			fmt.Println("\nCreating test device...")
			
			// Create the test device
			createQuery := `
				INSERT INTO device_setting_nodepath 
				(id, id_device, api_key_option, provider, instance, created_at, updated_at)
				VALUES (UUID(), ?, ?, ?, ?, NOW(), NOW())
			`
			
			_, err = db.Exec(createQuery, "FakhriAidilTLW-001", "openai/gpt-4.1", "wablas", "test-instance-001")
			if err != nil {
				log.Printf("Failed to create test device: %v", err)
				return
			}
			
			fmt.Println("✅ Test device created successfully!")
			fmt.Println("ID Device: FakhriAidilTLW-001")
			fmt.Println("Instance: test-instance-001")
			fmt.Println("Provider: wablas")
			fmt.Println("API Key Option: openai/gpt-4.1")
		} else {
			log.Printf("Error querying device: %v", err)
		}
		return
	}

	fmt.Println("✅ Test device found:")
	fmt.Printf("ID Device: %s\n", idDevice)
	fmt.Printf("Instance: %s\n", instance)
	fmt.Printf("Provider: %s\n", provider)
	fmt.Printf("API Key Option: %s\n", apiKeyOption)
	fmt.Printf("\nWebhook URL should be: /webhook/%s/%s\n", idDevice, instance)
}