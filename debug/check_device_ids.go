package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable not set")
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	query := `SELECT id, id_device, provider, phone_number, instance FROM device_setting_nodepath WHERE provider = "waha"`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-12s %-15s %-30s\n", "ID", "ID_DEVICE", "PROVIDER", "PHONE_NUMBER", "INSTANCE")
	fmt.Println("===========================================================================================================")

	for rows.Next() {
		var id, idDevice, provider, phoneNumber, instance sql.NullString
		err := rows.Scan(&id, &idDevice, &provider, &phoneNumber, &instance)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("%-20s %-20s %-12s %-15s %-30s\n", 
			getStringValue(id), 
			getStringValue(idDevice), 
			getStringValue(provider), 
			getStringValue(phoneNumber), 
			getStringValue(instance))
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}