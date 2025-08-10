package whatsapp

import (
	"go.mau.fi/whatsmeow"
)

// IsDeviceOnline checks if a device is online (simple check)
func IsDeviceOnline(client *whatsmeow.Client) bool {
	return client != nil && client.IsConnected()
}

// GetSimpleStatus returns "online" or "offline" only
func GetSimpleStatus(client *whatsmeow.Client) string {
	if IsDeviceOnline(client) {
		return "online"
	}
	return "offline"
}

// Constants for device status
const (
	StatusOnline  = "online"
	StatusOffline = "offline"
)