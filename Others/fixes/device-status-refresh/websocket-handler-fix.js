// WebSocket message handler fix for DEVICE_CONNECTED event
// Add this to dashboard.html to handle device connection updates

// Listen for DEVICE_CONNECTED WebSocket messages
if (typeof websocket !== 'undefined' && websocket) {
    const originalOnMessage = websocket.onmessage;
    websocket.onmessage = function(event) {
        // Call original handler if exists
        if (originalOnMessage) {
            originalOnMessage.call(this, event);
        }
        
        try {
            const data = JSON.parse(event.data);
            
            // Handle DEVICE_CONNECTED message
            if (data.code === 'DEVICE_CONNECTED' && data.results) {
                console.log('Device connected event received:', data.results);
                
                // Extract device info
                const { jid, phone } = data.results;
                
                // Reload devices to show updated status
                console.log('Reloading device list to show connected status...');
                setTimeout(() => {
                    loadDevices(true); // Silent reload
                }, 1000); // Wait 1 second for database update
                
                // Show success notification
                if (typeof showNotification === 'function') {
                    showNotification('Device Connected', `WhatsApp connected: ${phone}`, 'success');
                }
            }
        } catch (e) {
            // Ignore JSON parse errors
        }
    };
}

// Alternative: Direct UI update without full reload
function updateDeviceStatusInUI(jid, phone) {
    // Find device card with "Not linked" status
    const deviceCards = document.querySelectorAll('.device-card');
    deviceCards.forEach(card => {
        const phoneElement = card.querySelector('.text-muted.small');
        if (phoneElement && phoneElement.textContent === 'Disconnected') {
            // Update status indicator
            phoneElement.textContent = 'Connected';
            
            // Update phone number if found
            const phoneDisplay = card.querySelector('h5').parentElement.nextElementSibling;
            if (phoneDisplay && phoneDisplay.textContent.includes('Not linked')) {
                phoneDisplay.innerHTML = `Phone Number<br>${phone}`;
            }
            
            // Update status dot color
            const statusDot = card.querySelector('.device-status');
            if (statusDot) {
                statusDot.classList.remove('disconnected');
                statusDot.classList.add('online');
            }
            
            // Update card style
            card.classList.add('connected');
        }
    });
}
