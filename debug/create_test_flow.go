package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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
	// Get database URI from environment or use default
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	} else {
		// Convert mysql:// format to Go driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Replace @ with @tcp( and add ) after port
			parts := strings.Split(mysqlURI, "/")
			if len(parts) >= 2 {
				connPart := parts[0]
				dbName := parts[1]
				// Split user:pass@host:port
				atIndex := strings.LastIndex(connPart, "@")
				if atIndex != -1 {
					userPass := connPart[:atIndex]
					hostPort := connPart[atIndex+1:]
					mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userPass, hostPort, dbName)
				}
			}
		}
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

	fmt.Println("Connected to database successfully!")

	// Create test flow with AI prompt nodes
	testFlowID := uuid.New().String()
	testDeviceID := "FakhriAidilTLW-001"

	// Define nodes with AI prompt node
	nodesData := []map[string]interface{}{
		{
			"id":   "start-1",
			"type": "start",
			"data": map[string]interface{}{
				"label": "Start",
			},
			"position": map[string]interface{}{
				"x": 100,
				"y": 100,
			},
		},
		{
			"id":   "ai-prompt-1",
			"type": "ai_prompt",
			"data": map[string]interface{}{
				"label":        "AI Welcome Prompt",
				"systemPrompt": "You are a helpful customer service assistant for a health and wellness company. Your name is Layla. You help customers with their health concerns and product inquiries. Always be friendly, professional, and empathetic. Start conversations by greeting the customer warmly and asking how you can help them today.",
				"instance":     "default",
				"openRouterKey": "",
			},
			"position": map[string]interface{}{
				"x": 300,
				"y": 100,
			},
		},
		{
			"id":   "ai-prompt-2",
			"type": "ai_prompt",
			"data": map[string]interface{}{
				"label":        "Problem Identification",
				"systemPrompt": "You are Layla, a health consultant. Help identify the customer's main health concern. Ask specific questions about symptoms like appetite loss, constipation, or frequent fever. Be caring and thorough in your questioning to understand their needs better.",
				"instance":     "default",
				"openRouterKey": "",
			},
			"position": map[string]interface{}{
				"x": 500,
				"y": 100,
			},
		},
	}

	// Define edges connecting the nodes
	edgesData := []map[string]interface{}{
		{
			"id":     "edge-1",
			"source": "start-1",
			"target": "ai-prompt-1",
		},
		{
			"id":     "edge-2",
			"source": "ai-prompt-1",
			"target": "ai-prompt-2",
		},
	}

	// Convert to JSON
	nodesJSON, err := json.Marshal(nodesData)
	if err != nil {
		log.Fatal("Failed to marshal nodes:", err)
	}

	edgesJSON, err := json.Marshal(edgesData)
	if err != nil {
		log.Fatal("Failed to marshal edges:", err)
	}

	nodesRaw := json.RawMessage(nodesJSON)
	edgesRaw := json.RawMessage(edgesJSON)

	// Create test flow
	testFlow := ChatbotFlow{
		ID:        testFlowID,
		Name:      "Test Health Consultation Flow",
		Niche:     "health-consultation",
		IdDevice:  testDeviceID,
		Nodes:     &nodesRaw,
		Edges:     &edgesRaw,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert test flow
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

	fmt.Printf("✅ Successfully created test flow with ID: %s for device: %s\n", testFlow.ID, testFlow.IdDevice)
	fmt.Printf("📋 Flow Name: %s\n", testFlow.Name)
	fmt.Printf("🏷️ Niche: %s\n", testFlow.Niche)
	fmt.Printf("🤖 Contains %d AI prompt nodes\n", len(nodesData)-1) // Exclude start node

	// Verify the insertion
	fmt.Println("\n🔍 Verifying insertion:")
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

	fmt.Printf("✅ Retrieved flow - ID: %s, Name: %s, Device: %s\n", 
		retrievedFlow.ID, retrievedFlow.Name, retrievedFlow.IdDevice)

	fmt.Println("\n🎉 Test flow created successfully! You can now test the flow-based AI prompt system.")
	fmt.Printf("📱 Test with device: %s and phone: 60179645043\n", testDeviceID)
	fmt.Println("💬 Send a message to trigger the flow-based AI response.")
}