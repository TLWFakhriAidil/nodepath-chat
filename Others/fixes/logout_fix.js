// Fix logout device function to actually logout from WhatsApp
function logoutDevice(deviceId) {
    if (confirm('Are you sure you want to logout this device from WhatsApp?')) {
        showLoading();
        
        // The actual logout endpoint is /app/logout
        fetch(`/app/logout`, {
            method: 'GET',
            credentials: 'include'
        })
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS') {
                // Update device status locally
                const device = devices.find(d => d.id === deviceId);
                if (device) {
                    device.status = 'offline';
                    device.phone = '';
                    device.jid = '';
                    device.lastSeen = new Date().toISOString();
                }
                
                // Update device status in database
                updateDeviceStatus(deviceId, 'offline');
                
                // Re-render devices
                renderDevices();
                showAlert('success', 'Device logged out successfully');
            } else {
                showAlert('danger', data.message || 'Failed to logout device');
            }
            hideLoading();
        })
        .catch(error => {
            console.error('Error logging out device:', error);
            showAlert('danger', 'Error logging out device');
            hideLoading();
        });
    }
}

// Helper function to update device status in database
function updateDeviceStatus(deviceId, status) {
    const userRepo = repository.GetUserRepository();
    if (userRepo && userRepo.UpdateDeviceStatus) {
        userRepo.UpdateDeviceStatus(deviceId, status, '', '');
    }
}
