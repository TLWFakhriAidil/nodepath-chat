package whatsapp

import (
	"strings"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
)

// ErrorMonitor monitors for specific errors and triggers auto-reconnection
type ErrorMonitor struct {
	lastErrors map[string]time.Time
	reconnecting map[string]bool
	mu sync.Mutex
}

var errorMonitor *ErrorMonitor

// InitErrorMonitor initializes the error monitoring system
func InitErrorMonitor() {
	errorMonitor = &ErrorMonitor{
		lastErrors: make(map[string]time.Time),
		reconnecting: make(map[string]bool),
	}
	
	// Set up logrus hook to monitor all error logs
	logrus.AddHook(&errorHook{monitor: errorMonitor})
	
	logrus.Info("Error monitor initialized - will auto-reconnect devices on connection errors")
}

// errorHook implements logrus.Hook
type errorHook struct {
	monitor *ErrorMonitor
}

func (hook *errorHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}

func (hook *errorHook) Fire(entry *logrus.Entry) error {
	message := entry.Message
	
	// Check for device connection error
	if strings.Contains(message, "Failed to get device connection: no device connection found for device") ||
	   strings.Contains(message, "device not connected") ||
	   strings.Contains(message, "connection closed") {
		
		// Extract device ID
		var deviceID string
		
		// Try to extract device ID from error message
		if strings.Contains(message, "device ") {
			parts := strings.Split(message, "device ")
			if len(parts) >= 2 {
				// Get first word after "device "
				deviceID = strings.Fields(parts[1])[0]
				// Clean up device ID
				deviceID = strings.Trim(deviceID, "\"',:.()")
			}
		}
		
		if deviceID != "" && len(deviceID) == 36 { // UUID length check
			hook.monitor.mu.Lock()
			
			// Check if already reconnecting
			if hook.monitor.reconnecting[deviceID] {
				hook.monitor.mu.Unlock()
				return nil
			}
			
			// Check if we already tried recently (within 1 minute)
			if lastTime, exists := hook.monitor.lastErrors[deviceID]; exists {
				if time.Since(lastTime) < 1*time.Minute {
					hook.monitor.mu.Unlock()
					return nil
				}
			}
			
			// Mark as reconnecting
			hook.monitor.reconnecting[deviceID] = true
			hook.monitor.lastErrors[deviceID] = time.Now()
			hook.monitor.mu.Unlock()
			
			// Trigger reconnection in background
			go func() {
				defer func() {
					hook.monitor.mu.Lock()
					delete(hook.monitor.reconnecting, deviceID)
					hook.monitor.mu.Unlock()
				}()
				
				logrus.Infof("🔄 Error monitor: Auto-reconnecting device %s due to connection error", deviceID)
				
				// Use the working reconnection logic
				err := RefreshDeviceConnectionByID(deviceID)
				if err != nil {
					logrus.Errorf("Failed to reconnect device %s: %v", deviceID, err)
				}
			}()
			
			// Clean up old entries (older than 1 hour)
			hook.monitor.mu.Lock()
			for id, timestamp := range hook.monitor.lastErrors {
				if time.Since(timestamp) > time.Hour {
					delete(hook.monitor.lastErrors, id)
				}
			}
			hook.monitor.mu.Unlock()
		}
	}
	
	return nil
}

// MonitorDeviceErrors starts monitoring for device connection errors
func MonitorDeviceErrors() {
	logrus.Info("Error monitor DISABLED - no auto reconnect")
	return // DISABLED - No auto reconnect
	/*
	if errorMonitor == nil {
		InitErrorMonitor()
	}
	
	// Also do an initial check for all devices that need reconnection
	go func() {
		// Wait a bit for services to initialize
		time.Sleep(10 * time.Second)
		
		logrus.Info("Error monitor: Performing initial device status check...")
		
		// This will reconnect all devices that were online before
		AutoReconnectOnlineDevices()
	}()
	*/
}
