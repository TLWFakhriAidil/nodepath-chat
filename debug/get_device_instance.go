package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Getting instance for FakhriAidilTLW-001 ===")

	// Query device settings
	query := `SELECT id_device, provider, instance FROM device_setting_nodepath WHERE id_device = ?`
	rows, err := db.Query(query, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatal("Failed to query device:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var idDevice, provider string
		var instance sql.NullString

		err := rows.Scan(&idDevice, &provider, &instance)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("Device: %s\n", idDevice)
		fmt.Printf("Provider: %s\n", provider)
		if instance.Valid {
			fmt.Printf("Instance: %s\n", instance.String)
		} else {
			fmt.Println("Instance: NULL")
		}
		fmt.Println("---")
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}
}