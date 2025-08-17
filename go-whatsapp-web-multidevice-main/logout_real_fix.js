// REAL FIX: Make logout clear WhatsApp session for specific device
// This is the proper solution - logout should clear the session data just like delete does

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
                location.reload(); // Reload page if timeout
            }, 15000); // 15 seconds timeout
            
            // The real fix: Use the delete endpoint which properly clears everything
            // but without actually deleting the device record
            fetch(`/api/devices/${deviceId}`, {
                method: 'DELETE',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                clearTimeout(loadingTimeout);
                hideLoading();
                
                if (data.code === 'SUCCESS') {
                    Swal.fire({
                        icon: 'success',
                        title: 'Device Logged Out!',
                        text: 'The device has been disconnected and session cleared. You can now scan QR code to connect again.',
                        timer: 3000,
                        showConfirmButton: false
                    });
                    
                    // Reload page to refresh device list
                    setTimeout(() => location.reload(), 1000);
                } else {
                    throw new Error(data.message || 'Failed to logout device');
                }
            })
            .catch(error => {
                clearTimeout(loadingTimeout);
                console.error('Error logging out device:', error);
                hideLoading();
                Swal.fire({
                    icon: 'error',
                    title: 'Operation Failed',
                    text: 'Failed to logout device. Please try again or refresh the page.'
                });
            });
        }
    });
}