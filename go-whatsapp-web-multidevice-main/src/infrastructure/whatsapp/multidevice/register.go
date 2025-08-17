package multidevice

import (
	"time"
	"go.mau.fi/whatsmeow"
	log "github.com/sirupsen/logrus"
)

// RegisterDevice registers a reconnected device with the DeviceManager
func (dm *DeviceManager) RegisterDevice(deviceID, userID, phone string, client *whatsmeow.Client) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.devices[deviceID] = &DeviceConnection{
		DeviceID:    deviceID,
		UserID:      userID,
		Phone:       phone,
		Client:      client,
		Store:       client.Store,
		Connected:   true,
		ConnectedAt: time.Now().Unix(),
	}
	
	log.Infof("Registered device %s with DeviceManager", deviceID)
}

// UnregisterDevice removes a device from the DeviceManager
func (dm *DeviceManager) UnregisterDevice(deviceID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	delete(dm.devices, deviceID)
	log.Infof("Unregistered device %s from DeviceManager", deviceID)
}
