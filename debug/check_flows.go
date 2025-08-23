package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection
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

	fmt.Println("=== Checking flows in chatbot_flows_nodepath ===")

	// Check all flows
	allFlowsQuery := `SELECT id_device, name, id FROM chatbot_flows_nodepath ORDER BY id_device`
	allRows, err := db.Query(allFlowsQuery)
	if err != nil {
		log.Fatal("Failed to query all flows:", err)
	}
	defer allRows.Close()

	fmt.Println("All flows in database:")
	flowCount := 0
	schqS94Count := 0
	for allRows.Next() {
		var idDevice, name, id string
		err := allRows.Scan(&idDevice, &name, &id)
		if err != nil {
			log.Fatal("Failed to scan flow:", err)
		}
		fmt.Printf("ID: %s, Device: %s, Name: %s\n", id, idDevice, name)
		flowCount++
		if idDevice == "SCHQ-S94" {
			schqS94Count++
		}
	}

	fmt.Printf("\nTotal flows: %d\n", flowCount)
	fmt.Printf("Flows for SCHQ-S94: %d\n", schqS94Count)

	// Specifically check for SCHQ-S94
	fmt.Println("\n=== Checking specifically for SCHQ-S94 ===")
	schqQuery := `SELECT id, name, niche FROM chatbot_flows_nodepath WHERE id_device = 'SCHQ-S94'`
	schqRows, err := db.Query(schqQuery)
	if err != nil {
		log.Fatal("Failed to query SCHQ-S94 flows:", err)
	}
	defer schqRows.Close()

	schqFlows := 0
	for schqRows.Next() {
		var id, name, niche string
		err := schqRows.Scan(&id, &name, &niche)
		if err != nil {
			log.Fatal("Failed to scan SCHQ-S94 flow:", err)
		}
		fmt.Printf("SCHQ-S94 Flow - ID: %s, Name: %s, Niche: %s\n", id, name, niche)
		schqFlows++
	}

	if schqFlows == 0 {
		fmt.Println("❌ No flows found for SCHQ-S94 device")
		fmt.Println("This explains why the webhook handler says 'No flows configured'")
	} else {
		fmt.Printf("✅ Found %d flows for SCHQ-S94\n", schqFlows)
	}

	// Check device settings
	fmt.Println("\n=== Checking device settings for SCHQ-S94 ===")
	deviceQuery := `SELECT id_device, provider, instance FROM device_setting_nodepath WHERE id_device = 'SCHQ-S94'`
	var deviceID, provider, instance string
	err = db.QueryRow(deviceQuery).Scan(&deviceID, &provider, &instance)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No device settings found for SCHQ-S94")
		} else {
			log.Fatal("Failed to query device settings:", err)
		}
	} else {
		fmt.Printf("✅ Device settings found - Device: %s, Provider: %s, Instance: %s\n", deviceID, provider, instance)
	}

	fmt.Println("\n=== Summary ===")
	if schqFlows == 0 {
		fmt.Println("The issue is that SCHQ-S94 has no flows configured in the database.")
		fmt.Println("The system correctly falls back to AI conversation when no flows are found.")
		fmt.Println("To test flow routing, you need to create a flow for SCHQ-S94 device.")
	} else {
		fmt.Println("SCHQ-S94 has flows configured. The routing logic should work.")
	}
}