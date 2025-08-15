package services

import (
	"context"
	"fmt"
	"sync"
	"time"

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
func (s *aiCronService) ProcessPendingResponses() error {
	logrus.Debug("Processing pending AI responses")

	// For now, we'll skip pending response processing as the method doesn't exist
	// TODO: Implement GetConversationsUpdatedBefore method in repository
	logrus.Debug("Pending response processing skipped - method not implemented")
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

	return nil
}