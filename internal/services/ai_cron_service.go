package services

import (
	"context"
	"fmt"
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
		deviceID = deviceSettings[0].IDDevice // Use first available device
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
func (s *aiCronService) ProcessPendingResponses() error {
	logrus.Debug("Processing pending AI responses")

	// Get conversations that haven't been updated in the last 5 minutes
	// and don't have human takeover active
	cutoffTime := time.Now().Add(-5 * time.Minute)
	pendingConversations, err := s.aiRepo.GetConversationsUpdatedBefore(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get pending conversations: %w", err)
	}

	processedCount := 0
	for _, conv := range pendingConversations {
		if conv.Human == 1 {
			continue // Skip conversations with human takeover
		}

		// Check if there are recent user messages without bot responses
		recentLogs, err := s.aiRepo.GetConversationHistory(conv.ProspectNum, 5)
		if err != nil {
			logrus.WithError(err).Error("Failed to get recent conversation history")
			continue
		}

		// Check if the last message was from user and needs a response
		if len(recentLogs) > 0 && recentLogs[0].Sender == "user" {
			// Check if there's already a bot response after this user message
			needsResponse := true
			for i := 1; i < len(recentLogs); i++ {
				if recentLogs[i].Sender == "bot" && recentLogs[i].CreatedAt.After(recentLogs[0].CreatedAt) {
					needsResponse = false
					break
				}
			}

			if needsResponse {
				// Process the pending message
				deviceSettings, err := s.deviceRepo.GetAllDeviceSettings()
				if err != nil || len(deviceSettings) == 0 {
					logrus.WithError(err).Error("No device settings available for pending response")
					continue
				}

				deviceID := deviceSettings[0].IDDevice
				_, err = s.aiWhatsappService.ProcessAIConversation(
					conv.ProspectNum,
					deviceID,
					recentLogs[0].Message,
					conv.Stage,
				)
				if err != nil {
					logrus.WithError(err).Error("Failed to process pending response")
					continue
				}

				processedCount++
			}
		}
	}

	if processedCount > 0 {
		logrus.WithField("processed_count", processedCount).Info("Processed pending AI responses")
	}

	return nil
}

// CleanupOldLogs removes conversation logs older than 30 days
func (s *aiCronService) CleanupOldLogs() error {
	logrus.Info("Starting cleanup of old conversation logs")

	cutoffDate := time.Now().AddDate(0, 0, -30) // 30 days ago
	deletedCount, err := s.aiRepo.DeleteOldConversationLogs(cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"deleted_count": deletedCount,
		"cutoff_date":   cutoffDate,
	}).Info("Completed cleanup of old conversation logs")

	return nil
}

// UpdateConversationStats updates conversation statistics
func (s *aiCronService) UpdateConversationStats() error {
	logrus.Debug("Updating conversation statistics")

	// Get conversation statistics
	stats, err := s.aiRepo.GetConversationStats()
	if err != nil {
		return fmt.Errorf("failed to get conversation stats: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_conversations": stats.TotalConversations,
		"active_conversations": stats.ActiveConversations,
		"human_takeovers":      stats.HumanTakeovers,
		"total_messages":       stats.TotalMessages,
	}).Info("Conversation statistics updated")

	return nil
}

// CheckInactiveConversations checks for conversations that have been inactive for too long
func (s *aiCronService) CheckInactiveConversations() error {
	logrus.Debug("Checking for inactive conversations")

	// Get conversations that haven't been updated in the last 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)
	inactiveConversations, err := s.aiRepo.GetConversationsUpdatedBefore(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get inactive conversations: %w", err)
	}

	for _, conv := range inactiveConversations {
		if conv.Human == 1 {
			continue // Skip conversations with human takeover
		}

		// Schedule a follow-up message for inactive conversations
		followUpMessage := "Hi! I noticed we haven't heard from you in a while. Is there anything I can help you with?"
		err = s.ScheduleFollowUp(conv.ProspectNum, 1*time.Minute, followUpMessage)
		if err != nil {
			logrus.WithError(err).WithField("prospect_num", conv.ProspectNum).Error("Failed to schedule follow-up for inactive conversation")
			continue
		}

		logrus.WithField("prospect_num", conv.ProspectNum).Info("Scheduled follow-up for inactive conversation")
	}

	if len(inactiveConversations) > 0 {
		logrus.WithField("inactive_count", len(inactiveConversations)).Info("Processed inactive conversations")
	}

	return nil
}