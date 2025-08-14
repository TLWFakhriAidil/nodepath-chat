package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ChatbotFlow struct {
	ID        string          `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Niche     string          `json:"niche" db:"niche"`
	IdDevice  string          `json:"id_device" db:"id_device"`
	Nodes     *json.RawMessage `json:"nodes" db:"nodes"`
	Edges     *json.RawMessage `json:"edges" db:"edges"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

func main() {
	// Connect to database
	dsn := "root:123456@tcp(localhost:3306)/nodepath_chat?charset=utf8mb4&parseTime=True&loc=Local"
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

	// Check table schema
	fmt.Println("\nChecking table schema:")
	rows, err := db.Query("DESCRIBE chatbot_flows_nodepath")
	if err != nil {
		log.Fatal("Failed to describe table:", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-10s %-10s %-20s %-10s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("%-20s %-20s %-10s %-10s %-20s %-10s\n", 
			field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
	}

	// Test inserting a flow with id_device
	fmt.Println("\nTesting flow insertion with id_device:")
	nodesJSON := json.RawMessage(`[{"id":"start-1","type":"start","position":{"x":250,"y":100},"data":{"label":"Start"}}]`)
	edgesJSON := json.RawMessage(`[]`)
	testFlow := ChatbotFlow{
		ID:        "test_flow_" + fmt.Sprintf("%d", time.Now().Unix()),
		Name:      "Test Flow",
		Niche:     "test",
		IdDevice:  "SCHQ-552",
		Nodes:     &nodesJSON,
		Edges:     &edgesJSON,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO chatbot_flows_nodepath 
		(id, name, niche, id_device, nodes, edges, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query,
		testFlow.ID, testFlow.Name, testFlow.Niche, testFlow.IdDevice,
		testFlow.Nodes, testFlow.Edges, testFlow.CreatedAt, testFlow.UpdatedAt,
	)

	if err != nil {
		log.Fatal("Failed to insert test flow:", err)
	}

	fmt.Printf("Successfully inserted test flow with ID: %s and id_device: %s\n", testFlow.ID, testFlow.IdDevice)

	// Verify the insertion
	fmt.Println("\nVerifying insertion:")
	var retrievedFlow ChatbotFlow
	err = db.QueryRow(`
		SELECT id, name, niche, id_device, nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id = ?
	`, testFlow.ID).Scan(
		&retrievedFlow.ID, &retrievedFlow.Name, &retrievedFlow.Niche, &retrievedFlow.IdDevice,
		&retrievedFlow.Nodes, &retrievedFlow.Edges, &retrievedFlow.CreatedAt, &retrievedFlow.UpdatedAt,
	)

	if err != nil {
		log.Fatal("Failed to retrieve test flow:", err)
	}

	fmt.Printf("Retrieved flow - ID: %s, Name: %s, Niche: %s, id_device: %s\n", 
		retrievedFlow.ID, retrievedFlow.Name, retrievedFlow.Niche, retrievedFlow.IdDevice)

	// Clean up
	_, err = db.Exec("DELETE FROM chatbot_flows_nodepath WHERE id = ?", testFlow.ID)
	if err != nil {
		log.Printf("Warning: Failed to clean up test flow: %v\n", err)
	} else {
		fmt.Println("Test flow cleaned up successfully")
	}

	fmt.Println("\nTest completed successfully!")
}