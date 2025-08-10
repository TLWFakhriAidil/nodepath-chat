package whatsapp

import (
	"context"
	"math/rand"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// KeepaliveManager manages keepalive for devices
type KeepaliveManager struct {
	mu              sync.RWMutex
	activeKeepalives map[string]chan struct{}
	lastActivity    map[string]time.Time
	ctx             context.Context
	cancel          context.CancelFunc
}

var (
	keepaliveManager *KeepaliveManager
	keepaliveOnce    sync.Once
)

// GetKeepaliveManager returns singleton instance
func GetKeepaliveManager() *KeepaliveManager {
	keepaliveOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		keepaliveManager = &KeepaliveManager{
			activeKeepalives: make(map[string]chan struct{}),
			lastActivity:    make(map[string]time.Time),
			ctx:             ctx,
			cancel:          cancel,
		}
	})
	return keepaliveManager
}

// StartKeepalive starts keepalive for a device (only for null platform devices)
func (km *KeepaliveManager) StartKeepalive(deviceID string, client *whatsmeow.Client) {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	// Check if keepalive already exists and is the same client
	if _, exists := km.activeKeepalives[deviceID]; exists {
		logrus.Debugf("Keepalive already active for device %s, skipping duplicate", deviceID)
		return
	}
	
	// Check if this is a platform device
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Warnf("Failed to get device %s: %v", deviceID, err)
		return
	}
	
	// Only start keepalive for null platform devices
	if device.Platform != "" {
		logrus.Infof("Skipping keepalive for platform device %s (platform: %s)", deviceID, device.Platform)
		return
	}
	
	// Create stop channel
	stopChan := make(chan struct{})
	km.activeKeepalives[deviceID] = stopChan
	km.lastActivity[deviceID] = time.Now()
	
	// Start keepalive goroutine
	go km.runKeepalive(deviceID, client, stopChan)
	
	logrus.Infof("Started keepalive for device %s (null platform)", deviceID)
}

// StopKeepalive stops keepalive for a device
func (km *KeepaliveManager) StopKeepalive(deviceID string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	if stopChan, exists := km.activeKeepalives[deviceID]; exists {
		close(stopChan)
		delete(km.activeKeepalives, deviceID)
		delete(km.lastActivity, deviceID)
		logrus.Infof("Stopped keepalive for device %s", deviceID)
	}
}

// UpdateActivity updates last activity time (call when sending messages)
func (km *KeepaliveManager) UpdateActivity(deviceID string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	km.lastActivity[deviceID] = time.Now()
}

// runKeepalive runs the keepalive routine
func (km *KeepaliveManager) runKeepalive(deviceID string, client *whatsmeow.Client, stop chan struct{}) {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Keepalive parameters
	minInterval := 3 * time.Minute // Reduced frequency for 3000 devices
	maxInterval := 5 * time.Minute // Reduced frequency for 3000 devices
	activityThreshold := 10 * time.Minute // Less aggressive keepalive
	
	for {
		// Calculate random interval
		interval := minInterval + time.Duration(rand.Int63n(int64(maxInterval-minInterval)))
		
		select {
		case <-stop:
			return
		case <-km.ctx.Done():
			return
		case <-time.After(interval):
			// Check if still connected
			if !client.IsConnected() || !client.IsLoggedIn() {
				logrus.Debugf("Device %s not connected/logged in, skipping keepalive", deviceID)
				continue
			}
			
			// Check last activity
			km.mu.RLock()
			lastActivity, exists := km.lastActivity[deviceID]
			km.mu.RUnlock()
			
			if !exists {
				continue
			}
			
			// Skip if recent activity
			if time.Since(lastActivity) < activityThreshold {
				logrus.Debugf("Recent activity on device %s, skipping keepalive", deviceID)
				continue
			}
			
			// Perform keepalive action
			km.performKeepalive(deviceID, client)
		}
	}
}

// performKeepalive performs the actual keepalive action
func (km *KeepaliveManager) performKeepalive(deviceID string, client *whatsmeow.Client) {
	// Rotate between different methods
	method := time.Now().Unix() % 4
	
	switch method {
	case 0:
		// Method 1: Set presence available
		err := client.SendPresence(types.PresenceAvailable)
		if err != nil {
			logrus.Debugf("Keepalive presence failed for %s: %v", deviceID, err)
		} else {
			logrus.Debugf("Keepalive presence sent for %s", deviceID)
		}
		
	case 1:
		// Method 2: Set presence unavailable then available
		client.SendPresence(types.PresenceUnavailable)
		time.Sleep(1 * time.Second)
		err := client.SendPresence(types.PresenceAvailable)
		if err != nil {
			logrus.Debugf("Keepalive presence toggle failed for %s: %v", deviceID, err)
		} else {
			logrus.Debugf("Keepalive presence toggle sent for %s", deviceID)
		}
		
	case 2:
		// Method 3: Subscribe to own presence
		if client.Store.ID != nil {
			err := client.SubscribePresence(types.JID{
				User:   client.Store.ID.User,
				Server: types.DefaultUserServer,
			})
			if err != nil {
				logrus.Debugf("Keepalive presence subscription failed for %s: %v", deviceID, err)
			} else {
				logrus.Debugf("Keepalive presence subscription sent for %s", deviceID)
			}
		}
		
	case 3:
		// Method 4: Get own profile picture (lightweight)
		if client.Store.ID != nil {
			_, err := client.GetProfilePictureInfo(*client.Store.ID, &whatsmeow.GetProfilePictureParams{
				Preview: true,
			})
			if err != nil {
				logrus.Debugf("Keepalive profile pic request failed for %s: %v", deviceID, err)
			} else {
				logrus.Debugf("Keepalive profile pic request sent for %s", deviceID)
			}
		}
	}
}

// GetActiveKeepaliveCount returns number of active keepalives
func (km *KeepaliveManager) GetActiveKeepaliveCount() int {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return len(km.activeKeepalives)
}

// StopAll stops all keepalives
func (km *KeepaliveManager) StopAll() {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	for deviceID, stopChan := range km.activeKeepalives {
		close(stopChan)
		logrus.Infof("Stopped keepalive for device %s", deviceID)
	}
	
	km.activeKeepalives = make(map[string]chan struct{})
	km.lastActivity = make(map[string]time.Time)
	
	if km.cancel != nil {
		km.cancel()
	}
}
