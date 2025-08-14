package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// BroadcastHandlers handles HTTP requests for broadcast operations
type BroadcastHandlers struct {
	broadcastService *services.BroadcastService
}

// NewBroadcastHandlers creates a new broadcast handlers instance
func NewBroadcastHandlers(broadcastService *services.BroadcastService) *BroadcastHandlers {
	return &BroadcastHandlers{
		broadcastService: broadcastService,
	}
}

// QueueCampaignMessageRequest represents the request for queuing a campaign message
type QueueCampaignMessageRequest struct {
	UserID         string `json:"user_id"`
	DeviceID       string `json:"device_id"`
	CampaignID     string `json:"campaign_id"`
	RecipientPhone string `json:"recipient_phone"`
	MessageType    string `json:"message_type"`
	Content        string `json:"content"`
	MediaURL       string `json:"media_url,omitempty"`
	MinDelay       int    `json:"min_delay,omitempty"`
	MaxDelay       int    `json:"max_delay,omitempty"`
}

// QueueSequenceMessageRequest represents the request for queuing a sequence message
type QueueSequenceMessageRequest struct {
	UserID         string `json:"user_id"`
	DeviceID       string `json:"device_id"`
	SequenceID     string `json:"sequence_id"`
	SequenceStepID string `json:"sequence_step_id"`
	RecipientPhone string `json:"recipient_phone"`
	MessageType    string `json:"message_type"`
	Content        string `json:"content"`
	MediaURL       string `json:"media_url,omitempty"`
	MinDelay       int    `json:"min_delay,omitempty"`
	MaxDelay       int    `json:"max_delay,omitempty"`
	StepDelay      int    `json:"step_delay,omitempty"` // in seconds
}

// QueueBulkMessagesRequest represents the request for queuing multiple messages
type QueueBulkMessagesRequest struct {
	Messages []models.BroadcastMessage `json:"messages"`
}

// MessageResponse represents a standard message response
type MessageResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"`
}

// BulkMessageResponse represents a bulk message response
type BulkMessageResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	MessageIDs []string `json:"message_ids,omitempty"`
	Count      int      `json:"count"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// MetricsResponse represents a metrics response
type MetricsResponse struct {
	Success bool                      `json:"success"`
	Metrics *services.BroadcastMetrics `json:"metrics"`
}

// QueueStatsResponse represents queue statistics response
type QueueStatsResponse struct {
	Success bool           `json:"success"`
	Stats   map[string]int `json:"stats"`
}

// QueueCampaignMessage handles queuing a campaign message
func (bh *BroadcastHandlers) QueueCampaignMessage(w http.ResponseWriter, r *http.Request) {
	var req QueueCampaignMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.DeviceID == "" || req.CampaignID == "" || req.RecipientPhone == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	messageID, err := bh.broadcastService.QueueCampaignMessage(
		req.UserID, req.DeviceID, req.CampaignID, req.RecipientPhone,
		req.MessageType, req.Content, req.MediaURL, req.MinDelay, req.MaxDelay,
	)
	if err != nil {
		logrus.Errorf("Failed to queue campaign message: %v", err)
		http.Error(w, "Failed to queue message", http.StatusInternalServerError)
		return
	}

	response := MessageResponse{
		Success:   true,
		Message:   "Campaign message queued successfully",
		MessageID: messageID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// QueueSequenceMessage handles queuing a sequence message
func (bh *BroadcastHandlers) QueueSequenceMessage(w http.ResponseWriter, r *http.Request) {
	var req QueueSequenceMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.DeviceID == "" || req.SequenceID == "" || req.SequenceStepID == "" || req.RecipientPhone == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	stepDelay := time.Duration(req.StepDelay) * time.Second
	messageID, err := bh.broadcastService.QueueSequenceMessage(
		req.UserID, req.DeviceID, req.SequenceID, req.SequenceStepID, req.RecipientPhone,
		req.MessageType, req.Content, req.MediaURL, req.MinDelay, req.MaxDelay, stepDelay,
	)
	if err != nil {
		logrus.Errorf("Failed to queue sequence message: %v", err)
		http.Error(w, "Failed to queue message", http.StatusInternalServerError)
		return
	}

	response := MessageResponse{
		Success:   true,
		Message:   "Sequence message queued successfully",
		MessageID: messageID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// QueueBulkMessages handles queuing multiple messages
func (bh *BroadcastHandlers) QueueBulkMessages(w http.ResponseWriter, r *http.Request) {
	var req QueueBulkMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "No messages provided", http.StatusBadRequest)
		return
	}

	messageIDs, err := bh.broadcastService.QueueBulkMessages(req.Messages)
	if err != nil {
		logrus.Errorf("Failed to queue bulk messages: %v", err)
		http.Error(w, "Failed to queue messages", http.StatusInternalServerError)
		return
	}

	response := BulkMessageResponse{
		Success:    true,
		Message:    "Bulk messages queued successfully",
		MessageIDs: messageIDs,
		Count:      len(messageIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMessageStatus handles getting message status
func (bh *BroadcastHandlers) GetMessageStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["messageId"]

	if messageID == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	status, err := bh.broadcastService.GetMessageStatus(messageID)
	if err != nil {
		logrus.Errorf("Failed to get message status: %v", err)
		http.Error(w, "Failed to get message status", http.StatusInternalServerError)
		return
	}

	response := StatusResponse{
		Success: true,
		Status:  status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetrics handles getting system metrics
func (bh *BroadcastHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := bh.broadcastService.GetMetrics()

	response := MetricsResponse{
		Success: true,
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetQueueStats handles getting queue statistics
func (bh *BroadcastHandlers) GetQueueStats(w http.ResponseWriter, r *http.Request) {
	stats, err := bh.broadcastService.GetQueueStats()
	if err != nil {
		logrus.Errorf("Failed to get queue stats: %v", err)
		http.Error(w, "Failed to get queue stats", http.StatusInternalServerError)
		return
	}

	response := QueueStatsResponse{
		Success: true,
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHealthStatus handles getting system health status
func (bh *BroadcastHandlers) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	status := bh.broadcastService.GetHealthStatus()

	response := StatusResponse{
		Success: true,
		Status:  status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelMessage handles canceling a message
func (bh *BroadcastHandlers) CancelMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["messageId"]

	if messageID == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	err := bh.broadcastService.CancelMessage(messageID)
	if err != nil {
		logrus.Errorf("Failed to cancel message: %v", err)
		http.Error(w, "Failed to cancel message", http.StatusInternalServerError)
		return
	}

	response := MessageResponse{
		Success: true,
		Message: "Message cancelled successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelCampaignMessages handles canceling all messages for a campaign
func (bh *BroadcastHandlers) CancelCampaignMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]

	if campaignID == "" {
		http.Error(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}

	count, err := bh.broadcastService.CancelCampaignMessages(campaignID)
	if err != nil {
		logrus.Errorf("Failed to cancel campaign messages: %v", err)
		http.Error(w, "Failed to cancel campaign messages", http.StatusInternalServerError)
		return
	}

	response := BulkMessageResponse{
		Success: true,
		Message: "Campaign messages cancelled successfully",
		Count:   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PauseDevice handles pausing a device
func (bh *BroadcastHandlers) PauseDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	err := bh.broadcastService.PauseDevice(deviceID)
	if err != nil {
		logrus.Errorf("Failed to pause device: %v", err)
		http.Error(w, "Failed to pause device", http.StatusInternalServerError)
		return
	}

	response := MessageResponse{
		Success: true,
		Message: "Device paused successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResumeDevice handles resuming a device
func (bh *BroadcastHandlers) ResumeDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	err := bh.broadcastService.ResumeDevice(deviceID)
	if err != nil {
		logrus.Errorf("Failed to resume device: %v", err)
		http.Error(w, "Failed to resume device", http.StatusInternalServerError)
		return
	}

	response := MessageResponse{
		Success: true,
		Message: "Device resumed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ForceCleanup handles manual queue cleanup
func (bh *BroadcastHandlers) ForceCleanup(w http.ResponseWriter, r *http.Request) {
	bh.broadcastService.ForceCleanup()

	response := MessageResponse{
		Success: true,
		Message: "Queue cleanup triggered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPendingMessageCount handles getting pending message count for a device
func (bh *BroadcastHandlers) GetPendingMessageCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	count, err := bh.broadcastService.GetPendingMessageCount(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get pending message count: %v", err)
		http.Error(w, "Failed to get pending message count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"count":   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all broadcast routes
func (bh *BroadcastHandlers) RegisterRoutes(router *mux.Router) {
	// Message queuing routes
	router.HandleFunc("/api/broadcast/campaign/queue", bh.QueueCampaignMessage).Methods("POST")
	router.HandleFunc("/api/broadcast/sequence/queue", bh.QueueSequenceMessage).Methods("POST")
	router.HandleFunc("/api/broadcast/bulk/queue", bh.QueueBulkMessages).Methods("POST")

	// Message status and management routes
	router.HandleFunc("/api/broadcast/message/{messageId}/status", bh.GetMessageStatus).Methods("GET")
	router.HandleFunc("/api/broadcast/message/{messageId}/cancel", bh.CancelMessage).Methods("POST")
	router.HandleFunc("/api/broadcast/campaign/{campaignId}/cancel", bh.CancelCampaignMessages).Methods("POST")

	// Device management routes
	router.HandleFunc("/api/broadcast/device/{deviceId}/pause", bh.PauseDevice).Methods("POST")
	router.HandleFunc("/api/broadcast/device/{deviceId}/resume", bh.ResumeDevice).Methods("POST")
	router.HandleFunc("/api/broadcast/device/{deviceId}/pending-count", bh.GetPendingMessageCount).Methods("GET")

	// System monitoring routes
	router.HandleFunc("/api/broadcast/metrics", bh.GetMetrics).Methods("GET")
	router.HandleFunc("/api/broadcast/queue/stats", bh.GetQueueStats).Methods("GET")
	router.HandleFunc("/api/broadcast/health", bh.GetHealthStatus).Methods("GET")
	router.HandleFunc("/api/broadcast/cleanup", bh.ForceCleanup).Methods("POST")
}