package usecase

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// Global sequence trigger processor instance
var sequenceTriggerProcessor *SequenceTriggerProcessor

// StartSequenceTriggerProcessor initializes and starts the sequence trigger processor
func StartSequenceTriggerProcessor() {
	logrus.Info("Initializing Direct Broadcast sequence processor...")
	
	// Get database connection
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// Create processor
	sequenceTriggerProcessor = NewSequenceTriggerProcessor(db)
	
	// Start processing
	go sequenceTriggerProcessor.StartProcessing()
	
	logrus.Info("Direct Broadcast sequence processor started successfully")
}

// StopSequenceTriggerProcessor stops the sequence trigger processor
func StopSequenceTriggerProcessor() {
	// Processing runs in a goroutine, no specific stop method needed
	logrus.Info("Sequence trigger processor stopped")
}

// GetSequenceTriggerProcessor returns the global instance
func GetSequenceTriggerProcessor() *SequenceTriggerProcessor {
	return sequenceTriggerProcessor
}
