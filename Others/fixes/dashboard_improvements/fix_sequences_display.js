const fs = require('fs');
const path = require('path');

console.log('Fixing sequences display...');

const dashboardPath = path.join(__dirname, '../../src/views/dashboard.html');
let dashboardContent = fs.readFileSync(dashboardPath, 'utf8');

// Update displaySequences function to properly show data
const updatedDisplaySequences = `function displaySequences(sequences) {
    const container = document.getElementById('sequencesContainer');
    
    if (!sequences || sequences.length === 0) {
        // Keep existing empty state
        return;
    }
    
    let html = '<div class="row g-3">';
    sequences.forEach(seq => {
        const statusColor = seq.status === 'active' ? 'success' : 
                           seq.status === 'paused' ? 'warning' : 'secondary';
        
        html += \`
            <div class="col-md-4">
                <div class="card sequence-card h-100">
                    <div class="card-body">
                        <h5 class="card-title">\${seq.name}</h5>
                        <p class="text-muted mb-2">\${seq.description || 'No description'}</p>
                        <div class="mb-3">
                            <small class="text-muted">Niche: \${seq.niche || 'Not set'}</small>
                        </div>
                        <div class="d-flex justify-content-between align-items-center mb-3">
                            <span class="badge bg-\${statusColor}">\${seq.status}</span>
                            <small>\${seq.contacts_count || 0} contacts</small>
                        </div>
                        <div class="progress mb-3" style="height: 5px;">
                            <div class="progress-bar bg-\${statusColor}" style="width: \${(seq.completed_count / seq.contacts_count * 100) || 0}%"></div>
                        </div>
                        <div class="d-flex gap-2">
                            <button class="btn btn-sm btn-outline-primary" onclick="window.location.href='/sequences/\${seq.id}'">
                                <i class="bi bi-eye"></i> View
                            </button>
                            <button class="btn btn-sm btn-outline-\${seq.status === 'active' ? 'warning' : 'success'}" 
                                    onclick="toggleSequence('\${seq.id}', '\${seq.status}')">
                                <i class="bi bi-\${seq.status === 'active' ? 'pause' : 'play'}"></i> 
                                \${seq.status === 'active' ? 'Pause' : 'Start'}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        \`;
    });
    html += '</div>';
    container.innerHTML = html;
}`;

// Replace the displaySequences function
const regex = /function displaySequences\(sequences\) {[\s\S]*?^}/m;
dashboardContent = dashboardContent.replace(regex, updatedDisplaySequences);

// Add toggleSequence function if it doesn't exist
if (!dashboardContent.includes('function toggleSequence')) {
    const toggleSequenceFunction = `
async function toggleSequence(id, currentStatus) {
    const action = currentStatus === 'active' ? 'pause' : 'start';
    const endpoint = \`/api/sequences/\${id}/\${action}\`;
    
    try {
        const response = await fetch(endpoint, { method: 'POST' });
        const data = await response.json();
        
        if (data.code === 'SUCCESS') {
            showToast(\`Sequence \${action}ed successfully!\`, 'success');
            loadSequences();
        } else {
            showToast(data.message || \`Failed to \${action} sequence\`, 'error');
        }
    } catch (error) {
        console.error(\`Error \${action}ing sequence:\`, error);
        showToast(\`Failed to \${action} sequence\`, 'error');
    }
}
`;
    dashboardContent = dashboardContent.replace(
        'async function loadSequences() {',
        toggleSequenceFunction + '\n\nasync function loadSequences() {'
    );
}

fs.writeFileSync(dashboardPath, dashboardContent, 'utf8');
console.log('Sequences display fixed successfully!');
