package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection string
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	
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
	fmt.Println("Connected to database successfully")

	// Update NULL values to empty strings
	fmt.Println("Updating NULL values in niche column...")
	_, err = db.Exec("UPDATE chatbot_flows_nodepath SET niche = '' WHERE niche IS NULL")
	if err != nil {
		fmt.Printf("Error updating niche column: %v\n", err)
	} else {
		fmt.Println("Niche column updated successfully")
	}

	fmt.Println("Updating NULL values in id_device column...")
	_, err = db.Exec("UPDATE chatbot_flows_nodepath SET id_device = '' WHERE id_device IS NULL")
	if err != nil {
		fmt.Printf("Error updating id_device column: %v\n", err)
	} else {
		fmt.Println("Id_device column updated successfully")
	}

	// Set default values for the columns
	fmt.Println("Setting default values for columns...")
	_, err = db.Exec("ALTER TABLE chatbot_flows_nodepath MODIFY COLUMN niche VARCHAR(255) NOT NULL DEFAULT ''")
	if err != nil {
		fmt.Printf("Error setting default for niche: %v\n", err)
	} else {
		fmt.Println("Default value set for niche column")
	}

	_, err = db.Exec("ALTER TABLE chatbot_flows_nodepath MODIFY COLUMN id_device VARCHAR(255) NOT NULL DEFAULT ''")
	if err != nil {
		fmt.Printf("Error setting default for id_device: %v\n", err)
	} else {
		fmt.Println("Default value set for id_device column")
	}

	fmt.Println("Database update completed")
}