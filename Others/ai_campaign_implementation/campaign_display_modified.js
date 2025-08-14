// Modified displayCampaigns function to show AI campaigns with special indicator
// Replace the existing displayCampaigns function in dashboard JavaScript

function displayCampaigns(campaigns) {
    const campaignList = document.getElementById('campaignList');
    campaignList.innerHTML = '';
    
    if (!campaigns || campaigns.length === 0) {
        campaignList.innerHTML = '<p class="text-center text-muted">No campaigns found</p>';
        return;
    }
    
    campaigns.forEach(campaign => {
        const isAICampaign = campaign.ai === 'ai';
        const card = document.createElement('div');
        card.className = 'col-md-6 col-lg-4 mb-3';
        
        card.innerHTML = `
            <div class="card campaign-card ${isAICampaign ? 'border-primary' : ''}">
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-start mb-2">
                        <h6 class="card-title mb-0">
                            ${isAICampaign ? '<i class="bi bi-robot text-primary"></i> ' : ''}
                            ${campaign.title}
                        </h6>
                        <span class="badge bg-${getStatusColor(campaign.status)}">${campaign.status}</span>
                    </div>
                    ${isAICampaign ? `<p class="text-muted small mb-1"><strong>AI Campaign</strong> - Limit: ${campaign.limit} per device</p>` : ''}
                    <p class="text-muted small mb-1">Date: ${campaign.campaign_date}</p>
                    <p class="text-muted small mb-1">Time: ${campaign.time_schedule || 'Not scheduled'}</p>
                    <p class="text-muted small mb-1">Niche: ${campaign.niche || 'All'}</p>
                    <p class="text-muted small mb-2">Target: ${campaign.target_status || 'All'}</p>
                    <div class="d-flex gap-2">
                        <button class="btn btn-sm btn-primary" onclick="viewCampaign(${campaign.id})">
                            <i class="bi bi-eye"></i> View
                        </button>
                        ${campaign.status === 'pending' && isAICampaign ? 
                            `<button class="btn btn-sm btn-success" onclick="triggerAICampaign(${campaign.id})">
                                <i class="bi bi-play-circle"></i> Trigger
                            </button>` : ''
                        }
                        ${campaign.status === 'pending' ? 
                            `<button class="btn btn-sm btn-warning" onclick="editCampaign(${campaign.id})">
                                <i class="bi bi-pencil"></i> Edit
                            </button>
                            <button class="btn btn-sm btn-danger" onclick="deleteCampaign(${campaign.id})">
                                <i class="bi bi-trash"></i> Delete
                            </button>` : ''
                        }
                    </div>
                </div>
            </div>
        `;
        
        campaignList.appendChild(card);
    });
}

function getStatusColor(status) {
    const colors = {
        'pending': 'warning',
        'triggered': 'info',
        'completed': 'success',
        'completed_with_errors': 'warning',
        'failed': 'danger'
    };
    return colors[status] || 'secondary';
}