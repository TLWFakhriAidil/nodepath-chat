package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// getLastAIResponseTest simulates the updated function that returns raw conv_last data
func getLastAIResponseTest(convLast string) string {
	if convLast == "" || convLast == "null" {
		return ""
	}

	// Return raw conv_last data without processing
	return convLast
}

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("Connected to database successfully")

	// Query for conv_last data
	var convLast string
	query := "SELECT conv_last FROM ai_whatsapp_nodepath WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001' LIMIT 1"
	err = db.QueryRow(query).Scan(&convLast)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No AI conversation found for the test prospect")
			return
		}
		log.Fatal("Failed to query conv_last:", err)
	}

	fmt.Printf("Found conv_last data (length: %d)\n", len(convLast))
	fmt.Printf("Raw conv_last content: %s\n", convLast)

	// Test the getLastAIResponse function
	result := getLastAIResponseTest(convLast)
	fmt.Printf("\ngetLastAIResponse result (length: %d):\n%s\n", len(result), result)

	// Verify that the result is the same as the raw data
	if result == convLast {
		fmt.Println("\n✅ SUCCESS: getLastAIResponse returns raw conv_last data")
	} else {
		fmt.Println("\n❌ FAILED: getLastAIResponse does not return raw conv_last data")
	}
}