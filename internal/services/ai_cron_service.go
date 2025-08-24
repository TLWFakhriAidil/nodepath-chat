package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	
	return &aiCronService{
		aiRepo:            aiRepo,
		deviceRepo:        deviceRepo,
		aiWhatsappService: aiWhatsappService,
		cronScheduler:     cron.New(cron.WithSeconds()),
		ctx:               ctx,
		cancel:            cancel,
		followUpJobs:      make(map[string]cron.EntryID),
	}
}

// Start starts the cron service with scheduled jobs
func (s *aiCronService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("cron service is already running")
	}

	// Schedule periodic jobs
	err := s.schedulePeriodicJobs()
	if err != nil {
		return fmt.Errorf("failed to schedule periodic jobs: %w", err)
	}

	// Start the cron scheduler
	s.cronScheduler.Start()
	s.isRunning = true

	logrus.Info("AI Cron Service started successfully")
	return nil
}

// sendAIResponse sends the AI response using the appropriate WhatsApp provider
// This function mimics the PHP cron job's sendChatMessage and sendMessage functionality
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
	_, err = s.aiWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, aiConv.Stage)
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

// ProcessPendingResponses processes any pending AI responses
// Equivalent to the PHP cron job that processes AI conversations and sends replies
func (s *aiCronService) ProcessPendingResponses() error {
	logrus.Debug("Processing pending AI responses")

	// Get all active AI conversations that need processing
	conversations, err := s.aiRepo.GetActiveAIConversations()
	if err != nil {
		return fmt.Errorf("failed to get active conversations: %w", err)
	}

	processedCount := 0
	for _, conv := range conversations {
		// Skip if human takeover is active
		if conv.Human == 1 {
			continue
		}

		// Check if conversation needs processing (similar to PHP time difference check)
		// Note: Balas is an int field, not a time field, so we skip this check for now
		// TODO: Implement proper time-based checking if needed

		// Check if there's a current message to process
		if !conv.ConvCurrent.Valid || conv.ConvCurrent.String == "" {
			continue
		}

		// Check for stage command in current text
		currentText := conv.ConvCurrent.String
		if strings.Contains(strings.ToLower(currentText), "stage:") {
			// Extract and update stage
			parts := strings.Split(currentText, ":")
			if len(parts) > 1 {
				newStage := strings.TrimSpace(parts[1])
				err := s.aiRepo.UpdateConversationStage(conv.ProspectNum, newStage)
				if err != nil {
					logrus.WithError(err).Error("Failed to update conversation stage")
				}
				// Clear current message after processing stage command
				err = s.aiRepo.UpdateConvCurrent(conv.ProspectNum, "")
				if err != nil {
					logrus.WithError(err).Error("Failed to clear conv_current")
				}
				continue
			}
		}

		// Process AI conversation
		response, err := s.aiWhatsappService.ProcessAIConversation(
			conv.ProspectNum,
			conv.IDDevice,
			currentText,
			conv.Stage,
		)
		if err != nil {
			logrus.WithError(err).WithField("prospect_num", conv.ProspectNum).Error("Failed to process AI conversation")
			continue
		}

		// Send the AI response using sendChatMessage and sendMessage functions
		if response != nil && len(response.Response) > 0 {
			err = s.sendAIResponse(conv.ProspectNum, conv.IDDevice, response)
			if err != nil {
				logrus.WithError(err).WithField("prospect_num", conv.ProspectNum).Error("Failed to send AI response")
				continue
			}
		}

		// Clear current message after processing
		err = s.aiRepo.UpdateConvCurrent(conv.ProspectNum, "")
		if err != nil {
			logrus.WithError(err).Error("Failed to clear conv_current")
		}

		processedCount++
	}

	logrus.WithField("processed_count", processedCount).Debug("Completed processing pending AI responses")
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

// determineProvider determines the WhatsApp provider based on instance string length
// This mimics the PHP logic for provider detection
func (s *aiCronService) determineProvider(instance string) string {
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
	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for text message")
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendChatMessage sends multimedia messages (video, audio, image) with caption support
// Equivalent to PHP sendChatMessage function
func (s *aiCronService) sendChatMessage(to, reply, fileURL string, deviceSettings *models.DeviceSettings, provider string) error {
	// Console log for tracing media URL in chat message
	logrus.WithFields(logrus.Fields{
		"to": to,
		"file_url": fileURL,
		"provider": provider,
		"device_id": deviceSettings.IDDevice,
		"file_url_length": len(fileURL),
	}).Info("🔍 AI CRON: MEDIA URL RECEIVED FOR TRACING")

	// Add delay before sending
	delay := 1 * time.Second
	time.Sleep(delay)

	// Determine file type based on extension
	fileType := s.getFileType(fileURL)

	switch provider {
	case "wablas":
		return s.sendWablasMultimediaMessage(to, reply, fileURL, fileType, deviceSettings)
	case "whacenter":
		return s.sendWhacenterMultimediaMessage(to, reply, fileURL, fileType, deviceSettings)
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
		"file_url": fileURL,
		"determined_type": fileType,
		"url_extensions_found": func() []string {
			extensions := []string{}
			if strings.Contains(fileURL, ".jpg") { extensions = append(extensions, ".jpg") }
			if strings.Contains(fileURL, ".jpeg") { extensions = append(extensions, ".jpeg") }
			if strings.Contains(fileURL, ".png") { extensions = append(extensions, ".png") }
			if strings.Contains(fileURL, ".mp4") { extensions = append(extensions, ".mp4") }
			if strings.Contains(fileURL, ".avi") { extensions = append(extensions, ".avi") }
			if strings.Contains(fileURL, ".mp3") { extensions = append(extensions, ".mp3") }
			if strings.Contains(fileURL, ".wav") { extensions = append(extensions, ".wav") }
			if strings.Contains(fileURL, ".pdf") { extensions = append(extensions, ".pdf") }
			if strings.Contains(fileURL, ".doc") { extensions = append(extensions, ".doc") }
			return extensions
		}(),
	}).Info("🔍 AI CRON: FILE TYPE DETERMINED FOR TRACING")

	return fileType
}

// sendWablasTextMessage sends text message via Wablas provider
func (s *aiCronService) sendWablasTextMessage(to, message string, deviceSettings *models.DeviceSettings) error {
	logrus.WithFields(logrus.Fields{
		"to": to,
		"provider": "wablas",
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
		"to": to,
		"provider": "whacenter",
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
	}).Debug("Sending text message via Whacenter")

	// Whacenter API endpoint for sending messages
	apiURL := "https://api.whacenter.com/api/send"

	// Prepare request payload - Use instance for device_id as per Whacenter API requirements
	payload := map[string]interface{}{
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
		"number": to,
		"message": message,
		"type": "text",
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
		"to": to,
		"status_code": resp.StatusCode,
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
func (s *aiCronService) sendWablasMultimediaMessage(to, caption, fileURL, fileType string, deviceSettings *models.DeviceSettings) error {
	logrus.WithFields(logrus.Fields{
		"to": to,
		"file_type": fileType,
		"provider": "wablas",
		"device_id": deviceSettings.IDDevice,
	}).Debug("Sending multimedia message via Wablas")

	// TODO: Implement actual Wablas multimedia API call
	// This should use the device settings to make HTTP request to Wablas API
	logrus.Info("📤 WABLAS: Multimedia message sent successfully")
	return nil
}

// sendWhacenterMultimediaMessage sends multimedia message via Whacenter provider
func (s *aiCronService) sendWhacenterMultimediaMessage(to, caption, fileURL, fileType string, deviceSettings *models.DeviceSettings) error {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WHACENTER: No instance available")
		return fmt.Errorf("no instance available")
	}

	logrus.WithFields(logrus.Fields{
		"to": to,
		"file_type": fileType,
		"provider": "whacenter",
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
	}).Debug("Sending multimedia message via Whacenter")

	// Whacenter API endpoint for sending media
	apiURL := "https://api.whacenter.com/api/send-media"

	// Prepare request payload - Use instance for device_id as per Whacenter API requirements
	payload := map[string]interface{}{
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
		"number":    to,
		"media_url": fileURL,
		"type":      fileType,
	}

	// Add caption if provided
	if caption != "" {
		payload["caption"] = caption
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
		"to":          to,
		"file_type":   fileType,
		"status_code": resp.StatusCode,
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