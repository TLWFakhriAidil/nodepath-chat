package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection
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

	fmt.Println("=== Checking device settings for FakhriAidilTLW-001 ===")

	// Query device settings for FakhriAidilTLW-001
	query := `SELECT id_device, provider, instance FROM device_setting_nodepath WHERE id_device = ?`
	rows, err := db.Query(query, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatal("Failed to query device settings:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var idDevice, provider, instance string
		err := rows.Scan(&idDevice, &provider, &instance)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("✅ Device: %s, Provider: %s, Instance: %s\n", idDevice, provider, instance)
		found = true
	}

	if !found {
		fmt.Println("❌ No device settings found for FakhriAidilTLW-001")
	}

	fmt.Println("\n=== Checking all device settings ===")
	// Query all device settings
	allQuery := `SELECT id_device, provider, instance FROM device_setting_nodepath`
	allRows, err := db.Query(allQuery)
	if err != nil {
		log.Fatal("Failed to query all device settings:", err)
	}
	defer allRows.Close()

	count := 0
	for allRows.Next() {
		var idDevice, provider, instance string
		err := allRows.Scan(&idDevice, &provider, &instance)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("Device: %s, Provider: %s, Instance: %s\n", idDevice, provider, instance)
		count++
	}

	fmt.Printf("\nTotal devices: %d\n", count)
}