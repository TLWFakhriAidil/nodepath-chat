package broadcast

import (
	"context"
	"fmt"
	"sync"
	"time"

	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// BasicBroadcastManager manages broadcasting for all devices without Redis
type BasicBroadcastManager struct {
	workers      map[string]*DeviceWorker // deviceID -> worker
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	maxWorkers   int
}

// NewBasicBroadcastManager creates a new basic broadcast manager
func NewBasicBroadcastManager() *BasicBroadcastManager {
	ctx, cancel := context.WithCancel(context.Background())
	bm := &BasicBroadcastManager{
		workers:    make(map[string]*DeviceWorker),
		ctx:        ctx,
		cancel:     cancel,
		maxWorkers: 100, // Limit concurrent workers
	}
	
	logrus.Info("BasicBroadcastManager: Starting manager and queue processor")
	
	// Start manager
	go bm.Run()
	
	// Start queue processor
	go bm.ProcessQueue()
	
	return bm
}

// Run manages the broadcast manager
func (bm *BasicBroadcastManager) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			bm.StopAllWorkers()
			return
		case <-ticker.C:
			bm.CheckWorkerHealth()
		}
	}
}

// ProcessQueue processes messages from database queue
func (bm *BasicBroadcastManager) ProcessQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.processQueueBatch()
		}
	}
}

// processQueueBatch processes a batch of queued messages
func (bm *BasicBroadcastManager) processQueueBatch() {
	repo := repository.GetBroadcastRepository()
	
	// Get ALL pending messages, not just for active workers
	messages, err := repo.GetAllPendingMessages(50)
	if err != nil {
		logrus.Errorf("Failed to get pending messages: %v", err)
		return
	}
	
	if len(messages) > 0 {
		logrus.Infof("BasicBroadcastManager: Processing %d pending messages", len(messages))
	}
	
	// Process each message
	for _, msg := range messages {
		logrus.Infof("Processing message %s for device %s to %s", msg.ID, msg.DeviceID, msg.RecipientPhone)
		
		// Get or create worker for the device
		worker := bm.GetOrCreateWorker(msg.DeviceID)
		if worker != nil {
			worker.SendMessage(msg)
			logrus.Infof("Message %s queued to worker for device %s", msg.ID, msg.DeviceID)
		} else {
			logrus.Warnf("Could not create worker for device %s", msg.DeviceID)
		}
	}
}

// SendMessage sends a broadcast message
func (bm *BasicBroadcastManager) SendMessage(msg domainBroadcast.BroadcastMessage) error {
	worker := bm.GetOrCreateWorker(msg.DeviceID)
	if worker == nil {
		return fmt.Errorf("no worker available for device %s", msg.DeviceID)
	}
	
	select {
	case worker.messageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("worker queue full for device %s", msg.DeviceID)
	}
}

// GetOrCreateWorker gets or creates a worker for a device
func (bm *BasicBroadcastManager) GetOrCreateWorker(deviceID string) *DeviceWorker {
	bm.mu.RLock()
	worker, exists := bm.workers[deviceID]
	bm.mu.RUnlock()
	
	if exists && worker.IsHealthy() {
		return worker
	}
	
	// Create new worker
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Check if we've hit max workers
	if len(bm.workers) >= bm.maxWorkers {
		logrus.Warnf("Max workers reached (%d), cannot create worker for device %s", bm.maxWorkers, deviceID)
		return nil
	}
	
	// Get device
	clientManager := whatsapp.GetClientManager()
	client, err := clientManager.GetClient(deviceID)
	if err != nil || client == nil {
		logrus.Errorf("Failed to get device %s: %v", deviceID, err)
		return nil
	}
	
	// Create new worker
	worker = NewDeviceWorker(deviceID, client, 10, 30)
	bm.workers[deviceID] = worker
	go worker.Run()
	
	logrus.Infof("Created new worker for device %s", deviceID)
	return worker
}

// CheckWorkerHealth checks health of all workers
func (bm *BasicBroadcastManager) CheckWorkerHealth() {
	bm.mu.RLock()
	deviceIDs := make([]string, 0, len(bm.workers))
	for deviceID := range bm.workers {
		deviceIDs = append(deviceIDs, deviceID)
	}
	bm.mu.RUnlock()
	
	for _, deviceID := range deviceIDs {
		bm.mu.RLock()
		worker := bm.workers[deviceID]
		bm.mu.RUnlock()
		
		if worker != nil && !worker.IsHealthy() {
			logrus.Warnf("Worker for device %s is unhealthy, restarting", deviceID)
			
			// Stop old worker
			worker.Stop()
			
			// Remove from map
			bm.mu.Lock()
			delete(bm.workers, deviceID)
			bm.mu.Unlock()
			
			// Will be recreated on next message
		}
	}
}

// GetWorkerStatus gets status of a specific worker
func (bm *BasicBroadcastManager) GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool) {
	bm.mu.RLock()
	worker, exists := bm.workers[deviceID]
	bm.mu.RUnlock()
	
	if !exists || worker == nil {
		return domainBroadcast.WorkerStatus{}, false
	}
	
	return worker.GetStatus(), true
}

// GetAllWorkerStatus gets status of all workers
func (bm *BasicBroadcastManager) GetAllWorkerStatus() []domainBroadcast.WorkerStatus {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	statuses := make([]domainBroadcast.WorkerStatus, 0, len(bm.workers))
	for _, worker := range bm.workers {
		if worker != nil {
			statuses = append(statuses, worker.GetStatus())
		}
	}
	
	return statuses
}

// StopAllWorkers stops all active workers
func (bm *BasicBroadcastManager) StopAllWorkers() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	for deviceID, worker := range bm.workers {
		if worker != nil {
			worker.Stop()
			delete(bm.workers, deviceID)
		}
	}
	
	logrus.Info("All workers stopped")
	return nil
}

// ResumeFailedWorkers resumes all failed workers
func (bm *BasicBroadcastManager) ResumeFailedWorkers() error {
	// For basic manager, we just check health which will restart unhealthy workers
	bm.CheckWorkerHealth()
	return nil
}
// StopWorker stops a specific worker
func (bm *BasicBroadcastManager) StopWorker(deviceID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	worker, exists := bm.workers[deviceID]
	if !exists {
		return fmt.Errorf("worker not found for device %s", deviceID)
	}
	
	if worker != nil {
		worker.Stop()
		delete(bm.workers, deviceID)
		logrus.Infof("Worker for device %s stopped", deviceID)
	}
	
	return nil
}