// Add this to DeviceWorkerGroup struct in ultra_scale_broadcast_manager.go

// DeviceWorkerGroup manages multiple workers for a single device
type DeviceWorkerGroup struct {
	deviceID      string
	workers       []*BroadcastWorker
	messageQueue  chan *domainBroadcast.BroadcastMessage
	currentWorker int32 // For round-robin distribution
	mu            sync.RWMutex
	
	// Rate limiting - ensures sequential sending
	lastSentTime  time.Time
	sendMutex     sync.Mutex  // Only one worker can send at a time
}

// Add this method to DeviceWorkerGroup

// acquireSendPermission ensures only one worker sends at a time with proper delay
func (dwg *DeviceWorkerGroup) acquireSendPermission(minDelay, maxDelay int) {
	dwg.sendMutex.Lock()
	// Don't unlock here - the worker will unlock after sending
	
	// Calculate time since last send
	timeSinceLastSend := time.Since(dwg.lastSentTime)
	
	// Calculate required delay
	requiredDelay := calculateRandomDelay(minDelay, maxDelay)
	
	// If not enough time has passed, wait
	if timeSinceLastSend < requiredDelay {
		waitTime := requiredDelay - timeSinceLastSend
		logrus.Debugf("Device %s: Waiting %v before next send (rate limiting)", dwg.deviceID, waitTime)
		time.Sleep(waitTime)
	}
}

// releaseSendPermission updates last sent time and releases the mutex
func (dwg *DeviceWorkerGroup) releaseSendPermission() {
	dwg.lastSentTime = time.Now()
	dwg.sendMutex.Unlock()
}

// Update the processMessage method:

func (bw *BroadcastWorker) processMessage(msg *domainBroadcast.BroadcastMessage) {
	bw.mu.Lock()
	bw.status = "processing"
	bw.lastActivity = time.Now()
	bw.mu.Unlock()
	
	// Get the device worker group
	bw.pool.mu.RLock()
	group, exists := bw.pool.deviceGroups[bw.deviceID]
	bw.pool.mu.RUnlock()
	
	if !exists {
		logrus.Errorf("Worker %d: Device group not found for %s", bw.workerID, bw.deviceID)
		return
	}
	
	// Log which broadcast this message belongs to
	broadcastInfo := "Unknown broadcast"
	if msg.CampaignID != nil {
		broadcastInfo = fmt.Sprintf("Campaign %d", *msg.CampaignID)
	} else if msg.SequenceID != nil {
		broadcastInfo = fmt.Sprintf("Sequence %s", *msg.SequenceID)
	}
	
	// CRITICAL: Acquire send permission (this enforces rate limiting)
	minDelay := msg.MinDelay
	maxDelay := msg.MaxDelay
	if minDelay <= 0 {
		minDelay = 5  // Default minimum
	}
	if maxDelay <= 0 {
		maxDelay = 15 // Default maximum
	}
	
	// This will block until it's this worker's turn to send
	group.acquireSendPermission(minDelay, maxDelay)
	
	// Now we have exclusive permission to send
	logrus.Debugf("Worker %d on device %s sending message %s for %s to %s", 
		bw.workerID, bw.deviceID, msg.ID, broadcastInfo, msg.RecipientPhone)
	
	// Send via WhatsApp
	err := bw.sendWhatsAppMessage(msg)
	
	// IMPORTANT: Release permission after sending
	group.releaseSendPermission()
	
	// Update database status
	db := database.GetDB()
	if err != nil {
		atomic.AddInt64(&bw.failedCount, 1)
		if bw.pool != nil {
			atomic.AddInt64(&bw.pool.failedCount, 1)
		}
		db.Exec(`UPDATE broadcast_messages SET status = 'failed', error_message = $1, updated_at = NOW() WHERE id = $2`, 
			err.Error(), msg.ID)
		logrus.Errorf("Failed to send message %s: %v", msg.ID, err)
	} else {
		atomic.AddInt64(&bw.processedCount, 1)
		if bw.pool != nil {
			atomic.AddInt64(&bw.pool.processedCount, 1)
		}
		db.Exec(`UPDATE broadcast_messages SET status = 'sent', sent_at = NOW() WHERE id = $1`, msg.ID)
		
		if msg.SequenceID != nil {
			db.Exec(`UPDATE sequence_contacts SET last_message_at = NOW() WHERE sequence_id = $1 AND contact_phone = $2`,
				*msg.SequenceID, msg.RecipientPhone)
			db.Exec(`SELECT update_sequence_progress($1)`, *msg.SequenceID)
		}
		
		logrus.Infof("Worker %d successfully sent message to %s", bw.workerID, msg.RecipientPhone)
	}
	
	// NO DELAY HERE - it's handled by acquireSendPermission
	
	bw.mu.Lock()
	bw.status = "idle"
	bw.lastActivity = time.Now()
	bw.mu.Unlock()
}