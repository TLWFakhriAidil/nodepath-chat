package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// FlowMode represents the execution mode of a chatbot flow
type FlowMode string

const (
	FlowModeAuto     FlowMode = "AUTO"
	FlowModeSemiAuto FlowMode = "SEMI-AUTO"
	FlowModeManual   FlowMode = "MANUAL"
)

// NodeType represents the type of a flow node
type NodeType string

const (
	NodeTypeStart            NodeType = "start"
	NodeTypeAIPrompt         NodeType = "ai_prompt"
	NodeTypeAdvancedAIPrompt NodeType = "advanced_ai_prompt"
	NodeTypeManual           NodeType = "manual"
	NodeTypeMessage          NodeType = "message"
	NodeTypeImage            NodeType = "image"
	NodeTypeAudio            NodeType = "audio"
	NodeTypeVideo            NodeType = "video"
	NodeTypeDelay            NodeType = "delay"
	NodeTypeCondition        NodeType = "condition"
	NodeTypeStage            NodeType = "stage"
	NodeTypeUserReply        NodeType = "user_reply"
	NodeTypeWaitingReplyTimes NodeType = "waiting_reply_times"
)

// ExecutionStatus represents the status of a flow execution
type ExecutionStatus string

const (
	ExecutionStatusActive    ExecutionStatus = "active"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
)



// User represents a user in the authentication system
type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	FullName  string    `json:"full_name" db:"full_name"`
	Password  string    `json:"-" db:"password"` // Don't include password in JSON responses
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceSetting represents device configuration linked to a user
type DeviceSetting struct {
	ID           string    `json:"id" db:"id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	APIKeyOption string    `json:"api_key_option" db:"api_key_option"`
	WebhookID    string    `json:"webhook_id" db:"webhook_id"`
	Provider     string    `json:"provider" db:"provider"`
	PhoneNumber  string    `json:"phone_number" db:"phone_number"`
	APIKey       string    `json:"-" db:"api_key"` // Don't include API key in JSON responses
	IDDevice     string    `json:"id_device" db:"id_device"`
	IDERP        string    `json:"id_erp" db:"id_erp"`
	IDAdmin      string    `json:"id_admin" db:"id_admin"`
	UserID       *int      `json:"user_id" db:"user_id"`
	Instance     string    `json:"instance" db:"instance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeText  MediaType = "text"
	MediaTypeImage MediaType = "image"
	MediaTypeAudio MediaType = "audio"
	MediaTypeVideo MediaType = "video"
)

// ChatbotFlow represents a chatbot flow configuration
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

// FlowNode represents a single node in a flow
type FlowNode struct {
	ID           string                 `json:"id"`
	Type         NodeType               `json:"type"`
	Data         map[string]interface{} `json:"data"`
	Position     Position               `json:"position"`
}

// FlowEdge represents a connection between nodes
type FlowEdge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// Position represents the position of a node in the flow builder
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ChatbotExecution struct removed - functionality consolidated into AIWhatsapp

// ConversationMessage represents a single message in a conversation
type ConversationMessage struct {
	Role    string `json:"role"`    // "USER" or "BOT"
	Content string `json:"content"`
}



// OpenRouterRequest represents a request to OpenRouter API
type OpenRouterRequest struct {
	Model    string                   `json:"model"`
	Messages []OpenRouterMessage     `json:"messages"`
	Stream   bool                     `json:"stream"`
	Other    map[string]interface{}   `json:"-"`
}

// OpenRouterMessage represents a message in OpenRouter format
type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents a response from OpenRouter API
type OpenRouterResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []OpenRouterChoice     `json:"choices"`
	Usage   OpenRouterUsage        `json:"usage"`
}

// OpenRouterChoice represents a choice in OpenRouter response
type OpenRouterChoice struct {
	Index        int                   `json:"index"`
	Message      OpenRouterMessage     `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// OpenRouterUsage represents usage statistics from OpenRouter
type OpenRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIPromptResponse represents a structured AI response for advanced prompt nodes
type AIPromptResponse struct {
	Stage    string            `json:"Stage"`
	Response []AIResponsePart  `json:"Response"`
}

// AIResponsePart represents a single part of an AI response
type AIResponsePart struct {
	Type    string `json:"type"`     // "text" or "image"
	Content string `json:"content,omitempty"`  // Text content
	URL     string `json:"url,omitempty"`      // Image URL
	Jenis   string `json:"Jenis,omitempty"`    // "onemessage" for combining text parts
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Test chat message struct removed

// AIWhatsapp represents an AI WhatsApp conversation record with flow execution capabilities
type AIWhatsapp struct {
	IDProspect      int             `json:"id_prospect" db:"id_prospect"`
	IDDevice        string          `json:"id_device" db:"id_device"`
	ProspectNum     string          `json:"prospect_num" db:"prospect_num"`
	Stage           string          `json:"stage" db:"stage"`
	DateOrder       *time.Time      `json:"date_order" db:"date_order"`
	ConvLast        json.RawMessage `json:"conv_last" db:"conv_last"`
	ConvCurrent     sql.NullString  `json:"conv_current" db:"conv_current"`
	Jam             string          `json:"jam" db:"jam"`
	Intro           string          `json:"intro" db:"intro"`
	Human           int             `json:"human" db:"human"` // 0 = AI active, 1 = human takeover
	CatatanStaff    string          `json:"catatan_staff" db:"catatan_staff"`
	Balas           int             `json:"balas" db:"balas"`
	DataImage       string          `json:"data_image" db:"data_image"`
	ConvStage       string          `json:"conv_stage" db:"conv_stage"`
	Niche           string          `json:"niche" db:"niche"`
	BotBalas        *time.Time      `json:"bot_balas" db:"bot_balas"`
	KeywordIklan    string          `json:"keywordiklan" db:"keywordiklan"`
	Marketer        string          `json:"marketer" db:"marketer"`
	UpdateToday     *time.Time      `json:"update_today" db:"update_today"`
	// Flow execution fields
	FlowReference   sql.NullString  `json:"flow_reference" db:"flow_reference"`   // Reference to chatbot flow being executed
	CurrentNode     sql.NullString  `json:"current_node" db:"current_node"`       // Current node in the flow execution
	Variables       json.RawMessage `json:"variables" db:"variables"`             // Flow execution variables (JSON)
	ExecutionStatus sql.NullString  `json:"execution_status" db:"execution_status"` // Flow execution status (active, completed, failed)
	ExecutionID     sql.NullString  `json:"execution_id" db:"execution_id"`       // Unique execution identifier
	// Flow tracking fields for user reply handling
	CurrentNodeID   sql.NullString  `json:"current_node_id" db:"current_node_id"`   // Current node ID in the chatbot flow
	WaitingForReply sql.NullInt32   `json:"waiting_for_reply" db:"waiting_for_reply"` // 1 = waiting for user reply, 0 = not waiting
	FlowID          sql.NullString  `json:"flow_id" db:"flow_id"`                 // ID of the current chatbot flow being executed
	LastNodeID      sql.NullString  `json:"last_node_id" db:"last_node_id"`       // Previous node ID for flow tracking
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}



// ConversationLog represents a log entry for AI conversations
type ConversationLog struct {
	ID          int       `json:"id" db:"id"`
	ProspectNum string    `json:"prospect_num" db:"prospect_num"`
	IDDevice    string    `json:"id_device" db:"id_device"`
	Message     string    `json:"message" db:"message"`
	Sender      string    `json:"sender" db:"sender"` // 'user' or 'bot'
	Stage       string    `json:"stage" db:"stage"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}