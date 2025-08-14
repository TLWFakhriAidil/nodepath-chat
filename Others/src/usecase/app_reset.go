package usecase

import (
	"context"
	"fmt"
	"time"
	
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// ResetWhatsAppConnection forces a complete reset of the WhatsApp connection
func (service serviceApp) ResetWhatsAppConnection(ctx context.Context) error {
	logrus.Info("Resetting WhatsApp connection...")
	
	// Disconnect if connected
	if service.WaCli != nil {
		service.WaCli.Disconnect()
		time.Sleep(1 * time.Second)
	}
	
	// Clear the device store to force new QR
	if service.db != nil {
		device, err := service.db.GetFirstDevice(ctx)
		if err == nil && device != nil {
			logrus.Info("Clearing device session...")
			err = device.Delete(ctx)
			if err != nil {
				logrus.Errorf("Failed to delete device: %v", err)
			}
		}
	}
	
	// Recreate the client
	device := service.db.NewDevice()
	
	service.WaCli = whatsmeow.NewClient(device, nil)
	logrus.Info("WhatsApp connection reset complete")
	
	return nil
}

// LoginWithReset attempts login with automatic reset on failure
func (service serviceApp) LoginWithReset(ctx context.Context) (response domainApp.LoginResponse, err error) {
	// First try normal login
	response, err = service.Login(ctx)
	
	// If QR channel error, reset and try again
	if err == pkgError.ErrQrChannel {
		logrus.Warn("QR channel error detected, attempting reset...")
		
		resetErr := service.ResetWhatsAppConnection(ctx)
		if resetErr != nil {
			logrus.Errorf("Reset failed: %v", resetErr)
			return response, fmt.Errorf("failed to reset connection: %v", resetErr)
		}
		
		// Try login again after reset
		return service.Login(ctx)
	}
	
	return response, err
}
