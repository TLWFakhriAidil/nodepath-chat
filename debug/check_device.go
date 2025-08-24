package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert to proper DSN format
	// mysql://user:password@host:port/database -> user:password@tcp(host:port)/database
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Checking device_setting_nodepath for FakhriAidilTLW-001 ===")

	// Check if device exists in device_setting_nodepath
	query := `SELECT id, device_id, provider, id_device, phone_number FROM device_setting_nodepath WHERE id_device = ?`
	rows, err := db.Query(query, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id, deviceID, provider, idDevice, phoneNumber sql.NullString
		err := rows.Scan(&id, &deviceID, &provider, &idDevice, &phoneNumber)
		if err != nil {
			log.Fatal("Scan failed:", err)
		}
		fmt.Printf("Found device: ID=%s, DeviceID=%s, Provider=%s, IDDevice=%s, Phone=%s\n",
			id.String, deviceID.String, provider.String, idDevice.String, phoneNumber.String)
	}

	if !found {
		fmt.Println("❌ Device FakhriAidilTLW-001 NOT found in device_setting_nodepath")
		fmt.Println("\n=== Checking all devices in device_setting_nodepath ===")
		
		// Show all devices
		allQuery := `SELECT id_device, provider FROM device_setting_nodepath LIMIT 10`
		allRows, err := db.Query(allQuery)
		if err != nil {
			log.Fatal("All devices query failed:", err)
		}
		defer allRows.Close()
		
		for allRows.Next() {
			var idDevice, provider sql.NullString
			err := allRows.Scan(&idDevice, &provider)
			if err != nil {
				log.Fatal("All devices scan failed:", err)
			}
			fmt.Printf("Device: %s (Provider: %s)\n", idDevice.String, provider.String)
		}
	} else {
		fmt.Println("✅ Device found in device_setting_nodepath")
	}

	fmt.Println("\n=== Checking whatsmeow_device table ===")
	// Check whatsmeow_device table for WhatsApp Web devices
	whatsmeowQuery := `SELECT jid, push_name, platform FROM whatsmeow_device LIMIT 5`
	whatsmeowRows, err := db.Query(whatsmeowQuery)
	if err != nil {
		fmt.Printf("WhatsApp Web devices query failed (table might not exist): %v\n", err)
	} else {
		defer whatsmeowRows.Close()
		fmt.Println("WhatsApp Web devices:")
		for whatsmeowRows.Next() {
			var jid, pushName, platform sql.NullString
			err := whatsmeowRows.Scan(&jid, &pushName, &platform)
			if err != nil {
				log.Fatal("WhatsApp Web scan failed:", err)
			}
			fmt.Printf("JID: %s, Name: %s, Platform: %s\n", jid.String, pushName.String, platform.String)
		}
	}
}