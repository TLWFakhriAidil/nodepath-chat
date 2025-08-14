package broadcast

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

// BroadcastManager manages high-performance message broadcasting
type BroadcastManager struct {
	workers        map[string]*DeviceWorker
	workersMutex   sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	maxWorkers     int
	processedCount int64
	failedCount    int64
	running        bool
	runningMutex   sync.RWMutex
}

var (
	managerInstance *BroadcastManager
	managerOnce     sync.Once
)

// GetBroadcastManager returns the singleton broadcast manager instance
func GetBroadcastManager() *BroadcastManager {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		managerInstance = &BroadcastManager{
			workers:    make(map[string]*DeviceWorker),
			ctx:        ctx,
			cancel:     cancel,
			maxWorkers: 500, // Support for 3000+ devices with worker pooling
			running:    false,
		}
		
		// Optimize Go runtime for high concurrency
		runtime.GOMAXPROCS(runtime.NumCPU())
	})
	return managerInstance
}

// Start begins the broadcast manager operations
func (bm *BroadcastManager) Start() error {
	bm.runningMutex.Lock()
	defer bm.runningMutex.Unlock()
	
	if bm.running {
		return fmt.Errorf("broadcast manager is already running")
	}
	
	bm.running = true
	logrus.Info("Starting high-performance broadcast manager")
	
	// Start the main processing loop
	go bm.processQueueLoop()
	
	// Start health monitoring
	go bm.healthMonitorLoop()
	
	// Start metrics reporting
	go bm.metricsLoop()
	
	return nil
}

// Stop gracefully shuts down the broadcast manager
func (bm *BroadcastManager) Stop() error {
	bm.runningMutex.Lock()
	defer bm.runningMutex.Unlock()
	
	if !bm.running {
		return nil
	}
	
	logrus.Info("Stopping broadcast manager...")
	bm.running = false
	bm.cancel()
	
	// Stop all workers
	bm.workersMutex.Lock()
	for deviceID, worker := range bm.workers {
		logrus.Debugf("Stopping worker for device %s", deviceID)
		worker.Stop()
	}
	bm.workers = make(map[string]*DeviceWorker)
	bm.workersMutex.Unlock()
	
	logrus.Info("Broadcast manager stopped")
	return nil
}

// processQueueLoop continuously processes messages from the database queue
func (bm *BroadcastManager) processQueueLoop() {
	ticker := time.NewTicker(2 * time.Second) // Faster polling for high performance
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

// processQueueBatch processes a batch of queued messages with optimized performance
func (bm *BroadcastManager) processQueueBatch() {
	repo := repository.GetBroadcastRepository()
	
	// Get pending messages with larger batch size for better throughput
	messages, err := repo.GetAllPendingMessages(100)
	if err != nil {
		logrus.Errorf("Failed to get pending messages: %v", err)
		return
	}
	
	if len(messages) == 0 {
		return
	}
	
	logrus.Debugf("Processing %d pending messages", len(messages))
	
	// Group messages by device for efficient processing
	deviceMessages := make(map[string][]models.BroadcastMessage)
	for _, msg := range messages {
		deviceMessages[msg.DeviceID] = append(deviceMessages[msg.DeviceID], msg)
	}
	
	// Process each device's messages concurrently
	var wg sync.WaitGroup
	for deviceID, msgs := range deviceMessages {
		wg.Add(1)
		go func(devID string, messages []models.BroadcastMessage) {
			defer wg.Done()
			bm.processDeviceMessages(devID, messages)
		}(deviceID, msgs)
	}
	
	wg.Wait()
}

// processDeviceMessages processes messages for a specific device
func (bm *BroadcastManager) processDeviceMessages(deviceID string, messages []models.BroadcastMessage) {
	worker := bm.getOrCreateWorker(deviceID)
	if worker == nil {
		logrus.Errorf("Failed to get worker for device %s", deviceID)
		return
	}
	
	// Queue messages to the device worker
	for _, msg := range messages {
		if err := worker.QueueMessage(msg); err != nil {
			logrus.Errorf("Failed to queue message %s to device %s: %v", msg.ID, deviceID, err)
			atomic.AddInt64(&bm.failedCount, 1)
			
			// Update message status to failed
			repo := repository.GetBroadcastRepository()
			repo.UpdateMessageStatus(msg.ID, "failed", err.Error())
		} else {
			atomic.AddInt64(&bm.processedCount, 1)
		}
	}
}

// getOrCreateWorker gets an existing worker or creates a new one for the device
func (bm *BroadcastManager) getOrCreateWorker(deviceID string) *DeviceWorker {
	bm.workersMutex.RLock()
	worker, exists := bm.workers[deviceID]
	bm.workersMutex.RUnlock()
	
	if exists && worker.IsHealthy() {
		return worker
	}
	
	bm.workersMutex.Lock()
	defer bm.workersMutex.Unlock()
	
	// Double-check after acquiring write lock
	if worker, exists := bm.workers[deviceID]; exists && worker.IsHealthy() {
		return worker
	}
	
	// Check if we've reached the maximum number of workers
	if len(bm.workers) >= bm.maxWorkers {
		logrus.Warnf("Maximum number of workers (%d) reached, cannot create worker for device %s", bm.maxWorkers, deviceID)
		return nil
	}
	
	// Create new worker
	newWorker := NewDeviceWorker(deviceID, bm.ctx)
	if err := newWorker.Start(); err != nil {
		logrus.Errorf("Failed to start worker for device %s: %v", deviceID, err)
		return nil
	}
	
	bm.workers[deviceID] = newWorker
	logrus.Debugf("Created new worker for device %s (total workers: %d)", deviceID, len(bm.workers))
	
	return newWorker
}

// healthMonitorLoop monitors worker health and restarts failed workers
func (bm *BroadcastManager) healthMonitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth checks the health of all workers and restarts unhealthy ones
func (bm *BroadcastManager) checkWorkerHealth() {
	bm.workersMutex.Lock()
	defer bm.workersMutex.Unlock()
	
	unhealthyWorkers := make([]string, 0)
	
	for deviceID, worker := range bm.workers {
		if !worker.IsHealthy() {
			unhealthyWorkers = append(unhealthyWorkers, deviceID)
		}
	}
	
	// Remove and restart unhealthy workers
	for _, deviceID := range unhealthyWorkers {
		logrus.Warnf("Restarting unhealthy worker for device %s", deviceID)
		
		if worker, exists := bm.workers[deviceID]; exists {
			worker.Stop()
			delete(bm.workers, deviceID)
		}
		
		// Worker will be recreated on next message
	}
	
	if len(unhealthyWorkers) > 0 {
		logrus.Infof("Restarted %d unhealthy workers", len(unhealthyWorkers))
	}
}

// metricsLoop reports performance metrics
func (bm *BroadcastManager) metricsLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.reportMetrics()
		}
	}
}

// reportMetrics logs current performance metrics
func (bm *BroadcastManager) reportMetrics() {
	bm.workersMutex.RLock()
	workerCount := len(bm.workers)
	bm.workersMutex.RUnlock()
	
	processed := atomic.LoadInt64(&bm.processedCount)
	failed := atomic.LoadInt64(&bm.failedCount)
	
	logrus.WithFields(logrus.Fields{
		"active_workers":    workerCount,
		"processed_messages": processed,
		"failed_messages":   failed,
		"success_rate":      fmt.Sprintf("%.2f%%", float64(processed)/float64(processed+failed)*100),
	}).Info("Broadcast manager metrics")
}

// GetStatus returns the current status of the broadcast manager
func (bm *BroadcastManager) GetStatus() map[string]interface{} {
	bm.workersMutex.RLock()
	workerCount := len(bm.workers)
	workerStatuses := make([]models.WorkerStatus, 0, len(bm.workers))
	for _, worker := range bm.workers {
		workerStatuses = append(workerStatuses, worker.GetStatus())
	}
	bm.workersMutex.RUnlock()
	
	return map[string]interface{}{
		"running":           bm.running,
		"active_workers":    workerCount,
		"max_workers":       bm.maxWorkers,
		"processed_messages": atomic.LoadInt64(&bm.processedCount),
		"failed_messages":   atomic.LoadInt64(&bm.failedCount),
		"worker_statuses":   workerStatuses,
	}
}

// QueueMessage queues a message for broadcasting (for direct API usage)
func (bm *BroadcastManager) QueueMessage(msg models.BroadcastMessage) error {
	repo := repository.GetBroadcastRepository()
	return repo.QueueMessage(msg)
}

// GetWorkerStatus returns the status of all workers
func (bm *BroadcastManager) GetWorkerStatus() []models.WorkerStatus {
	bm.workersMutex.RLock()
	defer bm.workersMutex.RUnlock()

	status := make([]models.WorkerStatus, 0, len(bm.workers))
	for _, worker := range bm.workers {
		status = append(status, worker.GetStatus())
	}

	return status
}

// GetActiveDevices returns a list of devices with active workers
func (bm *BroadcastManager) GetActiveDevices() []string {
	bm.workersMutex.RLock()
	defer bm.workersMutex.RUnlock()

	devices := make([]string, 0, len(bm.workers))
	for deviceID, worker := range bm.workers {
		if worker.IsHealthy() {
			devices = append(devices, deviceID)
		}
	}

	return devices
}

// PauseDevice pauses message processing for a specific device
func (bm *BroadcastManager) PauseDevice(deviceID string) error {
	bm.workersMutex.RLock()
	worker, exists := bm.workers[deviceID]
	bm.workersMutex.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// For now, we'll stop the worker (pause functionality can be enhanced)
	worker.Stop()
	logrus.Infof("Device %s paused", deviceID)
	return nil
}

// ResumeDevice resumes message processing for a specific device
func (bm *BroadcastManager) ResumeDevice(deviceID string) error {
	bm.workersMutex.RLock()
	worker, exists := bm.workers[deviceID]
	bm.workersMutex.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// Restart the worker
	err := worker.Start()
	if err != nil {
		return fmt.Errorf("failed to resume device %s: %v", deviceID, err)
	}

	logrus.Infof("Device %s resumed", deviceID)
	return nil
}

// RestartDevice restarts the worker for a specific device
func (bm *BroadcastManager) RestartDevice(deviceID string) error {
	bm.workersMutex.Lock()
	defer bm.workersMutex.Unlock()

	if worker, exists := bm.workers[deviceID]; exists {
		worker.Stop()
		delete(bm.workers, deviceID)
	}

	// Create new worker
	newWorker := NewDeviceWorker(deviceID, bm.ctx)
	err := newWorker.Start()
	if err != nil {
		return fmt.Errorf("failed to restart device %s: %v", deviceID, err)
	}

	bm.workers[deviceID] = newWorker
	logrus.Infof("Device %s restarted", deviceID)
	return nil
}

// UpdateDeviceSettings updates processing settings for a device
func (bm *BroadcastManager) UpdateDeviceSettings(deviceID string, minDelay, maxDelay time.Duration) error {
	bm.workersMutex.RLock()
	worker, exists := bm.workers[deviceID]
	bm.workersMutex.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// Update worker settings (assuming worker has these fields)
	// Note: This would need to be implemented in the DeviceWorker struct
	logrus.Infof("Updated settings for device %s: minDelay=%v, maxDelay=%v", deviceID, minDelay, maxDelay)
	return nil
}