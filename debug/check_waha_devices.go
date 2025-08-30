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
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Convert mysql:// to proper DSN format
	dsn := mysqlURI
	if len(dsn) > 8 && dsn[:8] == "mysql://" {
		// mysql://user:pass@host:port/db -> user:pass@tcp(host:port)/db
		dsn = dsn[8:] // Remove mysql:// prefix
		// Find the @ symbol to separate user:pass from host:port/db
		if atIndex := strings.Index(dsn, "@"); atIndex != -1 {
			userPass := dsn[:atIndex]
			hostPortDb := dsn[atIndex+1:]
			// Find the / to separate host:port from db
			if slashIndex := strings.Index(hostPortDb, "/"); slashIndex != -1 {
				hostPort := hostPortDb[:slashIndex]
				dbName := hostPortDb[slashIndex+1:]
				dsn = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, dbName)
			}
		}
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

	fmt.Println("✅ Connected to database successfully")

	// Check all devices in device_setting_nodepath
	fmt.Println("\n🔍 Checking all devices in device_setting_nodepath table...")
	query := `SELECT id_device, provider, phone_number, instance FROM device_setting_nodepath ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-12s %-15s %-30s\n", "ID_DEVICE", "PROVIDER", "PHONE_NUMBER", "INSTANCE")
	fmt.Println("=================================================================================")

	wahaCount := 0
	totalCount := 0

	for rows.Next() {
		var idDevice, provider, phoneNumber, instance sql.NullString
		err := rows.Scan(&idDevice, &provider, &phoneNumber, &instance)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		totalCount++
		if provider.String == "waha" {
			wahaCount++
		}

		fmt.Printf("%-20s %-12s %-15s %-30s\n", 
			getStringValue(idDevice), 
			getStringValue(provider), 
			getStringValue(phoneNumber), 
			getStringValue(instance))
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total devices: %d\n", totalCount)
	fmt.Printf("   WAHA devices: %d\n", wahaCount)

	// Check for specific device ID from the browser URL
	testDeviceID := "04bbcc4-026d-4034-b82f-147ff3ca13bd"
	fmt.Printf("\n🔍 Checking specific device ID: %s\n", testDeviceID)

	var deviceExists bool
	var deviceProvider string
	checkQuery := `SELECT provider FROM device_setting_nodepath WHERE id_device = ?`
	err = db.QueryRow(checkQuery, testDeviceID).Scan(&deviceProvider)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("❌ Device %s not found in database\n", testDeviceID)
			deviceExists = false
		} else {
			log.Printf("Error checking device: %v", err)
		}
	} else {
		deviceExists = true
		fmt.Printf("✅ Device %s found with provider: %s\n", testDeviceID, deviceProvider)
	}

	if !deviceExists {
		fmt.Printf("\n💡 The device ID from the browser URL doesn't exist in the database.\n")
		fmt.Printf("   This explains the 'Failed to get device settings' error.\n")
		fmt.Printf("   You need to create a device with WAHA provider first.\n")
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}