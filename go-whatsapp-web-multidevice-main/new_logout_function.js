// Enhanced Logout Device Function for WhatsApp Multi-Device System
// This replaces the existing logoutDevice function in dashboard.html
// It now combines logout + session removal in one action

// Logout Device (now includes session removal)
function logoutDevice(deviceId) {
    Swal.fire({
        title: 'Logout Device?',
        html: `
            <div class="text-start">
                <p>This will disconnect the device from WhatsApp and remove the session.</p>
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
            
            // First logout the device
            fetch('/app/logout?deviceId=' + deviceId, {
                method: 'GET',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                // Then reset the session to ensure complete removal
                return fetch(`/api/devices/${deviceId}/reset`, {
                    method: 'POST',
                    credentials: 'include'
                });
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    // Update device status
                    const device = devices.find(d => d.id === deviceId);
                    if (device) {
                        device.status = 'offline';
                        device.phone = '';
                        device.jid = '';
                        device.lastSeen = new Date().toISOString();
                    }
                    renderDevices();
                    
                    Swal.fire({
                        icon: 'success',
                        title: 'Device Logged Out!',
                        text: 'The device has been disconnected and session removed. You can now scan QR code to connect again.',
                        timer: 3000,
                        showConfirmButton: false
                    });
                    
                    // Refresh devices after a moment
                    setTimeout(loadDevices, 2000);
                } else {
                    Swal.fire({
                        icon: 'error',
                        title: 'Logout Failed',
                        text: data.message || 'Failed to logout device'
                    });
                }
                hideLoading();
            })
            .catch(error => {
                console.error('Error logging out device:', error);
                Swal.fire({
                    icon: 'error',
                    title: 'Connection Error',
                    text: 'Failed to logout device. Please try again.'
                });
                hideLoading();
            });
        }
    });
}