// Fix 1: Update dashboard.html WebSocket handler
// Add this case to the switch statement in connectWebSocket() function (around line 2835)

case 'DEVICE_LOGGED_OUT':
    // Update device status to offline when logged out
    console.log('Device logged out:', data.result);
    const loggedOutDeviceId = data.result?.deviceId;
    if (loggedOutDeviceId) {
        const device = devices.find(d => d.id === loggedOutDeviceId);
        if (device) {
            device.status = 'offline';
            device.phone = '';
            device.jid = '';
            device.lastSeen = new Date().toISOString();
            renderDevices();
            
            // Show notification
            showAlert('warning', `Device ${device.name} has been logged out`);
        }
    }
    break;
