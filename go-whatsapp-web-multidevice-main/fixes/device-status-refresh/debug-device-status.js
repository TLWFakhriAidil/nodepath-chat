// Debug script to check device status
// Add this to the browser console to debug device status issues

// Function to check device data
function debugDeviceStatus() {
    console.log('=== Device Status Debug ===');
    
    // Check if devices array exists
    if (typeof devices !== 'undefined') {
        console.log('Devices array:', devices);
        devices.forEach(device => {
            console.log(`Device ${device.name}:`, {
                id: device.id,
                status: device.status,
                phone: device.phone,
                statusType: typeof device.status,
                isOnline: device.status === 'online',
                isConnected: device.status === 'connected'
            });
        });
    } else {
        console.log('Devices array not found');
    }
    
    // Fetch fresh data from API
    console.log('\n=== Fetching fresh data from API ===');
    fetch('/api/devices', { credentials: 'include' })
        .then(response => response.json())
        .then(data => {
            console.log('API Response:', data);
            if (data.results && Array.isArray(data.results)) {
                data.results.forEach(device => {
                    console.log(`Device ${device.name} from API:`, {
                        id: device.id,
                        status: device.status,
                        phone: device.phone,
                        statusType: typeof device.status,
                        rawStatus: JSON.stringify(device.status)
                    });
                });
            }
        })
        .catch(error => console.error('Error fetching devices:', error));
    
    // Check DOM elements
    console.log('\n=== DOM Elements ===');
    const deviceCards = document.querySelectorAll('.device-card');
    deviceCards.forEach(card => {
        const name = card.querySelector('h5')?.textContent;
        const statusText = card.querySelector('.text-muted.small')?.textContent;
        console.log(`Device card ${name}:`, {
            statusText: statusText,
            hasConnectedClass: card.classList.contains('connected')
        });
    });
}

// Run the debug function
debugDeviceStatus();

// Also log when loadDevices is called
if (typeof loadDevices === 'function') {
    const originalLoadDevices = loadDevices;
    loadDevices = function(silent) {
        console.log('loadDevices called with silent:', silent);
        return originalLoadDevices.call(this, silent);
    };
}
