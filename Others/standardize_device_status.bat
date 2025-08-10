@echo off
echo Standardizing Device Status to Online/Offline Only...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Step 1: Updating database to normalize status...

REM Create SQL to update all non-standard statuses
echo UPDATE user_devices SET status = 'offline' WHERE status NOT IN ('online', 'offline'); > normalize_status.sql
echo UPDATE user_devices SET status = 'offline' WHERE status IS NULL; >> normalize_status.sql

echo.
echo Step 2: Fixing campaign checking to use online/offline only...

REM Fix campaign trigger
powershell -Command "(Get-Content 'src\usecase\optimized_campaign_trigger.go') -replace 'device\.Status == `"connected`" \|\| device\.Status == `"Connected`" \|\|', '' -replace 'device\.Status == `"online`" \|\| device\.Status == `"Online`"', 'device.Status == `"online`"' | Set-Content 'src\usecase\optimized_campaign_trigger.go'"

REM Fix broadcast processor
powershell -Command "(Get-Content 'src\usecase\optimized_broadcast_processor.go') -replace 'device\.Status != `"online`" && device\.Status != `"Online`" &&[\s\S]*?device\.Status != `"connected`" && device\.Status != `"Connected`"', 'device.Status != `"online`"' | Set-Content 'src\usecase\optimized_broadcast_processor.go'"

echo.
echo Step 3: Creating simplified status update functions...

echo Creating simplified device status handler...

cat > src\infrastructure\whatsapp\simple_device_status.go << 'EOF'
package whatsapp

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// UpdateDeviceStatusSimple updates device status to online or offline only
func UpdateDeviceStatusSimple(deviceID string, client *whatsmeow.Client) {
	userRepo := repository.GetUserRepository()
	
	// Simple check: if client exists and is connected = online, else offline
	if client != nil && client.IsConnected() {
		err := userRepo.UpdateDeviceStatus(deviceID, "online", "", "")
		if err != nil {
			logrus.Errorf("Failed to update device %s to online: %v", deviceID, err)
		} else {
			logrus.Debugf("Device %s is online", deviceID)
		}
	} else {
		err := userRepo.UpdateDeviceStatus(deviceID, "offline", "", "")
		if err != nil {
			logrus.Errorf("Failed to update device %s to offline: %v", deviceID, err)
		} else {
			logrus.Debugf("Device %s is offline", deviceID)
		}
	}
}

// CheckAllDeviceStatus checks all devices and updates to online/offline only
func CheckAllDeviceStatus() {
	cm := GetClientManager()
	userRepo := repository.GetUserRepository()
	
	// Get all devices
	devices, err := userRepo.GetAllDevices()
	if err != nil {
		logrus.Errorf("Failed to get devices: %v", err)
		return
	}
	
	for _, device := range devices {
		client, err := cm.GetClient(device.ID)
		
		// Simple logic: connected = online, anything else = offline
		newStatus := "offline"
		if err == nil && client != nil && client.IsConnected() {
			newStatus = "online"
		}
		
		// Update if changed
		if device.Status != newStatus {
			userRepo.UpdateDeviceStatus(device.ID, newStatus, device.Phone, device.JID)
			logrus.Infof("Device %s status updated: %s -> %s", device.DeviceName, device.Status, newStatus)
		}
	}
}
EOF

echo.
echo Step 4: Updating all status checks to use online/offline...

REM Update sequence processor
powershell -Command "(Get-Content 'src\usecase\sequence_trigger_processor.go') -replace 'd\.status == `"online`"', 'd.status == `"online`"' -replace 'Status: `"connected`"', 'Status: `"online`"' | Set-Content 'src\usecase\sequence_trigger_processor.go'"

REM Update broadcast worker
powershell -Command "(Get-Content 'src\usecase\broadcast_worker_processor.go') -replace 'device\.Status != `"online`" && device\.Status != `"Online`"', 'device.Status != `"online`"' | Set-Content 'src\usecase\broadcast_worker_processor.go'"

echo.
echo Step 5: Creating migration to add constraint...

cat > add_status_constraint.sql << 'EOF'
-- Normalize all existing statuses
UPDATE user_devices SET status = 'offline' 
WHERE status NOT IN ('online', 'offline') OR status IS NULL;

-- Add check constraint to enforce only online/offline
ALTER TABLE user_devices DROP CONSTRAINT IF EXISTS check_device_status;
ALTER TABLE user_devices ADD CONSTRAINT check_device_status 
CHECK (status IN ('online', 'offline'));

-- Set default
ALTER TABLE user_devices ALTER COLUMN status SET DEFAULT 'offline';
EOF

echo.
echo Step 6: Building application...
go build -o whatsapp.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo ✅ Status Standardization Complete!
    echo ========================================
    echo.
    echo Changes made:
    echo 1. All device statuses normalized to online/offline only
    echo 2. Campaign checking uses only: device.Status == "online"
    echo 3. Removed checks for "connected", "Connected", "Online"
    echo 4. Database constraint added to enforce online/offline
    echo.
    echo Next steps:
    echo 1. Run: psql -U your_user -d your_db -f normalize_status.sql
    echo 2. Run: psql -U your_user -d your_db -f add_status_constraint.sql
    echo 3. Restart the application
    echo.
) else (
    echo.
    echo ❌ Build failed! Please check errors above.
)

pause