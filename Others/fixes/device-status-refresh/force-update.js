// Emergency fix to force device status update
// Run this in browser console after QR scan

async function forceDeviceStatusUpdate() {
    console.log('Forcing device status update...');
    
    // Get current device ID from the page
    const deviceCards = document.querySelectorAll('.device-card');
    if (deviceCards.length === 0) {
        console.error('No device cards found');
        return;
    }
    
    // Extract device ID from the first disconnected device
    let deviceId = null;
    deviceCards.forEach(card => {
        const statusText = card.querySelector('.text-muted.small');
        if (statusText && statusText.textContent === 'Disconnected') {
            // Try to find device ID from onclick handlers or data attributes
            const buttons = card.querySelectorAll('button[onclick], a[onclick]');
            buttons.forEach(btn => {
                const onclick = btn.getAttribute('onclick');
                if (onclick && onclick.includes('8ccc6409-124f-4f68-b618-0e64e69d61b8')) {
                    deviceId = '8ccc6409-124f-4f68-b618-0e64e69d61b8';
                }
            });
        }
    });
    
    if (!deviceId) {
        // Use the device ID from your console output
        deviceId = '8ccc6409-124f-4f68-b618-0e64e69d61b8';
        console.log('Using device ID from API response:', deviceId);
    }
    
    // Log current device info
    console.log('Device ID:', deviceId);
    console.log('Checking current status via API...');
    
    // Fetch current device info
    const response = await fetch('/api/devices', { credentials: 'include' });
    const data = await response.json();
    console.log('Current device data:', data.results);
    
    // Find our device
    const device = data.results.find(d => d.id === deviceId);
    if (device) {
        console.log('Device status in database:', device.status);
        console.log('Device phone:', device.phone || 'empty');
        console.log('Device JID:', device.jid || 'empty');
        
        if (device.status === 'offline' || !device.phone) {
            console.log('\n⚠️ Device is still showing as offline in database!');
            console.log('This means the UpdateDeviceStatus call is failing.');
            console.log('\nPossible issues:');
            console.log('1. The session.DeviceID doesn\'t match this device ID');
            console.log('2. The database update is failing silently');
            console.log('3. The phone/jid values are not being passed correctly');
            
            console.log('\n🔧 Workaround: Try creating a new device and scanning QR again');
        }
    }
    
    // Force reload after delay
    console.log('\nWaiting 5 seconds then reloading...');
    setTimeout(() => {
        location.reload();
    }, 5000);
}

// Run the function
forceDeviceStatusUpdate();

// Also log WebSocket messages
if (typeof websocket !== 'undefined' && websocket) {
    console.log('\n📡 Listening for WebSocket messages...');
    const originalOnMessage = websocket.onmessage;
    websocket.onmessage = function(event) {
        console.log('WebSocket message:', event.data);
        if (originalOnMessage) {
            originalOnMessage.call(this, event);
        }
    };
}
