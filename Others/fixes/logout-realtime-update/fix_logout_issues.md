# Fix for WhatsApp Device Logout Issues

## Issue 1: Device logout status not updating in real-time
The WebSocket broadcast for `DEVICE_LOGGED_OUT` is sent from the backend but not handled in the frontend.

## Issue 2: Cannot reconnect after logout - Foreign key constraint error
When a device logs out, the WhatsApp session data isn't properly cleaned up, causing foreign key violations on reconnection.

## Solutions:

### 1. Add WebSocket handler for DEVICE_LOGGED_OUT in dashboard.html

In the WebSocket message handler switch statement (around line 2835), add:

```javascript
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
```

### 2. Fix the logout function to properly clear WhatsApp session

Update the device logout handler to clear all WhatsApp session data before removing the client.

### 3. Fix the reconnection issue

When reconnecting, ensure old session data is completely removed before creating a new session.
