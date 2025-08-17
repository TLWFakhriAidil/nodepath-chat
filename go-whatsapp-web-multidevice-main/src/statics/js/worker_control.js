// Worker Control Functions for Dashboard

// Resume Failed Workers
async function resumeFailedWorkers() {
    if (!confirm('Resume all failed/stopped workers?')) return;
    
    try {
        const response = await fetch('/api/workers/resume-failed', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Auth-Token': localStorage.getItem('auth_token') || ''
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast(`Resumed ${data.resumed} workers, ${data.failed} failed`, 'success');
            // Refresh worker status
            setTimeout(() => loadWorkerStatus(), 1000);
        } else {
            showToast(data.error || 'Failed to resume workers', 'error');
        }
    } catch (error) {
        console.error('Error resuming workers:', error);
        showToast('Error resuming workers', 'error');
    }
}

// Stop All Workers
async function stopAllWorkers() {
    if (!confirm('Are you sure you want to stop ALL workers? This will halt all message sending.')) return;
    
    try {
        const response = await fetch('/api/workers/stop-all', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Auth-Token': localStorage.getItem('auth_token') || ''
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast(`Stopped ${data.stopped} workers`, 'success');
            // Refresh worker status
            setTimeout(() => loadWorkerStatus(), 1000);
        } else {
            showToast(data.error || 'Failed to stop workers', 'error');
        }
    } catch (error) {
        console.error('Error stopping workers:', error);
        showToast('Error stopping workers', 'error');
    }
}

// Restart specific worker
async function restartWorker(deviceId) {
    try {
        const response = await fetch(`/api/workers/${deviceId}/restart`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Auth-Token': localStorage.getItem('auth_token') || ''
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast('Worker restarted successfully', 'success');
            // Refresh worker status
            setTimeout(() => loadWorkerStatus(), 1000);
        } else {
            showToast(data.error || 'Failed to restart worker', 'error');
        }
    } catch (error) {
        console.error('Error restarting worker:', error);
        showToast('Error restarting worker', 'error');
    }
}

// Start worker for device
async function startWorker(deviceId) {
    try {
        const response = await fetch(`/api/workers/${deviceId}/start`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Auth-Token': localStorage.getItem('auth_token') || ''
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast('Worker started successfully', 'success');
            // Refresh worker status
            setTimeout(() => loadWorkerStatus(), 1000);
        } else {
            showToast(data.error || 'Failed to start worker', 'error');
        }
    } catch (error) {
        console.error('Error starting worker:', error);
        showToast('Error starting worker', 'error');
    }
}

// Reconnect device
async function reconnectDevice(deviceId) {
    showToast('Attempting to reconnect device...', 'info');
    
    try {
        const response = await fetch(`/api/devices/${deviceId}/reconnect`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Auth-Token': localStorage.getItem('auth_token') || ''
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast('Device reconnected successfully', 'success');
            // Refresh worker status
            setTimeout(() => loadWorkerStatus(), 2000);
        } else {
            showToast(data.error || 'Failed to reconnect device', 'error');
            if (data.details && data.details.includes('needs QR scan')) {
                showToast('Device needs QR scan. Please go to Devices tab.', 'warning');
            }
        }
    } catch (error) {
        console.error('Error reconnecting device:', error);
        showToast('Error reconnecting device', 'error');
    }
}

// Show toast notification
function showToast(message, type = 'info') {
    // Create toast container if it doesn't exist
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        toastContainer.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 9999;';
        document.body.appendChild(toastContainer);
    }
    
    // Create toast element
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} alert-dismissible fade show`;
    toast.style.cssText = 'min-width: 250px; margin-bottom: 10px;';
    toast.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    toastContainer.appendChild(toast);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 150);
    }, 5000);
}