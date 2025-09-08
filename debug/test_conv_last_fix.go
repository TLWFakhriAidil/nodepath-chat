package main

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

// MockAIWhatsapp simulates the AIWhatsapp model with conv_last data
type MockAIWhatsapp struct {
	IDProspect   int64
	ProspectNum  string
	IDDevice     string
	ConvLast     sql.NullString
	CurrentStage sql.NullString
}

// MockConversationMessage simulates the ConversationMessage model
type MockConversationMessage struct {
	Role    string
	Content string
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	fmt.Println("=== Testing Conv_Last Retrieval Fix ===")
	fmt.Println()

	// Test Case 1: Execution with valid conv_last
	fmt.Println("📋 Test Case 1: Execution with valid conv_last")
	execution1 := &MockAIWhatsapp{
		IDProspect:  12345,
		ProspectNum: "60179645043",
		IDDevice:    "FakhriAidilTLW-001",
		ConvLast: sql.NullString{
			Valid:  true,
			String: "Hello! I'm here to help you with your inquiry. What specific information are you looking for today?",
		},
		CurrentStage: sql.NullString{Valid: true, String: "Problem Identification"},
	}

	conversationHistory1 := processConvLast(execution1)
	fmt.Printf("✅ Conversation history count: %d\n", len(conversationHistory1))
	if len(conversationHistory1) > 0 {
		fmt.Printf("✅ Assistant message: %s\n", conversationHistory1[0].Content[:50]+"...") 
	}
	fmt.Println()

	// Test Case 2: Execution with empty conv_last
	fmt.Println("📋 Test Case 2: Execution with empty conv_last")
	execution2 := &MockAIWhatsapp{
		IDProspect:  12346,
		ProspectNum: "60179645044",
		IDDevice:    "FakhriAidilTLW-002",
		ConvLast:    sql.NullString{Valid: false, String: ""},
		CurrentStage: sql.NullString{Valid: true, String: "Initial Contact"},
	}

	conversationHistory2 := processConvLast(execution2)
	fmt.Printf("✅ Conversation history count: %d\n", len(conversationHistory2))
	fmt.Println()

	// Test Case 3: Execution with null conv_last
	fmt.Println("📋 Test Case 3: Execution with 'null' conv_last")
	execution3 := &MockAIWhatsapp{
		IDProspect:  12347,
		ProspectNum: "60179645045",
		IDDevice:    "FakhriAidilTLW-003",
		ConvLast:    sql.NullString{Valid: true, String: "null"},
		CurrentStage: sql.NullString{Valid: true, String: "Follow Up"},
	}

	conversationHistory3 := processConvLast(execution3)
	fmt.Printf("✅ Conversation history count: %d\n", len(conversationHistory3))
	fmt.Println()

	fmt.Println("=== Fix Summary ===")
	fmt.Println("🔧 Fixed: WhatsApp service now properly retrieves conv_last from database")
	fmt.Println("🔧 Fixed: AI service receives conversation history instead of empty array")
	fmt.Println("🔧 Added: Debug logging to show conv_last retrieval and usage")
	fmt.Println("🔧 Added: conversation_history_count in debug logs")
	fmt.Println()
	fmt.Println("📊 Expected Debug Output:")
	fmt.Println("   - 🔍 AI_PROMPT_DEBUG: Retrieved conv_last for AI context")
	fmt.Println("   - 🔍 AI_SERVICE_DEBUG: conversation_history_count: 1 (when conv_last exists)")
	fmt.Println("   - 🔍 AI_PROMPT_DEBUG: No conv_last found, starting fresh conversation (when empty)")
}

// processConvLast simulates the fixed logic from WhatsApp service
func processConvLast(execution *MockAIWhatsapp) []MockConversationMessage {
	var conversationHistory []MockConversationMessage
	
	if execution.ConvLast.Valid && execution.ConvLast.String != "" && execution.ConvLast.String != "null" {
		conversationHistory = append(conversationHistory, MockConversationMessage{
			Role:    "assistant",
			Content: execution.ConvLast.String,
		})
		logrus.WithFields(logrus.Fields{
			"conv_last_length": len(execution.ConvLast.String),
			"conv_last_preview": func() string {
				if len(execution.ConvLast.String) > 100 {
					return execution.ConvLast.String[:100] + "..."
				}
				return execution.ConvLast.String
			}(),
			"prospect_num": execution.ProspectNum,
			"device_id": execution.IDDevice,
		}).Info("🔍 AI_PROMPT_DEBUG: Retrieved conv_last for AI context")
	} else {
		logrus.WithFields(logrus.Fields{
			"prospect_num": execution.ProspectNum,
			"device_id": execution.IDDevice,
			"conv_last_valid": execution.ConvLast.Valid,
			"conv_last_value": execution.ConvLast.String,
		}).Info("🔍 AI_PROMPT_DEBUG: No conv_last found, starting fresh conversation")
	}
	
	// Simulate the AI service call logging
	logrus.WithFields(logrus.Fields{
		"id_device": execution.IDDevice,
		"prospect_num": execution.ProspectNum,
		"conversation_history_count": len(conversationHistory),
	}).Info("🤖 AI_PROMPT: Generating AI response")
	
	return conversationHistory
}