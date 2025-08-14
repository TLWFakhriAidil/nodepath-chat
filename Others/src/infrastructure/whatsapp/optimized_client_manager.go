package whatsapp

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// OptimizedClientManager manages WhatsApp clients with performance optimizations
type OptimizedClientManager struct {
	// Sharded client storage for better concurrency
	shards    []*clientShard
	shardCount int
	
	// Connection pool for rate limiting
	connectionSemaphore chan struct{}
	
	// Background workers
	workerPool    *sync.Pool
	shutdownChan  chan struct{}
	wg            sync.WaitGroup
}

type clientShard struct {
	clients map[string]*deviceClient
	mutex   sync.RWMutex
}

type deviceClient struct {
	client      *whatsmeow.Client
	lastAccess  time.Time
	syncStatus  string
	messageBuffer []repository.WhatsAppMessage
	bufferMutex  sync.Mutex
}

var (
	optimizedManager *OptimizedClientManager
	optOnce          sync.Once
)

// GetOptimizedClientManager returns the optimized singleton client manager
func GetOptimizedClientManager() *OptimizedClientManager {
	optOnce.Do(func() {
		// Calculate optimal shard count based on CPU cores
		shardCount := runtime.NumCPU() * 4
		if shardCount < 16 {
			shardCount = 16
		}
		
		optimizedManager = &OptimizedClientManager{
			shardCount:          shardCount,
			shards:              make([]*clientShard, shardCount),
			connectionSemaphore: make(chan struct{}, 100), // Max 100 concurrent connections
			shutdownChan:        make(chan struct{}),
			workerPool: &sync.Pool{
				New: func() interface{} {
					return &syncWorker{}
				},
			},
		}
		
		// Initialize shards
		for i := 0; i < shardCount; i++ {
			optimizedManager.shards[i] = &clientShard{
				clients: make(map[string]*deviceClient),
			}
		}
		
		// Start background workers
		optimizedManager.startBackgroundWorkers()
	})
	return optimizedManager
}

type syncWorker struct {
	deviceID string
	client   *deviceClient
}

// Hash function to determine shard
func (ocm *OptimizedClientManager) getShard(deviceID string) *clientShard {
	hash := 0
	for _, char := range deviceID {
		hash = (hash * 31 + int(char)) % ocm.shardCount
	}
	return ocm.shards[hash]
}

// AddClient adds a client with optimizations
func (ocm *OptimizedClientManager) AddClient(deviceID string, client *whatsmeow.Client) {
	shard := ocm.getShard(deviceID)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	
	shard.clients[deviceID] = &deviceClient{
		client:        client,
		lastAccess:    time.Now(),
		syncStatus:    "pending",
		messageBuffer: make([]repository.WhatsAppMessage, 0, 100),
	}
	
	// Queue for background sync
	ocm.queueDeviceSync(deviceID)
}

// GetClient retrieves a client with minimal locking
func (ocm *OptimizedClientManager) GetClient(deviceID string) (*whatsmeow.Client, error) {
	shard := ocm.getShard(deviceID)
	shard.mutex.RLock()
	deviceClient, exists := shard.clients[deviceID]
	shard.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no client found for device %s", deviceID)
	}
	
	// Update last access time
	deviceClient.lastAccess = time.Now()
	
	if !deviceClient.client.IsConnected() {
		return nil, fmt.Errorf("client for device %s is not connected", deviceID)
	}
	
	return deviceClient.client, nil
}

// BatchGetClients retrieves multiple clients efficiently
func (ocm *OptimizedClientManager) BatchGetClients(deviceIDs []string) map[string]*whatsmeow.Client {
	// Group devices by shard to minimize lock contention
	shardGroups := make(map[int][]string)
	for _, deviceID := range deviceIDs {
		hash := 0
		for _, char := range deviceID {
			hash = (hash * 31 + int(char)) % ocm.shardCount
		}
		shardGroups[hash] = append(shardGroups[hash], deviceID)
	}
	
	results := make(map[string]*whatsmeow.Client)
	resultMutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	
	// Process each shard group concurrently
	for shardIdx, devices := range shardGroups {
		wg.Add(1)
		go func(idx int, devs []string) {
			defer wg.Done()
			
			shard := ocm.shards[idx]
			shard.mutex.RLock()
			defer shard.mutex.RUnlock()
			
			for _, deviceID := range devs {
				if dc, exists := shard.clients[deviceID]; exists && dc.client.IsConnected() {
					resultMutex.Lock()
					results[deviceID] = dc.client
					resultMutex.Unlock()
				}
			}
		}(shardIdx, devices)
	}
	
	wg.Wait()
	return results
}

// BufferMessage adds a message to the buffer for batch processing
func (ocm *OptimizedClientManager) BufferMessage(deviceID string, message repository.WhatsAppMessage) {
	shard := ocm.getShard(deviceID)
	shard.mutex.RLock()
	deviceClient, exists := shard.clients[deviceID]
	shard.mutex.RUnlock()
	
	if !exists {
		return
	}
	
	deviceClient.bufferMutex.Lock()
	deviceClient.messageBuffer = append(deviceClient.messageBuffer, message)
	
	// Flush if buffer is getting full
	if len(deviceClient.messageBuffer) >= 50 {
		messages := deviceClient.messageBuffer
		deviceClient.messageBuffer = make([]repository.WhatsAppMessage, 0, 100)
		deviceClient.bufferMutex.Unlock()
		
		// Process in background
		go ocm.flushMessages(messages)
	} else {
		deviceClient.bufferMutex.Unlock()
	}
}

// flushMessages saves messages to database in batch
func (ocm *OptimizedClientManager) flushMessages(messages []repository.WhatsAppMessage) {
	if len(messages) == 0 {
		return
	}
	
	repo := repository.GetWhatsAppRepository()
	for _, msg := range messages {
		if err := repo.SaveMessage(&msg); err != nil {
			fmt.Printf("Error saving message: %v\n", err)
		}
	}
}

// Background workers
func (ocm *OptimizedClientManager) startBackgroundWorkers() {
	// Periodic cleanup worker
	ocm.wg.Add(1)
	go func() {
		defer ocm.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ocm.cleanupInactiveClients()
			case <-ocm.shutdownChan:
				return
			}
		}
	}()
	
	// Message flush worker
	ocm.wg.Add(1)
	go func() {
		defer ocm.wg.Done()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ocm.flushAllBuffers()
			case <-ocm.shutdownChan:
				return
			}
		}
	}()
}

// cleanupInactiveClients removes clients that haven't been accessed recently
func (ocm *OptimizedClientManager) cleanupInactiveClients() {
	// Get start of today (midnight)
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	userRepo := repository.GetUserRepository()
	
	for _, shard := range ocm.shards {
		shard.mutex.Lock()
		for deviceID, dc := range shard.clients {
			// Check database status instead of WhatsApp connection
			device, err := userRepo.GetDeviceByID(deviceID)
			if err != nil || device == nil || device.Status != "online" {
				// Device is offline in database
				// Only remove if lastAccess is before today (yesterday or older)
				if dc.lastAccess.Before(startOfToday) {
					delete(shard.clients, deviceID)
					logrus.Debugf("Removed offline device %s from memory (last access: %s)", 
						deviceID, dc.lastAccess.Format("2006-01-02 15:04:05"))
				}
			}
		}
		shard.mutex.Unlock()
	}
}

// flushAllBuffers flushes all message buffers
func (ocm *OptimizedClientManager) flushAllBuffers() {
	for _, shard := range ocm.shards {
		shard.mutex.RLock()
		devices := make([]string, 0, len(shard.clients))
		for deviceID := range shard.clients {
			devices = append(devices, deviceID)
		}
		shard.mutex.RUnlock()
		
		for _, deviceID := range devices {
			shard.mutex.RLock()
			dc, exists := shard.clients[deviceID]
			shard.mutex.RUnlock()
			
			if exists {
				dc.bufferMutex.Lock()
				if len(dc.messageBuffer) > 0 {
					messages := dc.messageBuffer
					dc.messageBuffer = make([]repository.WhatsAppMessage, 0, 100)
					dc.bufferMutex.Unlock()
					
					go ocm.flushMessages(messages)
				} else {
					dc.bufferMutex.Unlock()
				}
			}
		}
	}
}

// queueDeviceSync queues a device for background sync
func (ocm *OptimizedClientManager) queueDeviceSync(deviceID string) {
	// Rate limit new syncs
	select {
	case ocm.connectionSemaphore <- struct{}{}:
		go func() {
			defer func() { <-ocm.connectionSemaphore }()
			time.Sleep(2 * time.Second) // Wait for connection to stabilize
			GetChatsForDevice(deviceID)
		}()
	default:
		// Queue is full, sync later
		time.AfterFunc(10*time.Second, func() {
			ocm.queueDeviceSync(deviceID)
		})
	}
}

// GetActiveDeviceCount returns the number of active devices
func (ocm *OptimizedClientManager) GetActiveDeviceCount() int {
	// Use database status instead of checking WhatsApp connection
	userRepo := repository.GetUserRepository()
	devices, err := userRepo.GetAllDevices()
	if err != nil {
		return 0
	}
	
	count := 0
	for _, device := range devices {
		if device.Status == "online" {
			count++
		}
	}
	return count
}

// Shutdown gracefully shuts down the manager
func (ocm *OptimizedClientManager) Shutdown() {
	close(ocm.shutdownChan)
	ocm.wg.Wait()
	
	// Flush all remaining buffers
	ocm.flushAllBuffers()
}
