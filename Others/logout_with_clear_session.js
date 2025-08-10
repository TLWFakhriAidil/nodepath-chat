// Enhanced Logout Device Function with Proper Session Clearing
// This version clears WhatsApp session tables for the specific device

function logoutDevice(deviceId) {
    Swal.fire({
        title: 'Logout Device?',
        html: `
            <div class="text-start">
                <p>This will disconnect the device from WhatsApp and clear its session.</p>
                <p class="text-warning mb-0"><i class="bi bi-exclamation-triangle me-2"></i>You will need to scan the QR code again to reconnect.</p>
            </div>
        `,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Yes, Logout',
        cancelButtonText: 'Cancel',
        confirmButtonColor: '#dc3545'
    }).then((result) => {
        if (result.isConfirmed) {
            showLoading();
            
            // Set a timeout to hide loading in case of any issues
            const loadingTimeout = setTimeout(() => {
                hideLoading();
                showToast('Operation took too long. Please refresh the page.', 'warning');
            }, 30000); // 30 seconds timeout
            
            // First logout the device
            fetch('/app/logout?deviceId=' + deviceId, {
                method: 'GET',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                // Clear the WhatsApp session tables (like clear all sessions but for this device)
                return fetch('/api/devices/clear-all-sessions', {
                    method: 'POST',
                    credentials: 'include',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ deviceId: deviceId }) // Pass device ID to clear only this device
                });
            })
            .then(response => response.json())
            .then(data => {
                clearTimeout(loadingTimeout); // Clear the timeout
                
                // Update device status to offline (not disconnected)
                const device = devices.find(d => d.id === deviceId);
                if (device) {
                    device.status = 'offline'; // Important: use 'offline' not 'disconnected'
                    device.phone = '';
                    device.jid = '';
                    device.lastSeen = new Date().toISOString();
                }
                renderDevices();
                
                hideLoading(); // Hide loading before showing success
                
                Swal.fire({
                    icon: 'success',
                    title: 'Device Logged Out!',
                    text: 'The device has been disconnected and session cleared. You can now scan QR code to connect again.',
                    timer: 3000,
                    showConfirmButton: false
                });
                
                // Refresh devices after a moment
                setTimeout(loadDevices, 2000);
            })
            .catch(error => {
                clearTimeout(loadingTimeout); // Clear the timeout
                console.error('Error logging out device:', error);
                hideLoading(); // Hide loading first
                
                // If clear sessions endpoint fails, still show success if logout worked
                if (error.message && error.message.includes('clear-all-sessions')) {
                    // Update status anyway
                    const device = devices.find(d => d.id === deviceId);
                    if (device) {
                        device.status = 'offline';
                        device.phone = '';
                        device.jid = '';
                    }
                    renderDevices();
                    
                    Swal.fire({
                        icon: 'warning',
                        title: 'Partial Success',
                        text: 'Device logged out but session clearing failed. Try using the device delete option for complete cleanup.'
                    });
                } else {
                    Swal.fire({
                        icon: 'error',
                        title: 'Logout Failed',
                        text: error.message || 'Failed to logout device. Please try again.'
                    });
                }
            });
        }
    });
}