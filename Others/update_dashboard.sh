#!/bin/bash

# This script updates the dashboard.html file to:
# 1. Remove the "Reset WhatsApp Session" menu item
# 2. Update the logout function to also reset the session

# Step 1: Remove Reset WhatsApp Session menu item
sed -i '/<li><a class="dropdown-item text-warning" href="#" onclick="resetDevice/,/<\/a><\/li>/d' src/views/dashboard.html

# Step 2: Update the logoutDevice function to include session reset
cat > temp_logout_function.txt << 'EOF'
        // Logout Device (also resets WhatsApp session)
        function logoutDevice(deviceId) {
            Swal.fire({
                title: 'Logout Device?',
                html: `
                    <div class="text-start">
                        <p>This will:</p>
                        <ul class="mb-0">
                            <li>Logout from WhatsApp</li>
                            <li>Clear the session data</li>
                            <li>Require QR code scan to reconnect</li>
                        </ul>
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
                    
                    // First logout from WhatsApp
                    fetch('/app/logout?deviceId=' + deviceId, {
                        method: 'GET',
                        credentials: 'include'
                    })
                    .then(response => response.json())
                    .then(data => {
                        if (data.code === 'SUCCESS') {
                            // Then reset the session
                            return fetch(`/api/devices/${deviceId}/reset`, {
                                method: 'POST',
                                credentials: 'include'
                            });
                        } else {
                            throw new Error(data.message || 'Logout failed');
                        }
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
                                renderDevices();
                            }
                            
                            Swal.fire({
                                icon: 'success',
                                title: 'Device Logged Out!',
                                text: 'The device has been logged out and session cleared.',
                                timer: 3000,
                                showConfirmButton: false
                            });
                            
                            // Refresh devices after a moment
                            setTimeout(loadDevices, 2000);
                        } else {
                            throw new Error(data.message || 'Session reset failed');
                        }
                        hideLoading();
                    })
                    .catch(error => {
                        hideLoading();
                        Swal.fire({
                            icon: 'error',
                            title: 'Logout Failed',
                            text: error.message || 'Failed to logout device'
                        });
                    });
                }
            });
        }
EOF

echo "Dashboard update complete!"
