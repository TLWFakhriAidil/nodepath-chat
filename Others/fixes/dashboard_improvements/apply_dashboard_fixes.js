const fs = require('fs');
const path = require('path');

console.log('Applying dashboard fixes...');

const dashboardPath = path.join(__dirname, '../../src/views/dashboard.html');
let dashboardContent = fs.readFileSync(dashboardPath, 'utf8');

// 1. Fix worker auto-refresh (remove checked attribute)
dashboardContent = dashboardContent.replace(
    '<input class="form-check-input" type="checkbox" id="workerAutoRefresh" checked>',
    '<input class="form-check-input" type="checkbox" id="workerAutoRefresh">'
);

// 2. Add worker control buttons
dashboardContent = dashboardContent.replace(
    `<button class="btn btn-primary btn-sm" onclick="loadWorkerStatus()">
                            <i class="bi bi-arrow-clockwise"></i> Refresh
                        </button>`,
    `<button class="btn btn-primary btn-sm" onclick="loadWorkerStatus()">
                            <i class="bi bi-arrow-clockwise"></i> Refresh
                        </button>
                        <button class="btn btn-success btn-sm ms-2" onclick="resumeFailedWorkers()">
                            <i class="bi bi-play-fill"></i> Resume Failed
                        </button>
                        <button class="btn btn-danger btn-sm ms-2" onclick="stopAllWorkers()">
                            <i class="bi bi-stop-fill"></i> Stop All
                        </button>`
);

// 3. Add worker control functions
const workerControlFunctions = `
// Worker Control Functions
async function resumeFailedWorkers() {
    try {
        const response = await fetch('/api/workers/resume-failed', { method: 'POST' });
        const data = await response.json();
        if (data.code === 'SUCCESS') {
            showToast('Failed workers resumed successfully', 'success');
            loadWorkerStatus();
        } else {
            showToast(data.message || 'Failed to resume workers', 'error');
        }
    } catch (error) {
        console.error('Error resuming workers:', error);
        showToast('Failed to resume workers', 'error');
    }
}

async function stopAllWorkers() {
    const result = await Swal.fire({
        title: 'Stop All Workers?',
        text: 'This will stop all running workers',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#d33',
        confirmButtonText: 'Yes, stop all!'
    });
    
    if (result.isConfirmed) {
        try {
            const response = await fetch('/api/workers/stop-all', { method: 'POST' });
            const data = await response.json();
            if (data.code === 'SUCCESS') {
                showToast('All workers stopped', 'success');
                loadWorkerStatus();
            } else {
                showToast(data.message || 'Failed to stop workers', 'error');
            }
        } catch (error) {
            console.error('Error stopping workers:', error);
            showToast('Failed to stop workers', 'error');
        }
    }
}
`;

// Insert worker control functions after loadWorkerStatus
dashboardContent = dashboardContent.replace(
    'function displayWorkerStatus(status) {',
    workerControlFunctions + '\n\nfunction displayWorkerStatus(status) {'
);

// 4. Fix toast function if missing
if (!dashboardContent.includes('function showToast')) {
    const toastFunction = `
// Toast notification function
function showToast(message, type = 'info') {
    const toastHtml = \`
        <div class="toast align-items-center text-white bg-\${type === 'error' ? 'danger' : type === 'success' ? 'success' : 'info'} border-0" role="alert">
            <div class="d-flex">
                <div class="toast-body">\${message}</div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        </div>
    \`;
    
    const toastContainer = document.getElementById('toastContainer') || (() => {
        const container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
        document.body.appendChild(container);
        return container;
    })();
    
    const toastElement = document.createElement('div');
    toastElement.innerHTML = toastHtml;
    const toast = new bootstrap.Toast(toastElement.firstElementChild);
    toastContainer.appendChild(toastElement.firstElementChild);
    toast.show();
}
`;
    dashboardContent = dashboardContent.replace('</script>', toastFunction + '\n</script>');
}

fs.writeFileSync(dashboardPath, dashboardContent, 'utf8');
console.log('Dashboard fixes applied successfully!');
