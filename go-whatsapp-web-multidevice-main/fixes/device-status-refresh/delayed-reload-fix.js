// Fix for device status not updating after connection
// The issue: loadDevices() is called too quickly after DEVICE_CONNECTED message

// Find this section in dashboard.html around line 2307:
case 'DEVICE_CONNECTED':
    // Device fully connected and logged in
    console.log('Device fully connected:', data.message);
    // Close QR modal if open
    const qrModal = bootstrap.Modal.getInstance(document.getElementById('qrModal'));
    if (qrModal) {
        qrModal.hide();
    }
    // Show success message
    alert('WhatsApp connected successfully!');
    // Reload devices to show updated status
    loadDevices();
    break;

// Replace with:
case 'DEVICE_CONNECTED':
    // Device fully connected and logged in
    console.log('Device fully connected:', data.message);
    // Close QR modal if open
    const qrModal = bootstrap.Modal.getInstance(document.getElementById('qrModal'));
    if (qrModal) {
        qrModal.hide();
    }
    // Show success message
    alert('WhatsApp connected successfully!');
    
    // Add delay before reloading devices to ensure database is updated
    console.log('Waiting for database update before refreshing device list...');
    setTimeout(() => {
        console.log('Reloading device list...');
        loadDevices();
        
        // If still showing disconnected, try again after another delay
        setTimeout(() => {
            const deviceCards = document.querySelectorAll('.device-card');
            const stillDisconnected = Array.from(deviceCards).some(card => {
                const statusText = card.querySelector('.text-muted.small');
                return statusText && statusText.textContent === 'Disconnected';
            });
            
            if (stillDisconnected) {
                console.log('Device still showing disconnected, reloading again...');
                loadDevices();
            }
        }, 3000); // Check again after 3 seconds
    }, 2000); // Initial delay of 2 seconds
    break;
