package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

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

	// Read SQL file
	sqlContent, err := ioutil.ReadFile("alter_schema.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Split SQL statements by semicolon
	statements := strings.Split(string(sqlContent), ";")
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		fmt.Printf("Executing: %s\n", stmt)
		_, err := db.Exec(stmt)
		if err != nil {
			fmt.Printf("Error executing statement: %v\n", err)
		} else {
			fmt.Println("Statement executed successfully")
		}
	}

	fmt.Println("Migration completed")
}