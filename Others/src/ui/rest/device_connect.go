package rest

import (
	"fmt"
	"sync"
	
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// DeviceLoginService manages login for specific devices
type DeviceLoginService struct {
	services map[string]domainApp.IAppUsecase // deviceID -> service
	db       *sqlstore.Container
	mutex    sync.RWMutex
}

var (
	deviceLoginService *DeviceLoginService
	loginServiceOnce   sync.Once
)

// GetDeviceLoginService returns the singleton device login service
func GetDeviceLoginService(db *sqlstore.Container) *DeviceLoginService {
	loginServiceOnce.Do(func() {
		deviceLoginService = &DeviceLoginService{
			services: make(map[string]domainApp.IAppUsecase),
			db:       db,
		}
	})
	return deviceLoginService
}

// GetOrCreateService gets or creates an app service for a specific device
func (d *DeviceLoginService) GetOrCreateService(deviceID string) domainApp.IAppUsecase {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// Check if service exists
	if service, exists := d.services[deviceID]; exists {
		return service
	}
	
	// Create new device and client
	device := d.db.NewDevice()
	client := whatsmeow.NewClient(device, waLog.Stdout(fmt.Sprintf("Device_%s", deviceID), config.WhatsappLogLevel, true))
	
	// Create new service
	service := usecase.NewAppService(client, d.db)
	d.services[deviceID] = service
	
	logrus.Infof("Created new app service for device: %s", deviceID)
	return service
}

// ConnectDevice handles device-specific QR code generation
func ConnectDevice(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "device_id is required",
		})
	}
	
	// Get user ID from session
	userID := c.Locals("userID")
	if userID == nil {
		// Try to get from cookie
		token := c.Cookies("session_token")
		if token == "" {
			return c.Status(401).JSON(utils.ResponseData{
				Status:  401,
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			})
		}
		
		userRepo := repository.GetUserRepository()
		session, err := userRepo.GetSession(token)
		if err != nil {
			return c.Status(401).JSON(utils.ResponseData{
				Status:  401,
				Code:    "UNAUTHORIZED",
				Message: "Invalid session",
			})
		}
		userID = session.UserID
	}
	
	// Verify device belongs to user
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil || device.UserID != userID.(string) {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device not found or access denied",
		})
	}
	
	// For now, return a message that this endpoint is ready
	// The actual implementation needs access to the database container
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS", 
		Message: "Device connection endpoint ready",
		Results: fiber.Map{
			"device_id": deviceID,
			"user_id":   userID,
			"device_name": device.DeviceName,
		},
	})
}
