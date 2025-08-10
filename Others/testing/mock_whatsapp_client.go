package whatsapp

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// MockWhatsAppClient simulates WhatsApp operations without actual sending
type MockWhatsAppClient struct {
	*whatsmeow.Client
	deviceID      string
	isConnected   bool
	isLoggedIn    bool
	messagesSent  int
	mu            sync.Mutex
	simulateDelay bool
	failureRate   float32 // 0.0 to 1.0 (e.g., 0.05 = 5% failure rate)
}

// MockClientManager manages mock WhatsApp clients for testing
type MockClientManager struct {
	clients sync.Map
	mu      sync.RWMutex
}

var (
	mockManager     *MockClientManager
	mockManagerOnce sync.Once
)

// GetMockClientManager returns singleton instance of mock manager
func GetMockClientManager() *MockClientManager {
	mockManagerOnce.Do(func() {
		mockManager = &MockClientManager{}
	})
	return mockManager
}

// CreateMockClient creates a new mock WhatsApp client
func (m *MockClientManager) CreateMockClient(deviceID string) *MockWhatsAppClient {
	client := &MockWhatsAppClient{
		deviceID:      deviceID,
		isConnected:   true,
		isLoggedIn:    true,
		simulateDelay: true,
		failureRate:   0.02, // 2% failure rate
	}
	
	m.clients.Store(deviceID, client)
	return client
}

// GetMockClient retrieves a mock client
func (m *MockClientManager) GetMockClient(deviceID string) (*MockWhatsAppClient, error) {
	if client, ok := m.clients.Load(deviceID); ok {
		return client.(*MockWhatsAppClient), nil
	}
	return nil, fmt.Errorf("client not found for device %s", deviceID)
}

// SendMessage simulates sending a message
func (c *MockWhatsAppClient) SendMessage(ctx context.Context, to types.JID, message *waProto.Message) (types.SendResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Simulate network delay
	if c.simulateDelay {
		delay := time.Duration(rand.Intn(200)+50) * time.Millisecond
		time.Sleep(delay)
	}
	
	// Simulate random failures
	if rand.Float32() < c.failureRate {
		logrus.Warnf("[MOCK] Simulated failure sending message from device %s to %s", c.deviceID, to.String())
		return types.SendResponse{}, fmt.Errorf("simulated send failure")
	}
	
	// Simulate successful send
	c.messagesSent++
	logrus.Debugf("[MOCK] Successfully 'sent' message #%d from device %s to %s", c.messagesSent, c.deviceID, to.String())
	
	// Return mock response
	return types.SendResponse{
		Timestamp: time.Now(),
	}, nil
}

// IsConnected returns mock connection status
func (c *MockWhatsAppClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

// IsLoggedIn returns mock login status  
func (c *MockWhatsAppClient) IsLoggedIn() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isLoggedIn
}

// Connect simulates connection
func (c *MockWhatsAppClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.isConnected {
		return nil
	}
	
	// Simulate connection delay
	time.Sleep(100 * time.Millisecond)
	
	// 95% success rate for connections
	if rand.Float32() < 0.95 {
		c.isConnected = true
		logrus.Infof("[MOCK] Device %s connected successfully", c.deviceID)
		return nil
	}
	
	return fmt.Errorf("simulated connection failure")
}

// Disconnect simulates disconnection
func (c *MockWhatsAppClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isConnected = false
	logrus.Infof("[MOCK] Device %s disconnected", c.deviceID)
}

// GetStats returns mock client statistics
func (c *MockWhatsAppClient) GetStats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return map[string]interface{}{
		"device_id":     c.deviceID,
		"is_connected":  c.isConnected,
		"is_logged_in":  c.isLoggedIn,
		"messages_sent": c.messagesSent,
		"failure_rate":  c.failureRate,
	}
}

// MockStore simulates WhatsApp Store
type MockStore struct {
	ID *types.JID
}

// GetStore returns mock store
func (c *MockWhatsAppClient) GetStore() *MockStore {
	jid := types.NewJID("mock", types.DefaultUserServer)
	return &MockStore{ID: &jid}
}
