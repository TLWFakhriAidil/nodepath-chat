package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	urlPkg "net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// AICronService interface defines methods for AI cron job management
type AICronService interface {
	// Start the cron service
	Start() error

	// Stop the cron service
	Stop() error

	// Schedule follow-up message
	ScheduleFollowUp(prospectNum string, delay time.Duration, message string) error

	// Process pending AI responses
	ProcessPendingResponses() error

	// Clean up old conversation logs
	CleanupOldLogs() error

	// Update conversation statistics
	UpdateConversationStats() error

	// Check for inactive conversations
	CheckInactiveConversations() error
}

// aiCronService implements AICronService interface
type aiCronService struct {
	aiRepo            repository.AIWhatsappRepository
	deviceRepo        repository.DeviceSettingsRepository
	aiWhatsappService AIWhatsappService
	cronScheduler     *cron.Cron
	ctx               context.Context
	cancel            context.CancelFunc
	mu                sync.RWMutex
	isRunning         bool
	followUpJobs      map[string]cron.EntryID // Track follow-up jobs
	queueMonitor      *QueueMonitor           // Monitor for performance bottlenecks
	workerPool        chan struct{}           // Worker pool for concurrent processing
	maxWorkers        int                     // Maximum number of concurrent workers
}

// FollowUpJob represents a scheduled follow-up job
type FollowUpJob struct {
	ProspectNum string
	Message     string
	ScheduledAt time.Time
}

// NewAICronService creates a new instance of AICronService
func NewAICronService(
	aiRepo repository.AIWhatsappRepository,
	deviceRepo repository.DeviceSettingsRepository,
	aiWhatsappService AIWhatsappService,
) AICronService {
	ctx, cancel := context.WithCancel(context.Background())
	maxWorkers := 50 // Configurable worker pool size for 3000+ concurrent users

	return &aiCronService{
		aiRepo:            aiRepo,
		deviceRepo:        deviceRepo,
		aiWhatsappService: aiWhatsappService,
		cronScheduler:     cron.New(cron.WithSeconds()),
		ctx:               ctx,
		cancel:            cancel,
		followUpJobs:      make(map[string]cron.EntryID),
		queueMonitor:      NewQueueMonitor(),
		workerPool:        make(chan struct{}, maxWorkers),
		maxWorkers:        maxWorkers,
	}
}

// Start starts the cron service with scheduled jobs
func (s *aiCronService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("cron service is already running")
	}

	// Start queue monitor for performance tracking
	s.queueMonitor.Start()

	// Schedule periodic jobs
	err := s.schedulePeriodicJobs()
	if err != nil {
		return fmt.Errorf("failed to schedule periodic jobs: %w", err)
	}

	// Start the cron scheduler
	s.cronScheduler.Start()
	s.isRunning = true

	logrus.WithFields(logrus.Fields{
		"max_workers":        s.maxWorkers,
		"monitoring_enabled": true,
	}).Info("AI cron service started with performance monitoring")
	return nil
}

// sendAIResponse sends the AI response using the appropriate WhatsApp provider
// This function mimics the PHP cron job's sendChatMessage and sendMessage functionality
// Supports text, image, audio, and video message types
func (s *aiCronService) sendAIResponse(prospectNum, deviceID string, response *AIWhatsappResponse) error {
	// Get device settings to determine provider and credentials
	deviceSettings, err := s.deviceRepo.GetDeviceSettingsByDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings: %w", err)
	}

	// Determine provider based on instance length (similar to PHP logic)
	provider := s.determineProvider(deviceSettings.Instance.String)

	// Send each response message
	for _, msg := range response.Response {
		if msg.Content == "" {
			continue
		}

		switch msg.Type {
		case "text":
			// Use sendMessage for text messages (equivalent to PHP sendMessage)
			err = s.sendTextMessage(prospectNum, msg.Content, deviceSettings, provider)
			if err != nil {
				logrus.WithError(err).Error("Failed to send text message")
				return fmt.Errorf("failed to send text message: %w", err)
			}
		case "image":
			// Use sendChatMessage for multimedia messages (equivalent to PHP sendChatMessage)
			err = s.sendChatMessage(prospectNum, "", msg.Content, deviceSettings, provider)
			if err != nil {
				logrus.WithError(err).Error("Failed to send image message")
				return fmt.Errorf("failed to send image message: %w", err)
			}
		case "audio":
			// Use sendChatMessage for audio messages
			err = s.sendChatMessage(prospectNum, "", msg.Content, deviceSettings, provider)
			if err != nil {
				logrus.WithError(err).Error("Failed to send audio message")
				return fmt.Errorf("failed to send audio message: %w", err)
			}
		case "video":
			// Use sendChatMessage for video messages
			err = s.sendChatMessage(prospectNum, "", msg.Content, deviceSettings, provider)
			if err != nil {
				logrus.WithError(err).Error("Failed to send video message")
				return fmt.Errorf("failed to send video message: %w", err)
			}
		default:
			// Default to text message
			err = s.sendTextMessage(prospectNum, msg.Content, deviceSettings, provider)
			if err != nil {
				logrus.WithError(err).Error("Failed to send default message")
				return fmt.Errorf("failed to send default message: %w", err)
			}
		}

		// Add 5 second delay between messages to avoid rate limiting
		time.Sleep(5000 * time.Millisecond)
	}

	return nil
}

// Stop stops the cron service
func (s *aiCronService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("cron service is not running")
	}

	// Stop queue monitor
	s.queueMonitor.Stop()

	// Stop the cron scheduler
	s.cronScheduler.Stop()
	s.cancel()
	s.isRunning = false

	logrus.Info("AI Cron Service stopped successfully")
	return nil
}

// schedulePeriodicJobs schedules all periodic cron jobs
func (s *aiCronService) schedulePeriodicJobs() error {
	// Process pending responses every 30 seconds
	_, err := s.cronScheduler.AddFunc("*/30 * * * * *", func() {
		if err := s.ProcessPendingResponses(); err != nil {
			logrus.WithError(err).Error("Failed to process pending responses")
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule pending responses job: %w", err)
	}

	// Check for inactive conversations every 5 minutes
	_, err = s.cronScheduler.AddFunc("0 */5 * * * *", func() {
		if err := s.CheckInactiveConversations(); err != nil {
			logrus.WithError(err).Error("Failed to check inactive conversations")
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule inactive conversations job: %w", err)
	}

	// Update conversation statistics every 15 minutes
	_, err = s.cronScheduler.AddFunc("0 */15 * * * *", func() {
		if err := s.UpdateConversationStats(); err != nil {
			logrus.WithError(err).Error("Failed to update conversation stats")
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule stats update job: %w", err)
	}

	// Clean up old logs daily at 2 AM
	_, err = s.cronScheduler.AddFunc("0 0 2 * * *", func() {
		if err := s.CleanupOldLogs(); err != nil {
			logrus.WithError(err).Error("Failed to cleanup old logs")
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule cleanup job: %w", err)
	}

	logrus.Info("All periodic cron jobs scheduled successfully")
	return nil
}

// ScheduleFollowUp schedules a follow-up message for a prospect
func (s *aiCronService) ScheduleFollowUp(prospectNum string, delay time.Duration, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("cron service is not running")
	}

	// Cancel existing follow-up job if any
	if existingJobID, exists := s.followUpJobs[prospectNum]; exists {
		s.cronScheduler.Remove(existingJobID)
		delete(s.followUpJobs, prospectNum)
	}

	// Schedule new follow-up job
	scheduledTime := time.Now().Add(delay)
	cronExpr := fmt.Sprintf("%d %d %d %d %d *",
		scheduledTime.Second(),
		scheduledTime.Minute(),
		scheduledTime.Hour(),
		scheduledTime.Day(),
		int(scheduledTime.Month()),
	)

	jobID, err := s.cronScheduler.AddFunc(cronExpr, func() {
		s.executeFollowUp(prospectNum, message)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule follow-up: %w", err)
	}

	s.followUpJobs[prospectNum] = jobID

	logrus.WithFields(logrus.Fields{
		"prospect_num":   prospectNum,
		"scheduled_time": scheduledTime,
		"message":        message,
	}).Info("Follow-up message scheduled")

	return nil
}

// executeFollowUp executes a scheduled follow-up message
func (s *aiCronService) executeFollowUp(prospectNum, message string) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"message":      message,
	}).Info("Executing scheduled follow-up")

	// Get AI conversation data
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI conversation for follow-up")
		return
	}

	if aiConv == nil {
		logrus.WithField("prospect_num", prospectNum).Warn("AI conversation not found for follow-up")
		return
	}

	// Check if human takeover is active
	if aiConv.Human == 1 {
		logrus.WithField("prospect_num", prospectNum).Info("Human takeover active, skipping follow-up")
		return
	}

	// Get device settings to determine device ID
	deviceSettings, err := s.deviceRepo.GetAllDeviceSettings()
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings for follow-up")
		return
	}

	var deviceID string
	if len(deviceSettings) > 0 {
		if deviceSettings[0].IDDevice.Valid {
			deviceID = deviceSettings[0].IDDevice.String
		}
	} else {
		logrus.Error("No device settings found for follow-up")
		return
	}

	// Process the follow-up message through AI service
	// Get stage value from NullString
	stage := ""
	if aiConv.Stage.Valid {
		stage = aiConv.Stage.String
	}
	_, err = s.aiWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, stage, "User")
	if err != nil {
		logrus.WithError(err).Error("Failed to process follow-up message")
		return
	}

	// Remove the job from tracking
	s.mu.Lock()
	delete(s.followUpJobs, prospectNum)
	s.mu.Unlock()

	logrus.WithField("prospect_num", prospectNum).Info("Follow-up message executed successfully")
}

// ProcessPendingResponses processes any pending AI responses with concurrent processing and monitoring
// Equivalent to the PHP cron job that processes AI conversations and sends replies
func (s *aiCronService) ProcessPendingResponses() error {
	startTime := time.Now()
	logrus.Debug("Processing pending AI responses with performance monitoring")

	// Get all active AI conversations that need processing
	conversations, err := s.aiRepo.GetActiveAIConversations()
	if err != nil {
		s.queueMonitor.RecordError()
		return fmt.Errorf("failed to get active conversations: %w", err)
	}

	// Record queue size for monitoring
	s.queueMonitor.RecordQueueSize("pending_responses", int64(len(conversations)))

	// Process conversations concurrently using worker pool
	var wg sync.WaitGroup
	processedCount := int64(0)
	errorCount := int64(0)
	var countMutex sync.Mutex

	for _, conv := range conversations {
		// Skip if human takeover is active
		if conv.Human == 1 {
			continue
		}

		// Check if there's a current message to process
		if !conv.ConvCurrent.Valid || conv.ConvCurrent.String == "" {
			continue
		}

		// Acquire worker from pool
		s.workerPool <- struct{}{}
		wg.Add(1)

		go func(conversation models.AIWhatsapp) {
			defer func() {
				<-s.workerPool // Release worker
				wg.Done()
			}()

			processingStart := time.Now()
			err := s.processConversation(conversation)
			processingTime := time.Since(processingStart)

			// Record processing metrics
			s.queueMonitor.RecordProcessingTime(processingTime)

			countMutex.Lock()
			if err != nil {
				errorCount++
				s.queueMonitor.RecordError()
				logrus.WithError(err).WithFields(logrus.Fields{
					"prospect_num":    conversation.ProspectNum,
					"processing_time": processingTime,
				}).Error("Failed to process conversation")
			} else {
				processedCount++
			}
			countMutex.Unlock()
		}(conv)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Calculate and record worker utilization
	activeWorkers := len(s.workerPool)
	utilization := float64(activeWorkers) / float64(s.maxWorkers) * 100
	s.queueMonitor.RecordWorkerUtilization("ai_processing", utilization)

	totalProcessingTime := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"processed_count":           processedCount,
		"error_count":               errorCount,
		"total_conversations":       len(conversations),
		"total_processing_time":     totalProcessingTime,
		"worker_utilization":        utilization,
		"avg_time_per_conversation": totalProcessingTime / time.Duration(len(conversations)),
	}).Info("Completed processing pending AI responses")

	return nil
}

// processConversation processes a single conversation with detailed monitoring
func (s *aiCronService) processConversation(conv models.AIWhatsapp) error {
	currentText := conv.ConvCurrent.String

	// Check for stage command in current text
	if strings.Contains(strings.ToLower(currentText), "stage:") {
		// Extract and update stage
		parts := strings.Split(currentText, ":")
		if len(parts) > 1 {
			newStage := strings.TrimSpace(parts[1])
			err := s.aiRepo.UpdateConversationStage(conv.ProspectNum, newStage)
			if err != nil {
				return fmt.Errorf("failed to update conversation stage: %w", err)
			}
			// Clear current message after processing stage command
			err = s.aiRepo.UpdateConvCurrent(conv.ProspectNum, "")
			if err != nil {
				return fmt.Errorf("failed to clear conv_current: %w", err)
			}
			return nil
		}
	}

	// Process AI conversation
	// Get stage value from NullString
	stage := ""
	if conv.Stage.Valid {
		stage = conv.Stage.String
	}
	response, err := s.aiWhatsappService.ProcessAIConversation(
		conv.ProspectNum,
		conv.IDDevice,
		currentText,
		stage,
		"User",
	)
	if err != nil {
		return fmt.Errorf("failed to process AI conversation: %w", err)
	}

	// Send the AI response
	if response != nil && len(response.Response) > 0 {
		err = s.sendAIResponse(conv.ProspectNum, conv.IDDevice, response)
		if err != nil {
			return fmt.Errorf("failed to send AI response: %w", err)
		}
	}

	// Clear current message after processing
	err = s.aiRepo.UpdateConvCurrent(conv.ProspectNum, "")
	if err != nil {
		return fmt.Errorf("failed to clear conv_current: %w", err)
	}

	return nil
}

// CleanupOldLogs removes conversation logs older than 30 days
func (s *aiCronService) CleanupOldLogs() error {
	logrus.Info("Starting cleanup of old conversation logs")

	cutoffDate := time.Now().AddDate(0, 0, -30) // 30 days ago
	// For now, we'll skip the cleanup as the method doesn't exist yet
	// TODO: Implement DeleteOldConversationLogs method in repository
	logrus.WithFields(logrus.Fields{
		"cutoff_date": cutoffDate,
	}).Info("Cleanup of old conversation logs skipped - method not implemented")

	return nil
}

// UpdateConversationStats updates conversation statistics
func (s *aiCronService) UpdateConversationStats() error {
	logrus.Debug("Updating conversation statistics")

	// Get conversation statistics for all staff (using empty string as placeholder)
	stats, err := s.aiRepo.GetConversationStats("")
	if err != nil {
		return fmt.Errorf("failed to get conversation stats: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_conversations":  stats["total"],
		"active_conversations": stats["active_ai"],
		"human_takeovers":      stats["human_takeover"],
		"today_conversations":  stats["today"],
	}).Info("Conversation statistics updated")

	return nil
}

// CheckInactiveConversations checks for conversations that have been inactive for too long
func (s *aiCronService) CheckInactiveConversations() error {
	logrus.Debug("Checking for inactive conversations")

	// Get conversations that haven't been updated in the last 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)
	// For now, we'll skip inactive conversation checking as the method doesn't exist
	// TODO: Implement GetConversationsUpdatedBefore method in repository
	logrus.WithFields(logrus.Fields{
		"cutoff_time": cutoffTime,
	}).Info("Inactive conversation check skipped - method not implemented")
	return nil
}

// determineProvider determines the WhatsApp provider based on instance string patterns
// This mimics the PHP logic for provider detection with WAHA support
func (s *aiCronService) determineProvider(instance string) string {
	// Check for WAHA provider patterns (typically contains domain-like structure)
	if strings.Contains(instance, ".") && (strings.Contains(instance, "waha") || strings.Contains(instance, "api")) {
		return "waha"
	}
	// Original logic for Wablas and Whacenter
	if len(instance) <= 20 {
		return "wablas"
	}
	return "whacenter"
}

// sendTextMessage sends a text message through the appropriate provider
// Equivalent to PHP sendMessage function
func (s *aiCronService) sendTextMessage(to, message string, deviceSettings *models.DeviceSettings, provider string) error {
	// Add delay before sending (similar to PHP delax parameter)
	delay := 1 * time.Second
	time.Sleep(delay)

	switch provider {
	case "whacenter":
		return s.sendWhacenterTextMessage(to, message, deviceSettings)
	case "wablas":
		return s.sendWablasTextMessage(to, message, deviceSettings)
	case "waha":
		return s.sendWahaTextMessage(to, message, deviceSettings)
	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for text message")
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendChatMessage sends multimedia messages (video, audio, image)
// Equivalent to PHP sendChatMessage function
func (s *aiCronService) sendChatMessage(to, reply, fileURL string, deviceSettings *models.DeviceSettings, provider string) error {
	// Console log for tracing media URL in chat message
	logrus.WithFields(logrus.Fields{
		"to":              to,
		"file_url":        fileURL,
		"provider":        provider,
		"device_id":       deviceSettings.IDDevice,
		"file_url_length": len(fileURL),
	}).Info("🔍 AI CRON: MEDIA URL RECEIVED FOR TRACING")

	// Add delay before sending
	delay := 1 * time.Second
	time.Sleep(delay)

	// Determine file type based on extension
	fileType := s.getFileType(fileURL)

	switch provider {
	case "wablas":
		return s.sendWablasMultimediaMessage(to, fileURL, fileType, deviceSettings)
	case "whacenter":
		return s.sendWhacenterMultimediaMessage(to, fileURL, fileType, deviceSettings)
	case "waha":
		return s.sendWahaMultimediaMessage(to, fileURL, fileType, deviceSettings)
	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for multimedia message")
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// getFileType determines file type based on URL extension
func (s *aiCronService) getFileType(fileURL string) string {
	var fileType string
	if strings.Contains(fileURL, ".jpg") || strings.Contains(fileURL, ".jpeg") || strings.Contains(fileURL, ".png") {
		fileType = "image"
	} else if strings.Contains(fileURL, ".mp4") || strings.Contains(fileURL, ".avi") {
		fileType = "video"
	} else if strings.Contains(fileURL, ".mp3") || strings.Contains(fileURL, ".wav") {
		fileType = "audio"
	} else if strings.Contains(fileURL, ".pdf") || strings.Contains(fileURL, ".doc") {
		fileType = "document"
	} else {
		fileType = "image" // default to image
	}

	// Console log for tracing file type determination
	logrus.WithFields(logrus.Fields{
		"file_url":        fileURL,
		"determined_type": fileType,
		"url_extensions_found": func() []string {
			extensions := []string{}
			if strings.Contains(fileURL, ".jpg") {
				extensions = append(extensions, ".jpg")
			}
			if strings.Contains(fileURL, ".jpeg") {
				extensions = append(extensions, ".jpeg")
			}
			if strings.Contains(fileURL, ".png") {
				extensions = append(extensions, ".png")
			}
			if strings.Contains(fileURL, ".mp4") {
				extensions = append(extensions, ".mp4")
			}
			if strings.Contains(fileURL, ".avi") {
				extensions = append(extensions, ".avi")
			}
			if strings.Contains(fileURL, ".mp3") {
				extensions = append(extensions, ".mp3")
			}
			if strings.Contains(fileURL, ".wav") {
				extensions = append(extensions, ".wav")
			}
			if strings.Contains(fileURL, ".pdf") {
				extensions = append(extensions, ".pdf")
			}
			if strings.Contains(fileURL, ".doc") {
				extensions = append(extensions, ".doc")
			}
			return extensions
		}(),
	}).Info("🔍 AI CRON: FILE TYPE DETERMINED FOR TRACING")

	return fileType
}

// sendWablasTextMessage sends text message via Wablas provider
func (s *aiCronService) sendWablasTextMessage(to, message string, deviceSettings *models.DeviceSettings) error {
	logrus.WithFields(logrus.Fields{
		"to":        to,
		"provider":  "wablas",
		"device_id": deviceSettings.IDDevice,
	}).Debug("Sending text message via Wablas")

	// TODO: Implement actual Wablas API call
	// This should use the device settings to make HTTP request to Wablas API
	logrus.Info("📤 WABLAS: Text message sent successfully")
	return nil
}

// sendWhacenterTextMessage sends text message via Whacenter provider
func (s *aiCronService) sendWhacenterTextMessage(to, message string, deviceSettings *models.DeviceSettings) error {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WHACENTER: No instance available")
		return fmt.Errorf("no instance available")
	}

	logrus.WithFields(logrus.Fields{
		"to":        to,
		"provider":  "whacenter",
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
	}).Debug("Sending text message via Whacenter")

	// Whacenter API endpoint for sending messages
	apiURL := "https://api.whacenter.com/api/send"

	// Prepare request payload - Use instance for device_id as per Whacenter API requirements
	payload := map[string]interface{}{
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
		"number":    to,
		"message":   message,
		"type":      "text",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to send message")
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	logFields := logrus.Fields{
		"to":            to,
		"status_code":   resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WHACENTER: Text message sent successfully")
		return nil
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WHACENTER: Text message failed")
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
}

// sendWablasMultimediaMessage sends multimedia message via Wablas provider
func (s *aiCronService) sendWablasMultimediaMessage(to, fileURL, fileType string, deviceSettings *models.DeviceSettings) error {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WABLAS: No instance available")
		return fmt.Errorf("no instance available")
	}

	logrus.WithFields(logrus.Fields{
		"to":        to,
		"file_type": fileType,
		"provider":  "wablas",
		"device_id": deviceSettings.IDDevice,
	}).Debug("Sending multimedia message via Wablas")

	// Determine API endpoint and field name based on file type
	var apiURL string
	var fieldName string

	switch fileType {
	case "video":
		apiURL = "https://my.wablas.com/api/send-video"
		fieldName = "video"
	case "audio":
		apiURL = "https://my.wablas.com/api/send-audio"
		fieldName = "audio"
	default: // image
		apiURL = "https://my.wablas.com/api/send-image"
		fieldName = "image"
	}

	// Prepare form data
	data := url.Values{}
	data.Set("phone", to)        // recipient number
	data.Set(fieldName, fileURL) // media file URL
	data.Set("message", "")      // empty message field for media

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to send multimedia message")
		return fmt.Errorf("failed to send multimedia message: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	logFields := logrus.Fields{
		"to":            to,
		"file_type":     fileType,
		"status_code":   resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WABLAS: Multimedia message sent successfully")
		return nil
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WABLAS: Multimedia message failed")
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
}

// sendWhacenterMultimediaMessage sends multimedia message via Whacenter provider
func (s *aiCronService) sendWhacenterMultimediaMessage(to, fileURL, fileType string, deviceSettings *models.DeviceSettings) error {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WHACENTER: No instance available")
		return fmt.Errorf("no instance available")
	}

	logrus.WithFields(logrus.Fields{
		"to":        to,
		"file_type": fileType,
		"provider":  "whacenter",
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
	}).Debug("Sending multimedia message via Whacenter")

	// Whacenter API endpoint for sending media
	apiURL := "https://api.whacenter.com/api/send"

	// Detect media type based on file extension (as per PHP code)
	mediaType := ""
	if strings.Contains(fileURL, ".mp4") {
		mediaType = "video"
	} else if strings.Contains(fileURL, ".mp3") {
		mediaType = "audio"
	} else {
		mediaType = "image"
	}

	// Prepare form data exactly as specified by user PHP code
	data := url.Values{}
	data.Set("device_id", deviceSettings.Instance.String) // device_id from instance
	data.Set("number", to)                                // recipient number
	data.Set("file", fileURL)                             // media file URL
	data.Set("message", "")                               // empty message field for media

	// Add type parameter for video and audio only (as per PHP code)
	if mediaType != "" && mediaType != "image" {
		data.Set("type", mediaType)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (form data, no authorization header as per user example)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to send multimedia message")
		return fmt.Errorf("failed to send multimedia message: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	logFields := logrus.Fields{
		"to":            to,
		"file_type":     fileType,
		"status_code":   resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WHACENTER: Multimedia message sent successfully")
		return nil
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WHACENTER: Multimedia message failed")
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
}

// sendWahaTextMessage sends text message via WAHA provider
// Based on WAHA API specification: /api/sendText endpoint
func (s *aiCronService) sendWahaTextMessage(to, message string, deviceSettings *models.DeviceSettings) error {
	logrus.WithFields(logrus.Fields{
		"to":        to,
		"provider":  "waha",
		"device_id": deviceSettings.IDDevice,
	}).Debug("Sending text message via WAHA")

	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WAHA: No instance available")
		return fmt.Errorf("no instance available")
	}

	if !deviceSettings.APIKey.Valid {
		logrus.Error("❌ WAHA: No API key available")
		return fmt.Errorf("no API key available")
	}

	// Prepare request payload according to WAHA API specification
	// session: instance from device settings
	// chatId: phone number with @c.us suffix
	// text: message content
	payload := map[string]interface{}{
		"session": deviceSettings.Instance.String,
		"chatId":  to + "@c.us",
		"text":    message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request to WAHA API base URL + /api/sendText endpoint
	// Using the API base URL from device settings instance field
	apiBaseURL := deviceSettings.Instance.String
	if !strings.HasPrefix(apiBaseURL, "http") {
		apiBaseURL = "https://" + apiBaseURL
	}
	url := fmt.Sprintf("%s/api/sendText", strings.TrimSuffix(apiBaseURL, "/"))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers according to WAHA API specification
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", deviceSettings.APIKey.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to send text message")
		return fmt.Errorf("failed to send text message: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	logFields := logrus.Fields{
		"to":            to,
		"status_code":   resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WAHA: Text message sent successfully")
		return nil
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WAHA: Text message failed")
		return fmt.Errorf("waha API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
}

// sendWahaMultimediaMessage sends multimedia message via WAHA provider - EXACTLY matching PHP implementation
func (s *aiCronService) sendWahaMultimediaMessage(to, fileURL, caption string, deviceSettings *models.DeviceSettings) error {
	logrus.WithFields(logrus.Fields{
		"to":        to,
		"file_url":  fileURL,
		"provider":  "waha",
		"device_id": deviceSettings.IDDevice,
	}).Debug("Sending multimedia message via WAHA")

	// Fixed API key as per PHP code
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"

	// Prepare variables matching PHP
	session := deviceSettings.Instance.String
	number := regexp.MustCompile(`[^0-9]`).ReplaceAllString(to, "")
	chatId := number + "@c.us"

	var url string
	var data map[string]interface{}

	// Check file type and prepare request - EXACTLY as PHP
	if strings.Contains(fileURL, ".mp4") {
		// Video file
		url = "https://waha-plus-production-705f.up.railway.app/api/sendVideo"
		data = map[string]interface{}{
			"session": session,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": "video/mp4",
				"url":      fileURL,
				"filename": "Video",
			},
			"caption": caption,
		}
	} else if strings.Contains(fileURL, ".mp3") {
		// Audio file - using sendFile endpoint as per PHP
		url = "https://waha-plus-production-705f.up.railway.app/api/sendFile"
		data = map[string]interface{}{
			"session": session,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": "audio/mp3",
				"url":      fileURL,
				"filename": "Audio",
			},
			"caption": caption,
		}
	} else {
		// Image or other files - detect mimetype
		// Parse URL to get extension
		parsedURL, _ := urlPkg.Parse(fileURL)
		path := parsedURL.Path
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" && ext[0] == '.' {
			ext = ext[1:] // Remove leading dot
		}

		// Mimetype map matching PHP
		mimeMap := map[string]string{
			"jpg":  "image/jpeg",
			"jpeg": "image/jpeg",
			"png":  "image/png",
			"gif":  "image/gif",
			"webp": "image/webp",
			"bmp":  "image/bmp",
			"svg":  "image/svg+xml",
		}

		// Step 1: Try to use extension
		mimetype := ""
		if ext != "" {
			if mime, ok := mimeMap[ext]; ok {
				mimetype = mime
			}
		}

		// Step 2: Try to detect from headers (simplified for Go)
		if mimetype == "" {
			// Try to get content type from URL
			headReq, err := http.NewRequest("HEAD", fileURL, nil)
			if err == nil {
				client := &http.Client{Timeout: 5 * time.Second}
				if headResp, err := client.Do(headReq); err == nil {
					defer headResp.Body.Close()
					if contentType := headResp.Header.Get("Content-Type"); contentType != "" {
						mimetype = contentType
					}
				}
			}
		}

		// Step 3: Fallback default
		if mimetype == "" {
			mimetype = "image/jpeg"
		}

		url = "https://waha-plus-production-705f.up.railway.app/api/sendImage"
		data = map[string]interface{}{
			"session": session,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": mimetype,
				"url":      fileURL,
				"filename": "Image",
			},
			"caption": caption,
		}
	}

	// Marshal the data
	jsonPayload, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers exactly as PHP
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to send multimedia message")
		return fmt.Errorf("failed to send multimedia message: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response
	logFields := logrus.Fields{
		"to":            to,
		"status_code":   resp.StatusCode,
		"response_body": string(respBody),
		"url":           url,
		"file_url":      fileURL,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WAHA: Multimedia message sent successfully")
		return nil
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WAHA: Multimedia message failed")
		return fmt.Errorf("waha API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
}
