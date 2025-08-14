package sequence

import (
	"context"
	"database/sql"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	whatsapp "go.mau.fi/whatsmeow"
	"github.com/sirupsen/logrus"
)

// SequenceWorkerManager manages all sequence workers
type SequenceWorkerManager struct {
	workers       map[string]*SequenceDeviceWorker
	mu            sync.RWMutex
	maxWorkers    int
	db            *sql.DB
	checkInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewSequenceWorkerManager creates a new manager
func NewSequenceWorkerManager(db *sql.DB, maxWorkers int) *SequenceWorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SequenceWorkerManager{
		workers:       make(map[string]*SequenceDeviceWorker),
		maxWorkers:    maxWorkers,
		db:            db,
		checkInterval: 30 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the worker manager
func (swm *SequenceWorkerManager) Start() {
	logrus.Info("Starting Sequence Worker Manager...")
	go swm.processLoop()
	go swm.healthCheckLoop()
	go swm.metricsLoop()
}

// processLoop checks for contacts ready to process
func (swm *SequenceWorkerManager) processLoop() {
	ticker := time.NewTicker(swm.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-swm.ctx.Done():
			return
		case <-ticker.C:
			swm.processSequenceContacts()
		}
	}
}

// processSequenceContacts gets contacts and assigns to workers
func (swm *SequenceWorkerManager) processSequenceContacts() {
	start := time.Now()
	
	// Query to get contacts grouped by assigned device
	query := `
		WITH ready_contacts AS (
			SELECT sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
				sc.current_trigger, sc.next_trigger_time,
				COALESCE(sc.assigned_worker_id, l.device_id) AS device_id,
				ss.content, ss.message_type, ss.media_url, ss.caption,
				ss.next_trigger, ss.trigger_delay_hours,
				s.min_delay_seconds, s.max_delay_seconds
			FROM sequence_contacts sc
			JOIN sequences s ON s.id = sc.sequence_id
			JOIN sequence_steps ss ON ss.sequence_id = sc.sequence_id 
				AND ss.trigger = sc.current_trigger
			LEFT JOIN leads l ON l.phone = sc.contact_phone
			WHERE sc.status = 'active'
				AND s.status = 'active'
				AND sc.next_trigger_time <= NOW()
				AND sc.processing_device_id IS NULL
			ORDER BY sc.next_trigger_time ASC
			LIMIT 1000
		)
		SELECT * FROM ready_contacts
	`
	
	rows, err := swm.db.Query(query)
	if err != nil {
		logrus.Errorf("Failed to query ready contacts: %v", err)
		return
	}
	defer rows.Close()
	
	// Group contacts by device
	contactsByDevice := make(map[string][]SequenceContact)
	totalContacts := 0
	
	for rows.Next() {
		var contact SequenceContact
		err := rows.Scan(
			&contact.ID, &contact.SequenceID, &contact.Phone, &contact.Name,
			&contact.CurrentTrigger, &contact.NextTriggerTime, &contact.DeviceID,
			&contact.Content, &contact.MessageType, &contact.MediaURL, &contact.Caption,
			&contact.NextTrigger, &contact.TriggerDelayHours,
			&contact.MinDelay, &contact.MaxDelay,
		)
		if err != nil {
			logrus.Errorf("Error scanning contact: %v", err)
			continue
		}
		
		contactsByDevice[contact.DeviceID] = append(contactsByDevice[contact.DeviceID], contact)
		totalContacts++
	}
	
	// Assign contacts to workers
	assignedCount := 0
	for deviceID, contacts := range contactsByDevice {
		worker := swm.GetOrCreateWorker(deviceID)
		if worker != nil {
			assigned := worker.QueueContacts(contacts)
			assignedCount += assigned
		}
	}
	
	if totalContacts > 0 {
		logrus.Infof("Sequence processing: found=%d, assigned=%d, duration=%v", 
			totalContacts, assignedCount, time.Since(start))
	}
}

// GetOrCreateWorker gets existing or creates new worker
func (swm *SequenceWorkerManager) GetOrCreateWorker(deviceID string) *SequenceDeviceWorker {
	swm.mu.RLock()
	worker, exists := swm.workers[deviceID]
	swm.mu.RUnlock()
	
	if exists && worker.IsHealthy() {
		return worker
	}
	
	swm.mu.Lock()
	defer swm.mu.Unlock()
	
	// Check again after acquiring write lock
	if worker, exists := swm.workers[deviceID]; exists && worker.IsHealthy() {
		return worker
	}
	
	// Check worker limit
	if len(swm.workers) >= swm.maxWorkers {
		// Find and remove idle workers
		swm.cleanupIdleWorkers()
		
		if len(swm.workers) >= swm.maxWorkers {
			logrus.Warnf("Max workers reached (%d), cannot create worker for device %s", 
				swm.maxWorkers, deviceID)
			return nil
		}
	}
	
	// Create new worker
	worker = NewSequenceDeviceWorker(deviceID, swm.db)
	if err := worker.Initialize(); err != nil {
		logrus.Errorf("Failed to initialize worker for device %s: %v", deviceID, err)
		return nil
	}
	
	swm.workers[deviceID] = worker
	go worker.Run()
	
	logrus.Infof("Created sequence worker for device %s (total workers: %d)", 
		deviceID, len(swm.workers))
	
	return worker
}

// cleanupIdleWorkers removes workers that have been idle
func (swm *SequenceWorkerManager) cleanupIdleWorkers() {
	idleTimeout := 5 * time.Minute
	now := time.Now()
	
	for deviceID, worker := range swm.workers {
		if worker.GetIdleTime() > idleTimeout {
			worker.Stop()
			delete(swm.workers, deviceID)
			logrus.Infof("Removed idle worker for device %s", deviceID)
		}
	}
}

// healthCheckLoop monitors worker health
func (swm *SequenceWorkerManager) healthCheckLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-swm.ctx.Done():
			return
		case <-ticker.C:
			swm.checkWorkerHealth()
		}
	}
}

// Stop gracefully shuts down the manager
func (swm *SequenceWorkerManager) Stop() {
	logrus.Info("Stopping Sequence Worker Manager...")
	swm.cancel()
	
	swm.mu.Lock()
	defer swm.mu.Unlock()
	
	for _, worker := range swm.workers {
		worker.Stop()
	}
}
