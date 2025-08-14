package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection
	dsn := "root:nodepath123@tcp(127.0.0.1:3306)/nodepath_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println("\nQuerying device settings...")

	// Query device settings
	query := `
		SELECT id, provider, COALESCE(id_device, 'NULL') as id_device, 
		       COALESCE(instance, 'NULL') as instance, 
		       COALESCE(phone_number, 'NULL') as phone_number,
		       COALESCE(device_id, 'NULL') as device_id
		FROM device_setting_nodepath 
		WHERE provider IN ('whacenter', 'wablas')
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query device settings:", err)
	}
	defer rows.Close()

	fmt.Printf("%-36s %-12s %-20s %-20s %-15s %-20s\n", "ID", "Provider", "ID Device", "Instance", "Phone Number", "Device ID")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	count := 0
	for rows.Next() {
		var id, provider, idDevice, instance, phoneNumber, deviceID string
		err := rows.Scan(&id, &provider, &idDevice, &instance, &phoneNumber, &deviceID)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("%-36s %-12s %-20s %-20s %-15s %-20s\n", id, provider, idDevice, instance, phoneNumber, deviceID)
		count++
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Error iterating rows:", err)
	}

	fmt.Printf("\nTotal devices found: %d\n", count)

	if count == 0 {
		fmt.Println("\nNo Whacenter or Wablas devices found in the database.")
		fmt.Println("This explains why the status check is returning 404 - there are no devices to check!")
	}
}