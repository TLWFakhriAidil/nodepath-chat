package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection string
	dsn := "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	
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

	// Check table structure
	rows, err := db.Query("DESCRIBE chatbot_flows_nodepath")
	if err != nil {
		log.Fatal("Failed to describe table:", err)
	}
	defer rows.Close()

	fmt.Println("Current table structure:")
	fmt.Println("Field\t\tType\t\tNull\tKey\tDefault\tExtra")
	fmt.Println("-----\t\t----\t\t----\t---\t-------\t-----")

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		
		fmt.Printf("%s\t\t%s\t\t%s\t%s\t%s\t%s\n", 
			field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
	}
}