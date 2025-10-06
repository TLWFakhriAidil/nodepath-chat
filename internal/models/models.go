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
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time `json:"last_login" db:"last_login"`
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

// ========================================
// BILLING AND SUBSCRIPTION MODELS
// ========================================

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
	SubscriptionStatusPending   SubscriptionStatus = "pending"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Subscription represents a user subscription
type Subscription struct {
	ID              string              `json:"id" db:"id"`
	UserID          string              `json:"user_id" db:"user_id"`
	PlanName        string              `json:"plan_name" db:"plan_name"`
	PlanPrice       float64             `json:"plan_price" db:"plan_price"`
	PlanPeriod      string              `json:"plan_period" db:"plan_period"`
	Status          SubscriptionStatus  `json:"status" db:"status"`
	NextBillingDate time.Time           `json:"next_billing_date" db:"next_billing_date"`
	Features        *json.RawMessage    `json:"features" db:"features"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
}

// Payment represents a payment record
type Payment struct {
	ID             string        `json:"id" db:"id"`
	SubscriptionID sql.NullString `json:"subscription_id" db:"subscription_id"`
	UserID         string        `json:"user_id" db:"user_id"`
	BillID         sql.NullString `json:"bill_id" db:"bill_id"`
	InvoiceNumber  string        `json:"invoice_number" db:"invoice_number"`
	Amount         float64       `json:"amount" db:"amount"`
	Currency       string        `json:"currency" db:"currency"`
	Description    sql.NullString `json:"description" db:"description"`
	Status         PaymentStatus `json:"status" db:"status"`
	PaymentMethod  string        `json:"payment_method" db:"payment_method"`
	BillplzURL     sql.NullString `json:"billplz_url" db:"billplz_url"`
	PaidAt         *time.Time    `json:"paid_at" db:"paid_at"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// BillingHistory represents a billing history record for display
type BillingHistory struct {
	ID            string         `json:"id" db:"id"`
	UserID        string         `json:"user_id" db:"user_id"`
	PaymentID     sql.NullString `json:"payment_id" db:"payment_id"`
	InvoiceNumber string         `json:"invoice_number" db:"invoice_number"`
	Amount        float64        `json:"amount" db:"amount"`
	Currency      string         `json:"currency" db:"currency"`
	Description   string         `json:"description" db:"description"`
	Status        PaymentStatus  `json:"status" db:"status"`
	PaymentDate   *time.Time     `json:"payment_date" db:"payment_date"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}

// BillplzPaymentRequest represents a request to create a Billplz payment
type BillplzPaymentRequest struct {
	CollectionID string  `json:"collection_id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	Amount       int     `json:"amount"` // Amount in cents
	Description  string  `json:"description"`
	CallbackURL  string  `json:"callback_url"`
	RedirectURL  string  `json:"redirect_url"`
	Reference1   string  `json:"reference_1"`
	Reference1Label string `json:"reference_1_label"`
}

// BillplzPaymentResponse represents a response from Billplz API
type BillplzPaymentResponse struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CreatePaymentRequest represents a request to create a new payment
type CreatePaymentRequest struct {
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	CustomerEmail string  `json:"customer_email"`
	CustomerName  string  `json:"customer_name"`
}

// BillingResponse represents the response for billing data
type BillingResponse struct {
	Subscription    *Subscription     `json:"subscription"`
	BillingHistory  []BillingHistory  `json:"billing_history"`
	TotalCount      int               `json:"total_count"`
}
