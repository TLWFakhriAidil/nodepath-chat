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
	NodeTypeStart             NodeType = "start"
	NodeTypeAIPrompt          NodeType = "ai_prompt"
	NodeTypeAdvancedAIPrompt  NodeType = "advanced_ai_prompt"
	NodeTypeManual            NodeType = "manual"
	NodeTypeMessage           NodeType = "message"
	NodeTypeImage             NodeType = "image"
	NodeTypeAudio             NodeType = "audio"
	NodeTypeVideo             NodeType = "video"
	NodeTypeDelay             NodeType = "delay"
	NodeTypeCondition         NodeType = "condition"
	NodeTypeStage             NodeType = "stage"
	NodeTypeUserReply         NodeType = "user_reply"
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
	ID        string     `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	FullName  string     `json:"full_name" db:"full_name"`
	Password  string     `json:"-" db:"password"` // Don't include password in JSON responses
	Gmail     *string    `json:"gmail" db:"gmail"`
	Phone     *string    `json:"phone" db:"phone"`
	Status    string     `json:"status" db:"status"`
	Expired   *string    `json:"expired" db:"expired"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time `json:"last_login" db:"last_login"`
}

// UserProfileUpdate represents the payload for updating user profile
type UserProfileUpdate struct {
	FullName    string  `json:"full_name"`
	Gmail       *string `json:"gmail"`
	Phone       *string `json:"phone"`
	Password    *string `json:"password,omitempty"` // Optional password update
	NewPassword *string `json:"new_password,omitempty"`
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
	ID        string           `json:"id" db:"id"`
	Name      string           `json:"name" db:"name"`
	Niche     string           `json:"niche" db:"niche"`
	IdDevice  string           `json:"id_device" db:"id_device"`
	Nodes     *json.RawMessage `json:"nodes" db:"nodes"`
	Edges     *json.RawMessage `json:"edges" db:"edges"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

// FlowNode represents a single node in a flow
type FlowNode struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Position Position               `json:"position"`
}

// FlowEdge represents a connection between nodes
type FlowEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
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
	Role    string `json:"role"` // "USER" or "BOT"
	Content string `json:"content"`
}

// OpenRouterRequest represents a request to OpenRouter API
// Updated to match PHP payload structure with temperature, top_p, and repetition_penalty
type OpenRouterRequest struct {
	Model             string                 `json:"model"`
	Messages          []OpenRouterMessage    `json:"messages"`
	Stream            bool                   `json:"stream"`
	Temperature       float64                `json:"temperature"`        // Recommended setting: 0.67
	TopP              float64                `json:"top_p"`              // Keep responses within natural probability range: 1
	RepetitionPenalty float64                `json:"repetition_penalty"` // Avoid repetitive responses: 1
	Other             map[string]interface{} `json:"-"`
}

// OpenRouterMessage represents a message in OpenRouter format
type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents a response from OpenRouter API
type OpenRouterResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenRouterChoice `json:"choices"`
	Usage   OpenRouterUsage    `json:"usage"`
}

// OpenRouterChoice represents a choice in OpenRouter response
type OpenRouterChoice struct {
	Index        int               `json:"index"`
	Message      OpenRouterMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

// OpenRouterUsage represents usage statistics from OpenRouter
type OpenRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIPromptResponse represents a structured AI response for advanced prompt nodes
type AIPromptResponse struct {
	Stage    string           `json:"Stage"`
	Response []AIResponsePart `json:"Response"`
}

// AIResponsePart represents a single part of an AI response
type AIResponsePart struct {
	Type    string `json:"type"`              // "text" or "image"
	Content string `json:"content,omitempty"` // Text content
	URL     string `json:"url,omitempty"`     // Image URL
	Jenis   string `json:"Jenis,omitempty"`   // "onemessage" for combining text parts
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Test chat message struct removed

// AIWhatsapp represents an AI WhatsApp conversation record with flow execution capabilities
// Updated to match the new ai_whatsapp_nodepath schema - removed deprecated columns:
// jam, conv_stage, variables, catatan_staff, data_image, current_node, bot_balas
type AIWhatsapp struct {
	IDProspect      int            `json:"id_prospect" db:"id_prospect"`
	FlowReference   sql.NullString `json:"flow_reference" db:"flow_reference"` // Reference to chatbot flow being executed
	ExecutionID     sql.NullString `json:"execution_id" db:"execution_id"`     // Unique execution identifier
	DateOrder       *time.Time     `json:"date_order" db:"date_order"`
	IDDevice        string         `json:"id_device" db:"id_device"`
	Niche           string         `json:"niche" db:"niche"`
	ProspectName    sql.NullString `json:"prospect_name" db:"prospect_name"`
	ProspectNum     string         `json:"prospect_num" db:"prospect_num"`
	Intro           sql.NullString `json:"intro" db:"intro"` // Changed to sql.NullString to handle NULL values
	Stage           sql.NullString `json:"stage" db:"stage"`
	ConvLast        sql.NullString `json:"conv_last" db:"conv_last"` // Changed from json.RawMessage to sql.NullString for TEXT field
	ConvCurrent     sql.NullString `json:"conv_current" db:"conv_current"`
	ExecutionStatus sql.NullString `json:"execution_status" db:"execution_status"`   // Flow execution status (active, completed, failed)
	FlowID          sql.NullString `json:"flow_id" db:"flow_id"`                     // ID of the current chatbot flow being executed
	CurrentNodeID   sql.NullString `json:"current_node_id" db:"current_node_id"`     // Current node ID in the chatbot flow
	LastNodeID      sql.NullString `json:"last_node_id" db:"last_node_id"`           // Previous node ID for flow tracking
	WaitingForReply sql.NullInt32  `json:"waiting_for_reply" db:"waiting_for_reply"` // 1 = waiting for user reply, 0 = not waiting
	Balas           sql.NullString `json:"balas" db:"balas"`
	Human           int            `json:"human" db:"human"` // 0 = AI active, 1 = human takeover
	KeywordIklan    sql.NullString `json:"keywordiklan" db:"keywordiklan"`
	Marketer        sql.NullString `json:"marketer" db:"marketer"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	UpdateToday     *time.Time     `json:"update_today" db:"update_today"`
}

// ConversationLog represents a log entry for AI conversations
type ConversationLog struct {
	ID          int            `json:"id" db:"id"`
	ProspectNum string         `json:"prospect_num" db:"prospect_num"`
	IDDevice    string         `json:"id_device" db:"id_device"`
	Message     string         `json:"message" db:"message"`
	Sender      string         `json:"sender" db:"sender"` // 'user' or 'bot'
	Stage       sql.NullString `json:"stage" db:"stage"`
	Timestamp   time.Time      `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}
