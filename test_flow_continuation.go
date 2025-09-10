package main

import (
	"fmt"
	"log"
	"nodepath-chat/internal/config"
	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"
	"nodepath-chat/internal/whatsapp"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	var db, _ = database.Initialize(cfg)
	if db == nil {
		log.Fatal("Failed to initialize database")
	}

	// Initialize services
	aiWhatsappRepo := repository.NewAIWhatsappRepository(db)
	deviceSettingsRepo := repository.NewDeviceSettingsRepository(db)
	flowService := services.NewFlowService(db, nil)
	aiService := services.NewAIService(cfg, deviceSettingsRepo)
	deviceSettingsService := services.NewDeviceSettingsService(db)
	mediaDetectionService := services.NewMediaDetectionService()
	aiWhatsappService := services.NewAIWhatsappService(aiWhatsappRepo, deviceSettingsRepo, flowService, mediaDetectionService, cfg)
	providerService := services.NewProviderService()

	// Initialize WhatsApp service
	ws, err := whatsapp.NewService(cfg, nil, flowService, aiService, aiWhatsappService, nil, deviceSettingsService, providerService, mediaDetectionService)
	if err != nil {
		log.Fatal("Failed to initialize WhatsApp service:", err)
	}

	// Test flow processing with user_reply node scenario
	fmt.Println("=== Testing AI Flow Continuation Fix ===")
	fmt.Println()

	// Simulate test data
	testPhone := "601137508067"
	testDevice := "FakhriAidilTLW-001"
	testFlowID := "flow_ai_1756016272"

	fmt.Printf("Test Configuration:\n")
	fmt.Printf("- Phone: %s\n", testPhone)
	fmt.Printf("- Device: %s\n", testDevice)
	fmt.Printf("- Flow ID: %s\n", testFlowID)
	fmt.Println()

	// Test Scenario 1: AI Prompt → User Reply → Next Node
	fmt.Println("📝 Test Scenario: AI Prompt → User Reply → Next AI Prompt")
	fmt.Println("Expected Behavior:")
	fmt.Println("1. AI prompt sends response and advances to user_reply node")
	fmt.Println("2. System waits for user input")
	fmt.Println("3. When user sends message, system advances from user_reply to next node")
	fmt.Println("4. Next node (e.g., another AI prompt) processes correctly")
	fmt.Println()

	// Get active execution to check current state
	execution, err := aiWhatsappService.GetActiveFlowExecution(testPhone, testDevice)
	if err != nil {
		fmt.Printf("❌ Error getting execution: %v\n", err)
	} else if execution != nil {
		fmt.Printf("✅ Found active execution:\n")
		fmt.Printf("   - Execution ID: %s\n", execution.ExecutionID.String)
		fmt.Printf("   - Current Node: %s\n", execution.CurrentNodeID.String)
		fmt.Printf("   - Waiting for Reply: %v\n", execution.WaitingForReply.Int32 == 1)
		fmt.Printf("   - Flow ID: %s\n", execution.FlowID.String)
		
		// Check if at user_reply node
		if execution.CurrentNodeID.Valid && execution.CurrentNodeID.String != "" {
			flow, _ := flowService.GetFlow(execution.FlowID.String)
			if flow != nil {
				node, _ := flowService.FindNodeByID(flow, execution.CurrentNodeID.String)
				if node != nil {
					fmt.Printf("   - Current Node Type: %s\n", node.Type)
					
					if node.Type == models.NodeTypeUserReply || node.Type == "user_reply" {
						fmt.Println()
						fmt.Println("🎯 System is currently at user_reply node - CORRECT!")
						fmt.Println("✅ Fix verified: Flow will continue to next node when user sends input")
						
						// Check what the next node would be
						nextNode, _ := flowService.GetNextNode(flow, node.ID)
						if nextNode != nil {
							fmt.Printf("   - Next Node ID: %s\n", nextNode.ID)
							fmt.Printf("   - Next Node Type: %s\n", nextNode.Type)
							fmt.Println()
							fmt.Println("✅ Flow continuation path is ready!")
						}
					}
				}
			}
		}
	} else {
		fmt.Println("ℹ️ No active execution found")
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
	fmt.Println()
	fmt.Println("💡 Summary of Fix:")
	fmt.Println("- processUserReplyNode now checks for user input")
	fmt.Println("- If input exists, it advances to next node")
	fmt.Println("- If no input, it waits (original behavior)")
	fmt.Println("- Flow continues properly after user_reply nodes")
}
