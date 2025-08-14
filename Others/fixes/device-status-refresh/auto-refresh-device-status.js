// Auto-refresh device status after QR scan
// Add this to dashboard.html to automatically update device status

// Function to check and update device status
function checkDeviceStatusUpdates() {
    // Only run if we're on the devices tab
    const devicesTab = document.getElementById('devices-tab');
    if (!devicesTab || !devicesTab.classList.contains('active')) {
        return;
    }
    
    // Check each device card for "Disconnected" status
    const deviceCards = document.querySelectorAll('.device-card');
    const disconnectedDevices = Array.from(deviceCards).filter(card => {
        const statusText = card.querySelector('.text-muted.small');
        return statusText && statusText.textContent === 'Disconnected';
    });
    
    // If we have disconnected devices, reload the device list
    if (disconnectedDevices.length > 0) {
        console.log('Found disconnected devices, checking for updates...');
        loadDevices(true); // silent reload
    }
}

// Start checking for status updates every 5 seconds
let statusCheckInterval;
function startDeviceStatusCheck() {
    // Clear any existing interval
    if (statusCheckInterval) {
        clearInterval(statusCheckInterval);
    }
    
    // Check every 5 seconds
    statusCheckInterval = setInterval(checkDeviceStatusUpdates, 5000);
}

// Stop checking when switching tabs
function stopDeviceStatusCheck() {
    if (statusCheckInterval) {
        clearInterval(statusCheckInterval);
        statusCheckInterval = null;
    }
}

// Hook into tab switching
document.addEventListener('DOMContentLoaded', function() {
    // Start checking when devices tab is active
    const devicesTabButton = document.querySelector('[data-bs-target="#devices"]');
    if (devicesTabButton) {
        devicesTabButton.addEventListener('shown.bs.tab', function() {
            startDeviceStatusCheck();
        });
        
        // Stop checking when switching to other tabs
        const otherTabs = document.querySelectorAll('.nav-link[data-bs-toggle="tab"]:not([data-bs-target="#devices"])');
        otherTabs.forEach(tab => {
            tab.addEventListener('shown.bs.tab', function() {
                stopDeviceStatusCheck();
            });
        });
    }
    
    // If devices tab is initially active, start checking
    const devicesTab = document.getElementById('devices-tab');
    if (devicesTab && devicesTab.classList.contains('active')) {
        startDeviceStatusCheck();
    }
});

// Also listen for WebSocket messages about device status
if (typeof websocket !== 'undefined' && websocket) {
    const originalOnMessage = websocket.onmessage;
    websocket.onmessage = function(event) {
        // Call original handler
        if (originalOnMessage) {
            originalOnMessage.call(this, event);
        }
        
        // Check for login success or device connection messages
        try {
            const data = JSON.parse(event.data);
            if (data.code === 'LOGIN_SUCCESS' || 
                data.code === 'CONNECTED' || 
                data.message && data.message.includes('Successfully paired')) {
                console.log('Device connection detected, refreshing device list...');
                setTimeout(() => {
                    loadDevices(true);
                }, 2000); // Wait 2 seconds for database update
            }
        } catch (e) {
            // Ignore parse errors
        }
    };
}

console.log('Device status auto-refresh initialized');
