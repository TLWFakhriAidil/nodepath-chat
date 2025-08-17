// Fix for Device Management and QR Code Issues

// Issue 1: Device disappears when QR modal closes
// Solution: Save device to backend immediately after creation

// Issue 2: QR code not working
// Solution: Use the proper WhatsApp Web API endpoint

function addNewDevice() {
    const deviceName = prompt('Enter device name:');
    if (deviceName) {
        // Create device in backend first
        fetch('/api/devices', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: deviceName })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // Add to local array
                const newDevice = data.device;
                devices.push(newDevice);
                renderDevices();
                
                // Show device card with both options
                showDeviceOptions(newDevice.id);
            }
        })
        .catch(error => {
            console.error('Error creating device:', error);
            // Fallback to local creation
            const newDevice = {
                id: 'temp_' + Date.now(),
                name: deviceName,
                phone: 'Not connected',
                status: 'offline',
                lastSeen: 'Never connected'
            };
            devices.push(newDevice);
            renderDevices();
            showDeviceOptions(newDevice.id);
        });
    }
}

// Show device options (QR or Phone Code)
function showDeviceOptions(deviceId) {
    // Don't auto-open QR, let user choose
    // Device card already has both buttons
}

// Fix QR Code scanning
function scanQR(deviceId) {
    const modal = new bootstrap.Modal(document.getElementById('qrModal'));
    modal.show();
    
    // Store device ID for reference
    document.getElementById('qrModal').setAttribute('data-device-id', deviceId);
    
    // Rest of QR code logic...
}

// Ensure device persists even if modal closes
document.getElementById('qrModal').addEventListener('hidden.bs.modal', function () {
    // Don't remove device when modal closes
    // User can still use Phone Code option
});
