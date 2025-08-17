package whatsapp

import (
	"context"
	"fmt"
	"time"
	
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

// CreateCleanDevice creates a new WhatsApp device with proper initialization
func CreateCleanDevice(container *sqlstore.Container, deviceID string) (*store.Device, error) {
	// First, ensure any existing device data is cleared
	err := ClearWhatsAppSessionData(deviceID)
	if err != nil {
		logrus.Warnf("Failed to clear existing session (continuing): %v", err)
	}
	
	// Create new device
	device := container.NewDevice()
	
	return device, nil
}

// SafeConnectWithRetry connects a WhatsApp client with retry logic
func SafeConnectWithRetry(client *whatsmeow.Client, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := client.Connect()
		if err == nil {
			return nil
		}
		
		// Check if already connected
		if err == whatsmeow.ErrAlreadyConnected {
			return nil
		}
		
		logrus.Warnf("Connect attempt %d failed: %v", i+1, err)
		
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

// WaitForLogin waits for the client to be fully logged in
func WaitForLogin(ctx context.Context, client *whatsmeow.Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for login")
		case <-ticker.C:
			if client.IsLoggedIn() {
				return nil
			}
		}
	}
}
