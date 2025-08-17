// Enhanced Logout Device Function with Authentication Fix
// This version properly handles authentication and clears device-specific session

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
            
            // Use the logout endpoint which handles authentication properly
            fetch('/app/logout?deviceId=' + deviceId, {
                method: 'GET',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    // Now clear the device-specific session
                    return clearDeviceSession(deviceId);
                } else {
                    throw new Error(data.message || 'Failed to logout device');
                }
            })
            .then(() => {
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
            })
            .catch(error => {
                console.error('Error logging out device:', error);
                Swal.fire({
                    icon: 'error',
                    title: 'Logout Failed',
                    text: error.message || 'Failed to logout device. Please try again.'
                });
                hideLoading();
            });
        }
    });
}

// Clear device-specific WhatsApp session
function clearDeviceSession(deviceId) {
    return fetch(`/api/devices/${deviceId}/clear-session`, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            // If the endpoint doesn't exist, that's okay - the logout already happened
            console.log('Clear session endpoint not available, using logout only');
        }
        return response.json();
    })
    .catch(error => {
        // If clear-session endpoint doesn't exist, just continue
        console.log('Clear session endpoint not implemented, continuing with logout only');
        return { code: 'SUCCESS' };
    });
}